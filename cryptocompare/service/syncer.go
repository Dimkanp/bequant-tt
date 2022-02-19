package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bequant-tt/core"
	"bequant-tt/cryptocompare/repository"

	uuid "github.com/satori/go.uuid"
)

const url = "https://min-api.cryptocompare.com/data/pricemultifull"

var (
	fsyms = []string{"BTC", "XRP", "ETH", "BCH", "EOS", "LTC", "XMR", "DASH"}
	tsyms = []string{"USD", "EUR", "GBP", "JPY", "RUR"}
)

type syncerService struct {
	repository repository.Repository
	service    *Service
}

func newSyncerService(repository repository.Repository, service *Service) *syncerService {
	return &syncerService{
		repository: repository,
		service:    service,
	}
}

func (s *syncerService) UpdateData(ctx context.Context) error {
	_, err := s.service.Syncer.Get(ctx, fsyms, tsyms)
	return err
}

func (s *syncerService) Get(ctx context.Context, fs, ts []string) ([]*core.Compare, error) {
	httpClient := http.DefaultClient

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Accept", "application/json")

	q := request.URL.Query()

	for _, f := range fs {
		q.Add("fsyms", f)
	}
	for _, t := range ts {
		q.Add("tsyms", t)
	}

	request.URL.RawQuery = q.Encode()

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	var res result
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	compares := make([]*core.Compare, len(fs))

	for i, f := range fs {
		compares[i] = &core.Compare{
			Fsym:  f,
			Tsyms: make(map[string]*core.CompareData),
		}

		for _, t := range ts {
			pair := &core.Pair{
				Fsym: f,
				Tsym: t,
			}

			pair.Raw, err = filterFields(res.Raw[f][t])
			if err != nil {
				return nil, err
			}

			pair.Display, err = filterFields(res.Display[f][t])
			if err != nil {
				return nil, err
			}

			go func(pair *core.Pair) {
				err = s.store(ctx, pair)
				if err != nil {
					log.Println(err)
				}
			}(pair)

			compares[i].Tsyms[t] = &core.CompareData{
				Raw:     pair.Raw,
				Display: pair.Display,
			}
		}
	}

	return compares, nil
}

func (s *syncerService) store(ctx context.Context, pair *core.Pair) error {
	pair.ID = uuid.NewV4()
	_, err := s.repository.Pairs().Insert(ctx, pair)
	return err
}

type result struct {
	Raw     map[string]map[string]map[string]interface{} `json:"RAW"`
	Display map[string]map[string]map[string]interface{} `json:"DISPLAY"`
}

var targetFields = []string{
	"CHANGE24HOUR",
	"CHANGEPCT24HOUR",
	"OPEN24HOUR",
	"VOLUME24HOUR",
	"VOLUME24HOURTO",
	"LOW24HOUR",
	"HIGH24HOUR",
	"PRICE",
	"SUPPLY",
	"MKTCAP",
}

func filterFields(in map[string]interface{}) (string, error) {
	out := make(map[string]interface{})

	for _, field := range targetFields {
		value, ok := in[field]
		if !ok {
			return "", fmt.Errorf("field '%s' not found", field)
		}

		out[field] = value
	}

	bytes, err := json.Marshal(out)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
