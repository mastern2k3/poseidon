package statemachine

import (
	"context"
	"database/sql"
	"testing"

	mk "github.com/golang/mock/gomock"
	"github.com/heroiclabs/nakama/runtime"

	"github.com/mastern2k3/poseidon/tests/mocks"
)

var (
	initState = &StateDef{
		Name: "init",
		OnLoop: func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) stateAction {
			return TransitionTo{"next", nil}
		},
	}

	nextState = &StateDef{
		Name: "next",
		OnLoop: func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) stateAction {
			return Terminate
		},
	}

	initToNext = &TransitionDef{
		From: "init",
		To:   "next",
		OnTransition: func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) {
			logger.Info("OnTransition :D")
		},
	}
)

func makeStateMachine() *StateMachine {
	return NewStateMachine(
		[]*StateDef{initState, nextState},
		[]*TransitionDef{initToNext},
	)
}

func TestTransitionTo(t *testing.T) {

	sm := makeStateMachine()
	someState := "test"
	ctrl := mk.NewController(t)
	logger := mocks.WithTestLogging(mocks.NewMockLogger(ctrl), t)
	ctx := context.Background()

	for {
		newState, err := sm.Loop(ctx, logger, nil, nil, nil, 0, someState, nil)
		if err != nil {
			t.Fatalf("error while looping state machine: %s", err)
		}
		if newState == nil {
			break
		}
		someState = newState.(string)
	}
}
