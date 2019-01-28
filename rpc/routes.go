package rpc

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama/runtime"
)

type RPCRoute interface {
	Register(init runtime.Initializer) error
}

type StringRoute struct {
	Name    string
	Handler func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error)
}

func (h *StringRoute) Register(init runtime.Initializer) error {
	return init.RegisterRpc(h.Name, h.Handler)
}

type JsonRoute struct {
	Name       string
	InputModel func() interface{}
	Handler    func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, input interface{}) (interface{}, error)
}

func (h *JsonRoute) Register(init runtime.Initializer) error {
	return init.RegisterRpc(h.Name, func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, inputJson string) (string, error) {

		var inputModel interface{}

		if h.InputModel != nil {
			inputModel = h.InputModel()

			if err := json.Unmarshal([]byte(inputJson), inputModel); err != nil {
				return "", err
			}
		}

		outputModel, err := h.Handler(ctx, logger, db, nk, inputModel)

		if err != nil {
			return "", err
		}

		bytes, err := json.Marshal(outputModel)

		if err != nil {
			return "", err
		}

		return string(bytes), nil
	})
}

func RegisterRoutes(init runtime.Initializer, routes []RPCRoute) error {
	for _, route := range routes {
		if err := route.Register(init); err != nil {
			return err
		}
	}
	return nil
}
