package postgres

import (
	"bequant-tt/core"
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
)

type PairsRepository struct {
	*Repository
}

func (r PairsRepository) Insert(ctx context.Context, pair *core.Pair) (*core.Pair, error) {
	dbxModel, err := r.methods.Create_Pair(ctx,
		Pair_Id(pair.ID.String()),
		Pair_Fsym(pair.Fsym),
		Pair_Tsym(pair.Tsym),
		Pair_Created(pair.Created),
		Pair_Raw(pair.Raw),
		Pair_Display(pair.Display),
	)
	if err != nil {
		return nil, err
	}

	return r.fromDbx(dbxModel)
}

func (r PairsRepository) Latest(ctx context.Context, f, t string) (p *core.Pair, err error) {
	dbx, err := r.methods.Get_Pair_By_Fsym_And_Tsym_OrderBy_Desc_Created(ctx,
		Pair_Fsym(f),
		Pair_Tsym(t),
	)
	if err != nil {
		return nil, err
	}

	return r.fromDbx(dbx)
}

func (r PairsRepository) fromDbx(pair *Pair) (_ *core.Pair, err error) {
	if pair == nil {
		return nil, fmt.Errorf("pair parameter is nil")
	}

	id, err := uuid.FromString(pair.Id)
	if err != nil {
		return nil, err
	}

	return &core.Pair{
		ID:      id,
		Fsym:    pair.Fsym,
		Tsym:    pair.Tsym,
		Created: pair.Created,
		Raw:     pair.Raw,
		Display: pair.Display,
	}, nil
}

func (r PairsRepository) fromDbxSlice(categoriesDbx []*Pair) (_ []*core.Pair, err error) {
	var pairs []*core.Pair

	// Generating []dbo from []dbx and collecting all errors
	for _, dbx := range categoriesDbx {
		quest, err := r.fromDbx(dbx)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, quest)
	}

	return pairs, nil
}
