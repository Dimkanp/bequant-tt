package repository

import (
	"context"
)

type transactioner struct {
	Transaction

	// Must be a pointer to give Commit method ability to change its value.
	// Commit method cannot receive pointer to transactioner [ (t *transactioner) Commit()... ]
	// because of collision with Transaction interface.
	commitTx *bool
	// done marks that finalize function was
	done bool
}

type keyType int

const key keyType = 0

// WithTx makes transactioner that control transaction state,
// if current transaction already defined it does nothing because we don't have to control a transaction that we didn't create
// otherwise it creates a new transaction and commit or rollback changes at the transactioner.Finalize function
//
// --- ATTENTION ---
//
// - tx.Commit() will not commit changes, it just notifies finalize function to commit changes instead rollback.
//
// - tx.Rollback() has no effect, use finalize function without calling tx.Commit() before.
//
// - finalize function is safe for multiple calls, like in defer, to be sure that changes will be rolled back on error,
// and directly in code. You can pass pointer(s) to an error to receive resulting error,
// if value of pointer isn't nil it will not be replaced.
func (d *DatabaseRepository) WithTx(ctx *context.Context) (tx Transaction, finalize func(...*error), err error) {
	fal := false
	t := transactioner{
		commitTx: &fal,
		done:     false,
	}

	tr, ok := (*ctx).Value(key).(*transactioner)
	if ok && tr != nil && !tr.done {
		t.Transaction = tr.Transaction
		return t, func(...*error) {}, nil
	}

	t.Transaction, err = d.BeginTx(*ctx)
	if err != nil {
		return nil, nil, err
	}

	*ctx = context.WithValue(*ctx, key, &t)

	return t, t.Finalize, nil
}

func (d *DatabaseRepository) State(ctx context.Context) Repository {
	tr, ok := (ctx).Value(key).(*transactioner)
	if ok && tr != nil && !tr.done {
		return tr
	}

	return d
}

// Finalize makes  rollback or commit if Commit() method was called
func (t *transactioner) Finalize(eps ...*error) {
	if t.done {
		return
	}
	t.done = true

	var err error
	if *t.commitTx {
		err = t.Transaction.Commit()
	} else {
		err = t.Transaction.Rollback()
	}

	if err != nil {
		for i := range eps {
			if eps[i] != nil && // error pointer isn't nil, and we can store our error into it
				*eps[i] == nil { // if error happened before don't replace it
				*eps[i] = err
			}
		}
	}
}

// Commit notifies transactioner that Finalize function must commit changes instead rollback,
// don't commit changes directly.
func (t transactioner) Commit() (_ error) {
	*t.commitTx = true
	return
}

// Rollback just blocks direct use of t.Transaction.Rollback method, has no effect,
//use finalize function (without calling Commit() before) instead.
func (t transactioner) Rollback() error {
	return nil
}
