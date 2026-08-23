package worker

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"

	"podcast-platform/pkg/logger"
)

type Worker struct {
	scheduler  gocron.Scheduler
	handlers   []Handler
	logger     *logger.Logger
	processing bool
}

type Handler interface {
	Name() string
	Run(ctx context.Context) error
}

func NewWorker(log *logger.Logger) (*Worker, error) {
	s, err := gocron.NewScheduler(gocron.WithLocation(time.Local))
	if err != nil {
		return nil, err
	}
	return &Worker{
		scheduler: s,
		logger:    log,
	}, nil
}

func (w *Worker) AddHandler(h Handler) {
	w.handlers = append(w.handlers, h)
}

func (w *Worker) RegisterCron(expression string, h Handler) error {
	_, err := w.scheduler.NewJob(
		gocron.CronJob(expression, false),
		gocron.NewTask(func() {
			w.safeRun(h)
		}),
	)
	if err != nil {
		return err
	}
	w.logger.Infof("[worker] registered cron job: %s -> %s", expression, h.Name())
	return nil
}

func (w *Worker) RegisterInterval(interval time.Duration, h Handler) error {
	_, err := w.scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(func() {
			w.safeRun(h)
		}),
	)
	if err != nil {
		return err
	}
	w.logger.Infof("[worker] registered interval job: %v -> %s", interval, h.Name())
	return nil
}

func (w *Worker) safeRun(h Handler) {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Errorf("[worker] job %s panicked: %v", h.Name(), r)
		}
	}()
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := h.Run(ctx); err != nil {
		w.logger.Errorf("[worker] job %s failed: %v (took %v)", h.Name(), err, time.Since(start))
		return
	}
	w.logger.Infof("[worker] job %s success (took %v)", h.Name(), time.Since(start))
}

func (w *Worker) Start() {
	w.logger.Infof("[worker] starting scheduler with %d handlers", len(w.handlers))
	w.processing = true
	w.scheduler.Start()
}

func (w *Worker) Stop(ctx context.Context) error {
	w.logger.Info("[worker] stopping scheduler")
	w.processing = false
	return w.scheduler.Shutdown()
}

func (w *Worker) RunOnce(name string) error {
	for _, h := range w.handlers {
		if h.Name() == name {
			w.safeRun(h)
			return nil
		}
	}
	return nil
}
