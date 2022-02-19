package repository

import (
	"bequant-tt/core"
	"context"
)

type DB interface {
	Migrate() error
	Close() error

	Repository
}

type Repository interface {
	// BeginTx use this function when you directly need to begin new transaction in the DB.
	BeginTx(ctx context.Context) (Transaction, error)
	// WithTx use this function when you need to do few actions in transaction,
	// but it can be created earlier (add user to the chat after joining quest, accept invite to the chat or any other action
	// that must be handled in one transaction).
	// Nice to pass transaction between services.
	WithTx(ctx *context.Context) (t Transaction, finalize func(...*error), err error)
	// State returns Repository interface with Transaction behind if WithTx method was called earlier without error
	// or just Repository as is.
	State(ctx context.Context) Repository

	Pairs() PairsRepository
}

type Transaction interface {
	Repository

	Commit() error
	Rollback() error
}

type PairsRepository interface {
	Insert(ctx context.Context, pair *core.Pair) (*core.Pair, error)
	Latest(ctx context.Context, fsym, tsym string) (*core.Pair, error)
}
