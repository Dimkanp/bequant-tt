package repository

import (
	"context"

	"bequant-tt/configuration"
	"bequant-tt/cryptocompare/repository/postgres"
)

type DatabaseRepository struct {
	DBRepository *postgres.Repository
	Config       configuration.DBConfig
}

func New(config configuration.DBConfig) (*DatabaseRepository, error) {
	repository := &DatabaseRepository{
		Config: config,
	}

	var err error
	repository.DBRepository, err = postgres.New(config)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func (d *DatabaseRepository) Migrate() error {
	return d.DBRepository.Migrate()
}

func (d *DatabaseRepository) Close() error {
	return d.DBRepository.Close()
}

func (d *DatabaseRepository) Pairs() PairsRepository {
	return d.DBRepository.Pairs()
}

func (d *DatabaseRepository) BeginTx(ctx context.Context) (Transaction, error) {
	tx, err := d.DBRepository.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	r := *d

	r.DBRepository = tx

	return &DBTx{
		DatabaseRepository: &r,
	}, nil
}

type DBTx struct {
	*DatabaseRepository
}

func (d *DBTx) Commit() error {
	return d.DBRepository.Commit()
}

func (d *DBTx) Rollback() error {
	return d.DBRepository.Rollback()
}
