package service

import (
	"context"

	"bequant-tt/core"
	"bequant-tt/cryptocompare/repository"
)

type Service struct {
	App       AppService
	Scheduler SchedulerService
	Syncer    SyncerService
}

type ServicesConfig struct {
	Scheduler *SchedulerConfig
}

type Config struct {
	Repository repository.Repository
	Services   *ServicesConfig
}

func New(cfg *Config) *Service {
	s := &Service{}

	s.App = newAppService(cfg.Repository, s)
	s.Scheduler = newSchedulerService(cfg.Repository, s, cfg.Services.Scheduler)
	s.Syncer = newSyncerService(cfg.Repository, s)

	return s
}

func (s *Service) Run() error {
	return s.Scheduler.Start()
}

func (s *Service) Stop() error {
	return s.Scheduler.Stop()
}

type AppService interface {
	Get(ctx context.Context, f, t []string) ([]*core.Compare, error)
}

type SchedulerService interface {
	Start() error
	Stop() error
}

type SyncerService interface {
	UpdateData(ctx context.Context) error
	Get(ctx context.Context, f, t []string) ([]*core.Compare, error)
}
