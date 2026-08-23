package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"podcast-platform/api/dto"
	"podcast-platform/internal/handler"
	"podcast-platform/internal/middleware"
	"podcast-platform/internal/repository"
	"podcast-platform/internal/service"
)

type Repositories struct {
	UserRepo       *repository.UserRepository
	ChannelRepo    *repository.ChannelRepository
	EpisodeRepo    *repository.EpisodeRepository
	ChapterRepo    *repository.ChapterRepository
	CommentRepo    *repository.CommentRepository
	PlaybackRepo   *repository.PlaybackRepository
	SubscriberRepo *repository.SubscriberRepository
	CategoryRepo   *repository.CategoryRepository
}

type Services struct {
	AuthSvc       *service.AuthService
	ChannelSvc    *service.ChannelService
	EpisodeSvc    *service.EpisodeService
	AudioSvc      *service.AudioService
	RSSSvc        *service.RSSService
	StatsSvc      *service.StatsService
	CommentSvc    *service.CommentService
	SubscriberSvc *service.SubscriberService
}

func SetupRouter(r *gin.Engine, repos *Repositories, svc *Services, storageServePath, webDir string) []*middleware.RateLimiter {
	var limiters []*middleware.RateLimiter

	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	globalLimiter, globalMW := middleware.GlobalRateLimit()
	limiters = append(limiters, globalLimiter)
	r.Use(globalMW)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, dto.Err(404, "not found"))
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.OK(gin.H{"status": "ok"}))
	})

	r.Static("/storage", storageServePath)
	r.Static("/static", webDir)
	r.StaticFile("/", webDir+"/index.html")
	r.StaticFile("/login", webDir+"/login.html")
	r.StaticFile("/register", webDir+"/register.html")
	r.StaticFile("/dashboard", webDir+"/dashboard.html")
	r.StaticFile("/channels", webDir+"/channels.html")
	r.StaticFile("/episodes", webDir+"/episodes.html")
	r.StaticFile("/stats", webDir+"/stats.html")
	r.StaticFile("/rss-view", webDir+"/rss-view.html")
	r.StaticFile("/player", webDir+"/player.html")

	authH := handler.NewAuthHandler(svc.AuthSvc)
	channelH := handler.NewChannelHandler(svc.ChannelSvc, svc.SubscriberSvc)
	episodeH := handler.NewEpisodeHandler(svc.EpisodeSvc, svc.AudioSvc, svc.CommentSvc, svc.StatsSvc)
	statsH := handler.NewStatsHandler(svc.StatsSvc, svc.CommentSvc)
	rssH := handler.NewRSSHandler(svc.RSSSvc, repos.ChannelRepo)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			registerLimiter, registerMW := middleware.PerUserRateLimit(10)
			limiters = append(limiters, registerLimiter)
			auth.POST("/register", registerMW, authH.Register)

			loginLimiter, loginMW := middleware.PerUserRateLimit(15)
			limiters = append(limiters, loginLimiter)
			auth.POST("/login", loginMW, authH.Login)

			auth.POST("/logout", authH.Logout)
			auth.GET("/me", middleware.AuthMiddleware(svc.AuthSvc), authH.Me)
			auth.PUT("/password", middleware.AuthMiddleware(svc.AuthSvc), authH.ChangePassword)
			auth.PUT("/profile", middleware.AuthMiddleware(svc.AuthSvc), authH.UpdateProfile)
		}

		channels := api.Group("/channels")
		{
			channels.GET("", middleware.OptionalAuthMiddleware(svc.AuthSvc), channelH.ListAll)
			channels.GET("/:id", channelH.GetByID)
			channels.GET("/:id/episodes", episodeH.ListByChannel)
			channels.GET("/:id/episodes/published", episodeH.ListPublished)
			channels.GET("/:id/rss", rssH.Feed)
			channels.GET("/:id/rss/validate", rssH.Validate)
			channels.GET("/:id/rss/links", rssH.SubscribeLinks)
			channels.POST("/:id/subscribe", channelH.Subscribe)
			channels.GET("/:id/comments", statsH.ListChannelComments)

			authCh := channels.Group("", middleware.AuthMiddleware(svc.AuthSvc))
			{
				authCh.POST("", channelH.Create)
				authCh.GET("/mine/list", channelH.ListMyChannels)
				authCh.PUT("/:id", channelH.Update)
				authCh.DELETE("/:id", channelH.Delete)
				authCh.POST("/:id/approve", middleware.AdminRequired(), channelH.Approve)
				authCh.POST("/:id/reject", middleware.AdminRequired(), channelH.Reject)
				authCh.GET("/:id/stats", statsH.ChannelDashboard)
				authCh.GET("/:id/stats/export", statsH.ExportCSV)
				authCh.GET("/:id/subscribers", channelH.ListSubscribers)
				authCh.GET("/:id/subscribers/export", channelH.ExportSubscribers)
				authCh.POST("/:id/rss/refresh", rssH.RefreshCache)
				authCh.POST("/:id/episodes", episodeH.Create)
			}
		}

		episodes := api.Group("/episodes")
		{
			episodes.GET("/search", episodeH.Search)
			episodes.GET("/:id", episodeH.GetByID)
			episodes.GET("/:id/comments", episodeH.ListComments)
			episodes.GET("/:id/danmaku", episodeH.GetDanmaku)
			episodes.GET("/:id/completion", statsH.EpisodeCompletion)

			authEp := episodes.Group("", middleware.AuthMiddleware(svc.AuthSvc))
			{
				authEp.PUT("/:id", episodeH.Update)
				authEp.DELETE("/:id", episodeH.Delete)
				authEp.POST("/:id/like", episodeH.Like)
				authEp.POST("/:id/comments", episodeH.AddComment)
				authEp.POST("/playback/start", episodeH.StartPlayback)
			}
		}

		upload := api.Group("/upload", middleware.AuthMiddleware(svc.AuthSvc))
		{
			upload.POST("/audio", episodeH.UploadAudio)
			upload.POST("/audio/chunk", episodeH.UploadChunk)
		}

		comments := api.Group("/comments", middleware.AuthMiddleware(svc.AuthSvc))
		{
			comments.GET("/pending", statsH.ListPendingComments)
			comments.POST("/:comment_id/approve", statsH.ApproveComment)
			comments.POST("/:comment_id/reject", statsH.RejectComment)
		}
	}

	rssGroup := r.Group("/rss")
	{
		rssGroup.GET("/:slug", rssH.FeedBySlug)
	}

	unsub := r.Group("/unsubscribe")
	{
		unsub.Use(func(c *gin.Context) {
			c.Set("subscriber_svc", svc.SubscriberSvc)
			c.Next()
		})
		unsub.GET("/:token", rssH.Unsubscribe)
	}

	return limiters
}
