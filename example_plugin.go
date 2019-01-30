package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/heroiclabs/nakama/runtime"

	"github.com/mastern2k3/poseidon/liveparams"
	"github.com/mastern2k3/poseidon/rpc"
)

var (
	myLiveInt    *int
	myLiveString *string

	testRoutes = []rpc.RPCRoute{
		&rpc.StringRoute{"test_live_int", testLiveInt},
	}
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := liveparams.RegisterLiveParameters(ctx, nk, initializer, func(reg liveparams.Registrar) {
		myLiveInt = reg.LiveInt("liveint", 12)
		myLiveString = reg.LiveString("livestring", "hello")
	}); err != nil {
		return err
	}
	if err := rpc.RegisterRoutes(initializer, testRoutes); err != nil {
		return err
	}
	return nil
}

func testLiveInt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	return fmt.Sprintf("%d %s", *myLiveInt, *myLiveString), nil
}
