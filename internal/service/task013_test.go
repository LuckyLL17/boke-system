package service

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"podcast-platform/internal/domain"
	"podcast-platform/internal/repository"
)

// RecordPlaybackStart -> StatsService -> IncrementPlayCount/IncrementTotalPlays
func TestConcurrentPlaybackCountersMatchDetails(t *testing.T) {
	runCounterRace(t, "episodes", func(db *gorm.DB) error {
		return db.AutoMigrate(&domain.Episode{})
	}, func(db *gorm.DB) error {
		return db.Create(&domain.Episode{ID: 1, ChannelID: 2, Title: "节目", AudioFileURL: "/audio/a.mp3"}).Error
	}, func(db *gorm.DB) error {
		return repository.NewEpisodeRepository(db).IncrementPlayCount(1, 1)
	}, func(db *gorm.DB) int64 {
		var ep domain.Episode
		_ = db.First(&ep, 1).Error
		return ep.PlayCount
	})

	runCounterRace(t, "channels", func(db *gorm.DB) error {
		return db.AutoMigrate(&domain.Channel{})
	}, func(db *gorm.DB) error {
		return db.Create(&domain.Channel{ID: 1, OwnerID: 2, Title: "频道", Slug: "race-channel"}).Error
	}, func(db *gorm.DB) error {
		return repository.NewChannelRepository(db).IncrementTotalPlays(1, 1)
	}, func(db *gorm.DB) int64 {
		var ch domain.Channel
		_ = db.First(&ch, 1).Error
		return ch.TotalPlays
	})
}

func runCounterRace(t *testing.T, table string, migrate func(*gorm.DB) error, seed func(*gorm.DB) error, increment func(*gorm.DB) error, read func(*gorm.DB) int64) {
	t.Helper()
	name := "file:" + t.TempDir() + "/task013-" + table + "-" + time.Now().Format("150405.000000000") + ".db?_busy_timeout=5000&_journal_mode=WAL"
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := seed(db); err != nil {
		t.Fatal(err)
	}

	var reads atomic.Int32
	readStarted := make(chan struct{})
	release := make(chan struct{})
	if err := db.Callback().Query().After("gorm:query").Register("task013_barrier_"+table, func(tx *gorm.DB) {
		if tx.Statement.Table != table {
			return
		}
		if reads.Add(1) == 2 {
			close(readStarted)
		}
		select {
		case <-release:
		case <-time.After(500 * time.Millisecond):
		}
	}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- increment(db)
		}()
	}
	close(start)
	select {
	case <-readStarted:
	case <-time.After(200 * time.Millisecond):
	}
	close(release)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := read(db); got != 2 {
		t.Fatalf("concurrent counter lost an update: table=%s value=%d", table, got)
	}
}
