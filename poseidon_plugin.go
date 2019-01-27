package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama/runtime"
	"github.com/mastern2k3/poseidon/graphql"
	"github.com/mastern2k3/poseidon/ui"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := graphql.RegisterGraphQL(initializer); err != nil {
		return err
	}
	if err := ui.RegisterUI(initializer); err != nil {
		return err
	}
	return nil
}
