package statemachine

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama/runtime"
)

type stateAction interface{}

type TransitionTo struct {
	Target string
	stateAction
}

type terminateAction struct {
	stateAction
}

type stayAction struct {
	stateAction
}

var (
	Terminate = terminateAction{}
	Stay      = stayAction{}
)

type StateDef struct {
	Name    string
	OnEnter func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData)
	OnLoop  func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) stateAction
	OnExit  func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData)
	Timeout *time.Duration
}

type TransitionDef struct {
	From, To     string
	OnTransition func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData)
}

type fromToKey struct {
	from, to string
}

type StateMachine struct {
	CurrentState *StateDef
	states       map[string]*StateDef
	transitions  map[fromToKey]*TransitionDef
}

func NewStateMachine(states []*StateDef, transitions []*TransitionDef) *StateMachine {

	new := &StateMachine{
		states:      make(map[string]*StateDef),
		transitions: make(map[fromToKey]*TransitionDef),
	}

	for _, state := range states {
		new.states[state.Name] = state
	}

	for _, transition := range transitions {
		new.transitions[fromToKey{transition.From, transition.To}] = transition
	}

	init, has := new.states["init"]
	if !has {
		panic(`state machine must include "init" state`)
	}

	new.CurrentState = init

	return new
}

func (sm *StateMachine) Loop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error) {

	action := sm.CurrentState.OnLoop(ctx, logger, db, nk, dispatcher, tick, state, messages)

	switch a := action.(type) {
	case TransitionTo:

		logger.Info("state transition from state `%s` to `%s`", sm.CurrentState.Name, a.Target)

		if sm.CurrentState.OnExit != nil {
			logger.Debug("OnExit on state `%s`", sm.CurrentState.Name)
			sm.CurrentState.OnExit(ctx, logger, db, nk, dispatcher, tick, state, messages)
		}

		transition, has := sm.transitions[fromToKey{sm.CurrentState.Name, a.Target}]
		if has {
			logger.Debug("OnTransition from: `%s` to: `%s`", sm.CurrentState.Name, a.Target)
			transition.OnTransition(ctx, logger, db, nk, dispatcher, tick, state, messages)
		}

		target, has := sm.states[a.Target]
		if !has {
			panic(fmt.Sprintf("attempt transition to nonexistant state `%s`", a.Target))
		}

		sm.CurrentState = target

		if target.OnEnter != nil {
			logger.Debug("OnEnter on state `%s`", target.Name)
			target.OnEnter(ctx, logger, db, nk, dispatcher, tick, state, messages)
		}

	case terminateAction:
		return nil, nil
	}

	return state, nil
}
