package statemachine

import (
	"context"
	"database/sql"
	"testing"
	"time"

	mk "github.com/golang/mock/gomock"
	"github.com/heroiclabs/nakama/runtime"

	"github.com/mastern2k3/poseidon/tests/mocks"
)

var (
	initState = &StateDef{
		Name: "init",
		OnLoop: func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) stateAction {
			if toCtx := ctx.Value(firstState); toCtx != nil {
				return TransitionTo{toCtx.(string)}
			}
			return TransitionTo{"next"}
		},
	}

	nextWithTimeout = &StateDef{
		Name:    "next_timeout",
		Timeout: time.Millisecond * 500,
		OnLoop: func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) stateAction {
			return Stay
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

type statemachineTestContextKey string

const (
	firstState statemachineTestContextKey = "first_state"
)

func makeStateMachine() *StateMachine {
	return NewStateMachine(
		[]*StateDef{initState, nextState, nextWithTimeout},
		[]*TransitionDef{initToNext},
	)
}

func assertCurrentStateName(t *testing.T, sm *StateMachine, expected string) {
	if sm.CurrentState.Name != expected {
		t.Fatalf(`expected current state to be "%s" but was "%s"`, expected, sm.CurrentState.Name)
	}
}

func TestTransitionTo(t *testing.T) {

	sm := makeStateMachine()
	someState := "test"
	ctrl := mk.NewController(t)
	logger := mocks.WithTestLogging(mocks.NewMockLogger(ctrl), t)
	ctx := context.Background()

	newState, err := sm.Loop(ctx, logger, nil, nil, nil, 0, someState, nil)
	if err != nil {
		t.Fatalf("error while looping state machine: %s", err)
	}
	someState = newState.(string)

	assertCurrentStateName(t, sm, nextState.Name)

	newState, err = sm.Loop(ctx, logger, nil, nil, nil, 0, someState, nil)
	if err != nil {
		t.Fatalf("error while looping state machine: %s", err)
	}
	if newState != nil {
		t.Fatalf("expected Loop to end with nil state, but was %s", newState)
	}
}

func TestTransitionToWithTimeout(t *testing.T) {

	sm := makeStateMachine()
	someState := "test"
	ctrl := mk.NewController(t)
	logger := mocks.WithTestLogging(mocks.NewMockLogger(ctrl), t)
	ctx := context.WithValue(context.Background(), firstState, nextWithTimeout.Name)

	newState, err := sm.Loop(ctx, logger, nil, nil, nil, 0, someState, nil)
	if err != nil {
		t.Fatalf("error while looping state machine: %s", err)
	}
	someState = newState.(string)

	assertCurrentStateName(t, sm, nextWithTimeout.Name)

	newState, err = sm.Loop(ctx, logger, nil, nil, nil, 0, someState, nil)
	if err != nil {
		t.Fatalf("error while looping state machine: %s", err)
	}
	someState = newState.(string)

	assertCurrentStateName(t, sm, nextWithTimeout.Name)

	time.Sleep(time.Second * 1)

	newState, err = sm.Loop(ctx, logger, nil, nil, nil, 0, someState, nil)
	if err == nil {
		t.Fatalf("expected error while looping state machine, got nil")
	}
	if _, is := err.(*TimeoutExpiredError); !is {
		t.Fatalf("expected error to be timeout error, got %T", err)
	}
}
