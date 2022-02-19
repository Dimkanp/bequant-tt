package service

import (
	"context"

	"bequant-tt/core"
	"bequant-tt/cryptocompare/repository"
)

type appService struct {
	repository repository.Repository
	service    *Service
}

func newAppService(repository repository.Repository, service *Service) *appService {
	return &appService{
		repository: repository,
		service:    service,
	}
}

func (s *appService) Get(ctx context.Context, f, t []string) ([]*core.Compare, error) {
	compares, err := s.service.Syncer.Get(ctx, f, t)
	if err == nil {
		return compares, nil
	}

	for i := range f {
		c := &core.Compare{
			Fsym:  f[i],
			Tsyms: make(map[string]*core.CompareData),
		}

		for j := range t {
			pair, err := s.repository.Pairs().Latest(ctx, f[i], t[j])
			if err != nil {
				return nil, err
			}

			c.Tsyms[f[i]] = &core.CompareData{
				Raw:     pair.Raw,
				Display: pair.Display,
			}
		}

		compares = append(compares, c)
	}

	return compares, nil
}
