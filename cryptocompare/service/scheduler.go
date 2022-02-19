package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"bequant-tt/cryptocompare/repository"
)

type SchedulerConfig struct {
	RefreshTime time.Duration
}

type schedulerService struct {
	service    *Service
	repository repository.Repository
	config     *SchedulerConfig

	ticker  *time.Ticker
	started chan bool
	done    chan bool
}

func newSchedulerService(repository repository.Repository, service *Service, config *SchedulerConfig) *schedulerService {
	return &schedulerService{
		repository: repository,
		service:    service,
		config:     config,
		done:       make(chan bool),
		started:    make(chan bool, 1),
	}
}

func (s *schedulerService) Start() (err error) {
	select {
	case <-s.started:
		s.started <- true
		return fmt.Errorf("already started")
	default: // When started chan is blocked, default unlock goroutine and allow us to execute code below.
	}

	err = s.service.Syncer.UpdateData(context.Background())
	if err != nil {
		return err
	}

	s.ticker = time.NewTicker(s.config.RefreshTime)
	s.started <- true
	for {
		select {
		case <-s.done:
			s.ticker.Stop()
			s.ticker = nil
			return
		case <-s.ticker.C:
			err := s.service.Syncer.UpdateData(context.Background())
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}

func (s *schedulerService) Stop() (err error) {
	select {
	case <-s.started:
	default:
		return fmt.Errorf("already stopped")
	}

	s.done <- true
	return nil
}
