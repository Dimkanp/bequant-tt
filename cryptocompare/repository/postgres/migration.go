package postgres

func (r *Repository) Migrate() error {
	// TODO here must be migration system
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS pairs (
			id uuid NOT NULL,
			fsym text NOT NULL,
			tsym text NOT NULL,
			created timestamp with time zone NOT NULL,
			raw text NOT NULL,
			display text NOT NULL,
			PRIMARY KEY ( id )
		);
	`)
	return err
}
