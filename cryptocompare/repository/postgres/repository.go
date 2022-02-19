package postgres

import (
	"bequant-tt/configuration"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

// go get storj.io/dbx
//go:generate dbx schema -d postgres ../dbx/db.dbx ../dbx
//go:generate dbx golang -d postgres -p postgres ../dbx/db.dbx .

const (
	postgres = "postgres"
)

type executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Repository struct {
	db      *DB
	tx      *Tx
	methods Methods
	// use state when you need to execute custom query in transaction
	state executor
}

func New(config configuration.DBConfig) (*Repository, error) {
	dbURL := fmt.Sprintf("%s://%s:%s@%s:%d/%s%s", postgres, config.UserName, config.Password, config.Host, config.Port, config.DBName, config.SSL)

	store, err := Open(postgres, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed opening database: %v", err)
	}
	log.Println(fmt.Sprintf("Connected to: %s %s", "db source", dbURL))

	serverDB := &Repository{
		db:      store,
		methods: store,
		state:   store,
	}

	return serverDB, nil
}

func (r *Repository) Pairs() PairsRepository {
	return PairsRepository{r}
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) BeginTx(ctx context.Context) (*Repository, error) {
	if r.db == nil {
		return nil, errors.New("db is not initialized")
	}

	tx, err := r.db.Open(ctx)
	if err != nil {
		return nil, err
	}

	ptx := *r

	ptx.tx = tx
	ptx.methods = tx
	ptx.state = tx.Tx

	return &ptx, nil
}

func (r *Repository) Commit() error {
	if r.tx == nil {
		return errors.New("begin transaction before commit it")
	}

	return r.tx.Commit()
}

func (r *Repository) Rollback() error {
	if r.tx == nil {
		return errors.New("begin transaction before rollback it")
	}

	return r.tx.Rollback()
}
