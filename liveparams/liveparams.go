package liveparams

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	gql "github.com/graphql-go/graphql"
	"github.com/heroiclabs/nakama/runtime"

	"github.com/mastern2k3/poseidon/graphql"
	"github.com/mastern2k3/poseidon/rpc"
	"github.com/mastern2k3/poseidon/storage"
)

var (
	liveParametersAccessor = &storage.CollectionAccessor{
		CollectionID: "admin",
		KeyID:        "live_parameters",
		ModelFactory: func() interface{} { return new(LiveParamsModel) },
		DefaultFactory: func() interface{} {
			return &LiveParamsModel{
				Parameters: make(map[string]interface{}),
			}
		},
	}

	liveParametersRoutes = []rpc.RPCRoute{
		&rpc.JsonRoute{"liveparams_get", nil, getAll},
		&rpc.JsonRoute{"liveparams_set", func() interface{} { return new(SetLiveParam_Request) }, setLiveParam},
	}

	liveParameters = map[string]interface{}{}
)

var (
	liveParamType = gql.NewObject(gql.ObjectConfig{
		Name:        "LiveParameter",
		Description: "A live parameter registered in the server.",
		Fields: gql.Fields{
			"name": &gql.Field{
				Type:        gql.NewNonNull(gql.String),
				Description: "The name of the live parameter.",
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					return p.Source.(string), nil
				},
			},
			"value": &gql.Field{
				Type:        gql.NewNonNull(gql.String),
				Description: "The value of the live parameter.",
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					return GetLiveParamString(p.Source.(string))
				},
			},
		},
	})

	liveParamsField = &gql.Field{
		Description: "The live parameters registered in the server.",
		Type:        gql.NewNonNull(gql.NewList(gql.NewNonNull(liveParamType))),
		Resolve: func(p gql.ResolveParams) (interface{}, error) {
			names := []string{}
			for name := range liveParameters {
				names = append(names, name)
			}
			return names, nil
		},
	}
)

type LiveParamsModel struct {
	Parameters map[string]interface{} `json:"parameters"`
}

type Registrar interface {
	LiveInt(name string, defaultValue int) *int
	LiveFloat(name string, defaultValue float64) *float64
	LiveString(name string, defaultValue string) *string
	LiveBool(name string, defaultValue bool) *bool
}

type registrar struct {
	storageValues map[string]interface{}
	errors        []error
}

func (r *registrar) LiveInt(name string, defaultValue int) *int {
	liveValue := defaultValue
	val, has := r.storageValues[name]
	if has {
		var err bool
		liveValue, err = val.(int)
		if err {
			r.errors = append(r.errors, fmt.Errorf("could not produce live int `%s`, value in storage `%+v` is not castable to int", name, val))
		}
	}
	liveParam := &liveValue
	liveParameters[name] = liveParam
	return liveParam
}

func (r *registrar) LiveFloat(name string, defaultValue float64) *float64 {
	liveValue := defaultValue
	val, has := r.storageValues[name]
	if has {
		var err bool
		liveValue, err = val.(float64)
		if err {
			r.errors = append(r.errors, fmt.Errorf("could not produce live float `%s`, value in storage `%+v` is not castable to float", name, val))
		}
	}
	liveParam := &liveValue
	liveParameters[name] = liveParam
	return liveParam
}

func (r *registrar) LiveString(name string, defaultValue string) *string {
	liveValue := defaultValue
	val, has := r.storageValues[name]
	if has {
		var err bool
		liveValue, err = val.(string)
		if err {
			r.errors = append(r.errors, fmt.Errorf("could not produce live string `%s`, value in storage `%+v` is not castable to string", name, val))
		}
	}
	liveParam := &liveValue
	liveParameters[name] = liveParam
	return liveParam
}

func (r *registrar) LiveBool(name string, defaultValue bool) *bool {
	liveValue := defaultValue
	val, has := r.storageValues[name]
	if has {
		var err bool
		liveValue, err = val.(bool)
		if err {
			r.errors = append(r.errors, fmt.Errorf("could not produce live bool `%s`, value in storage `%+v` is not castable to bool", name, val))
		}
	}
	liveParam := &liveValue
	liveParameters[name] = liveParam
	return liveParam
}

func RegisterLiveParameters(ctx context.Context, nk runtime.NakamaModule, init runtime.Initializer, reg func(reg Registrar)) error {
	r, err := liveParametersAccessor.GetOrDefault(ctx, nk, "")
	if err != nil {
		return err
	}
	storageParams := r.(*LiveParamsModel)
	registrar := &registrar{storageParams.Parameters, make([]error, 0)}
	reg(registrar)
	if err := rpc.RegisterRoutes(init, liveParametersRoutes); err != nil {
		return err
	}
	if err := graphql.ConfigureRootQuery(func(rootQuery *gql.Object) error {
		rootQuery.AddFieldConfig("liveParams", liveParamsField)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func GetLiveParamString(name string) (string, error) {
	liveParam, has := liveParameters[name]
	if !has {
		return "", fmt.Errorf("cannot find a live parameter with name `%s`", name)
	}
	switch v := liveParam.(type) {
	case *int:
		return strconv.Itoa(*v), nil
	case *float64:
		return fmt.Sprintf("%f", *v), nil
	case *string:
		return *v, nil
	case *bool:
		return strconv.FormatBool(*v), nil
	default:
		return "", fmt.Errorf("cannot get live param of type `%T`", v)
	}
}

func SetLiveParam(name string, newValue string) error {
	liveParam, has := liveParameters[name]
	if !has {
		return fmt.Errorf("cannot find a live parameter with name `%s`", name)
	}
	switch v := liveParam.(type) {
	case *int:
		newInt, err := strconv.Atoi(newValue)
		if err != nil {
			return err
		}
		*v = newInt
	case *float64:
		newFloat, err := strconv.ParseFloat(newValue, 64)
		if err != nil {
			return err
		}
		*v = newFloat
	case *string:
		*v = newValue
	case *bool:
		newBool, err := strconv.ParseBool(newValue)
		if err != nil {
			return err
		}
		*v = newBool
	default:
		return fmt.Errorf("cannot set live param of type `%T`", v)
	}
	return nil
}

type SetLiveParam_Request struct {
	Name     string `json:"name"`
	NewValue string `json:"newValue"`
}

func setLiveParam(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, input interface{}) (interface{}, error) {
	req := input.(*SetLiveParam_Request)
	return nil, SetLiveParam(req.Name, req.NewValue)
}

func GetAll(ctx context.Context, nk runtime.NakamaModule) (map[string]interface{}, error) {
	return liveParameters, nil
}

func getAll(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, input interface{}) (interface{}, error) {
	return GetAll(ctx, nk)
}
