package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"podcast-platform/api"
	"podcast-platform/config"
	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
	"podcast-platform/internal/worker"
	"podcast-platform/pkg/ffmpeg"
	"podcast-platform/pkg/logger"
	"podcast-platform/pkg/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Server.Mode == "debug")
	defer log.Sync()

	_ = utils.EnsureDir(cfg.Storage.LocalPath)
	_ = utils.EnsureDir(filepath.Join(cfg.Storage.LocalPath, "audio"))
	_ = utils.EnsureDir(filepath.Join(cfg.Storage.LocalPath, "covers"))
	_ = utils.EnsureDir(cfg.Server.LogDir)

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}
	log.Info("database connected")

	if cfg.Database.AutoMigrate {
		models := []interface{}{
			&domain.User{}, &domain.Category{}, &domain.Channel{},
			&domain.Episode{}, &domain.Chapter{},
			&domain.Comment{}, &domain.Playback{}, &domain.Subscriber{},
		}
		for _, m := range models {
			if err := db.AutoMigrate(m); err != nil {
				log.Fatalf("migrate failed: %v", err)
			}
		}
		seedCategories(db)
		log.Info("migration completed")
	}

	repos := &api.Repositories{
		UserRepo:       repository.NewUserRepository(db),
		ChannelRepo:    repository.NewChannelRepository(db),
		EpisodeRepo:    repository.NewEpisodeRepository(db),
		ChapterRepo:    repository.NewChapterRepository(db),
		CommentRepo:    repository.NewCommentRepository(db),
		PlaybackRepo:   repository.NewPlaybackRepository(db),
		SubscriberRepo: repository.NewSubscriberRepository(db),
		CategoryRepo:   repository.NewCategoryRepository(db),
	}

	authSvc := service.NewAuthService(repos.UserRepo)
	channelSvc := service.NewChannelService(repos.ChannelRepo)

	processor := ffmpeg.NewProcessor()
	storageSvc := service.NewStorageService(&cfg.Storage, cfg.Storage.BaseURL)
	audioSvc := service.NewAudioService(processor, storageSvc)

	episodeSvc := service.NewEpisodeService(repos.EpisodeRepo, repos.ChapterRepo, repos.ChannelRepo, audioSvc)
	rssSvc := service.NewRSSService(repos.ChannelRepo, repos.EpisodeRepo, cfg.Server.BaseURL)

	statsSvc := service.NewStatsService(repos.PlaybackRepo, repos.EpisodeRepo, repos.ChannelRepo, repos.SubscriberRepo)
	commentSvc := service.NewCommentService(repos.CommentRepo, repos.EpisodeRepo, repos.ChannelRepo)
	subscriberSvc := service.NewSubscriberService(repos.SubscriberRepo, repos.ChannelRepo)

	svc := &api.Services{
		AuthSvc:       authSvc,
		ChannelSvc:    channelSvc,
		EpisodeSvc:    episodeSvc,
		AudioSvc:      audioSvc,
		RSSSvc:        rssSvc,
		StatsSvc:      statsSvc,
		CommentSvc:    commentSvc,
		SubscriberSvc: subscriberSvc,
	}

	ensureAdmin(repos.UserRepo, log)

	webDir := cfg.Server.WebDir
	storagePath := cfg.Storage.LocalPath

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	rateLimiters := api.SetupRouter(r, repos, svc, storagePath, webDir)

	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	workerInst, err := worker.NewWorker(log)
	if err != nil {
		log.Fatalf("create worker failed: %v", err)
	}
	publishHandler := worker.NewPublishHandler(repos.EpisodeRepo, rssSvc, log)
	aggregateHandler := worker.NewAggregateHandler(repos.PlaybackRepo, repos.ChannelRepo, statsSvc, log)
	refreshHandler := worker.NewFeedRefreshHandler(repos.ChannelRepo, rssSvc, log)
	workerInst.AddHandler(publishHandler)
	workerInst.AddHandler(aggregateHandler)
	workerInst.AddHandler(refreshHandler)
	_ = workerInst.RegisterCron("*/30 * * * *", publishHandler)
	_ = workerInst.RegisterCron("15 3 * * *", aggregateHandler)
	_ = workerInst.RegisterCron("0 */2 * * *", refreshHandler)
	workerInst.Start()

	go func() {
		log.Infof("server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = workerInst.Stop(ctx)
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("server shutdown error: %v", err)
	}
	for _, rl := range rateLimiters {
		rl.Close()
	}
	log.Info("server exited")
}

func seedCategories(db *gorm.DB) {
	categories := []domain.Category{
		{Name: "科技", Slug: "tech"},
		{Name: "商业", Slug: "business"},
		{Name: "教育", Slug: "education"},
		{Name: "生活", Slug: "lifestyle"},
		{Name: "文化", Slug: "culture"},
		{Name: "娱乐", Slug: "entertainment"},
		{Name: "健康", Slug: "health"},
		{Name: "新闻", Slug: "news"},
	}
	for i := range categories {
		var count int64
		db.Model(&domain.Category{}).Where("slug = ?", categories[i].Slug).Count(&count)
		if count == 0 {
			db.Create(&categories[i])
		}
	}
}

func ensureAdmin(userRepo *repository.UserRepository, log *logger.Logger) {
	_, err := userRepo.GetByUsername("admin")
	if err == nil {
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	u := &domain.User{
		Username:     "admin",
		Nickname:     "管理员",
		Email:        "admin@example.com",
		PasswordHash: string(hash),
		Role:         domain.RoleAdmin,
		Status:       1,
	}
	if err := userRepo.Create(u); err != nil {
		log.Errorf("create admin failed: %v", err)
		return
	}
	log.Info("admin created: admin / admin123")
}
