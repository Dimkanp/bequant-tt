package service

import (
	"bequant-tt/core"
)

type Service struct {
}

func New() *Service {
	return &Service{}
}

type App interface {
	Get(f, t []string) (*core.Pairs, error)
}

type Scheduler interface {
	Start()
}

type Syncer interface {
	UpdateData() error
	Get(f, t []string) (*core.Pairs, error)
}
