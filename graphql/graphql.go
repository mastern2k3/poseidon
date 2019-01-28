package graphql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/heroiclabs/nakama/api"
	"github.com/heroiclabs/nakama/runtime"

	"github.com/mastern2k3/poseidon/rpc"
)

type ContextKey string

const (
	GRAPHQL_CTX_NAKAMA_MODULE ContextKey = "nakama_module"
)

type ledgerItemChangesetItem = struct {
	key   string
	value float64
}

type ledgerItemMetadataItem = struct {
	key   string
	value string
}

var (
	ledgerItemChangesetItemType = graphql.NewObject(graphql.ObjectConfig{
		Name: "LedgerItemChangesetItem",
		// Description: "A storage object persisted on Nakama.",
		Fields: graphql.Fields{
			"key": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(ledgerItemChangesetItem).key, nil
				},
			},
			"value": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Float),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(ledgerItemChangesetItem).value, nil
				},
			},
		},
	})

	ledgerItemMetadataItemType = graphql.NewObject(graphql.ObjectConfig{
		Name: "LedgerItemMetadataItem",
		// Description: "A storage object persisted on Nakama.",
		Fields: graphql.Fields{
			"key": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(ledgerItemChangesetItem).key, nil
				},
			},
			"value": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(ledgerItemChangesetItem).value, nil
				},
			},
		},
	})

	ledgerItemType = graphql.NewObject(graphql.ObjectConfig{
		Name: "LedgerItem",
		// Description: "A storage object persisted on Nakama.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The ledger item Id.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(runtime.WalletLedgerItem).GetID(), nil
				},
			},
			"createTime": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.DateTime),
				Description: "The ledger item creation time.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return time.Unix(p.Source.(runtime.WalletLedgerItem).GetCreateTime(), 0), nil
				},
			},
			"updateTime": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.DateTime),
				Description: "The ledger update time.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return time.Unix(p.Source.(runtime.WalletLedgerItem).GetUpdateTime(), 0), nil
				},
			},
			"changeset": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(ledgerItemChangesetItemType))),
				Description: "The ledger item changeset.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					changeset := []ledgerItemChangesetItem{}
					for k, v := range p.Source.(runtime.WalletLedgerItem).GetChangeset() {
						changeset = append(changeset, ledgerItemChangesetItem{k, v.(float64)})
					}
					return changeset, nil
				},
			},
			"metadata": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(ledgerItemMetadataItemType))),
				Description: "The ledger item metadata.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					metadata := []ledgerItemMetadataItem{}
					for k, v := range p.Source.(runtime.WalletLedgerItem).GetMetadata() {
						metadata = append(metadata, ledgerItemMetadataItem{k, v.(string)})
					}
					return metadata, nil
				},
			},
			// "user": &graphql.Field{
			// 	Type:        graphql.NewNonNull(graphql.String),
			// 	Description: "The ledger item Id.",
			// 	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 		return p.Source.(runtime.WalletLedgerItem).GetUserID(), nil
			// 	},
			// },
		},
	})

	storageType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "StorageObject",
		Description: "A storage object persisted on Nakama.",
		Fields: graphql.Fields{
			"key": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The key defining the stored object.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*api.StorageObject).GetKey(), nil
				},
			},
			"value": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The value stored in the object.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*api.StorageObject).GetValue(), nil
				},
			},
		},
	})

	accountType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Account",
		Description: "A registered Nakama user account.",
		Fields: graphql.Fields{
			"customId": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The custom id of the account.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*api.Account).GetCustomId(), nil
				},
			},
			"email": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The email of the account.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*api.Account).GetEmail(), nil
				},
			},
			"wallet": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The wallet of the account.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*api.Account).GetWallet(), nil
				},
			},
		},
	})

	userType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "User",
		Description: "A registered Nakama user.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The id of the user.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*api.User).GetId(), nil
				},
			},
			"username": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The username of the user.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(*api.User).GetUsername(), nil
				},
			},
			"account": &graphql.Field{
				Type:        graphql.NewNonNull(accountType),
				Description: "The account of the user.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					nk := p.Context.Value(GRAPHQL_CTX_NAKAMA_MODULE).(runtime.NakamaModule)
					acc, err := nk.AccountGetId(p.Context, p.Source.(*api.User).GetId())
					if err != nil {
						return nil, err
					}
					return acc, nil
				},
			},
			"storage": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(storageType))),
				Args: graphql.FieldConfigArgument{
					"collection": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Description: "The account of the user.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					nk := p.Context.Value(GRAPHQL_CTX_NAKAMA_MODULE).(runtime.NakamaModule)
					objs, _, err := nk.StorageList(p.Context, p.Source.(*api.User).GetId(), p.Args["collection"].(string), 10, "")
					if err != nil {
						return nil, err
					}
					return objs, nil
				},
			},
			"ledger": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(ledgerItemType))),
				Description: "The user's wallet transaction ledger.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					nk := p.Context.Value(GRAPHQL_CTX_NAKAMA_MODULE).(runtime.NakamaModule)
					items, err := nk.WalletLedgerList(p.Context, p.Source.(*api.User).GetId())
					if err != nil {
						return nil, err
					}
					return items, nil
				},
			},
		},
	})

	rootQuery = graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"userByUsername": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"username": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					nk := p.Context.Value(GRAPHQL_CTX_NAKAMA_MODULE).(runtime.NakamaModule)
					usernameParam := p.Args["username"].(string)
					users, err := nk.UsersGetUsername(p.Context, []string{usernameParam})
					if err != nil {
						return nil, err
					}
					if len(users) < 1 {
						return nil, fmt.Errorf("no user with username `%s`", usernameParam)
					}
					return users[0], nil
				},
			},
		},
	})

	schemaConfig = graphql.SchemaConfig{Query: rootQuery}
	schema       graphql.Schema
)

var (
	graphQLRoutes = []rpc.RPCRoute{
		&rpc.JsonRoute{"graphql", func() interface{} { return new(GraphQLRequest) }, query},
	}
)

func ConfigureRootQuery(conf func(rootQuery *graphql.Object) error) error {
	return conf(rootQuery)
}

func RegisterGraphQL(init runtime.Initializer) error {
	var err error
	schema, err = graphql.NewSchema(schemaConfig)
	if err != nil {
		return err
	}
	return rpc.RegisterRoutes(init, graphQLRoutes)
}

type GraphQLRequest struct {
	Query string `json:"query"`
}

func query(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, request interface{}) (interface{}, error) {
	query := request.(*GraphQLRequest)
	logger.Info("query: %+v %s", query, runtime.RUNTIME_CTX_MATCH_NODE)
	newCtx := context.WithValue(ctx, GRAPHQL_CTX_NAKAMA_MODULE, nk)
	params := graphql.Params{
		Schema:        schema,
		RequestString: query.Query,
		Context:       newCtx,
	}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		logger.Error("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	return r, nil
}
