package statemachine

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama/runtime"
)

// TransitionTo is a stateAction used to transition between states
type TransitionTo struct {
	Target string
}

func (a TransitionTo) apply(ctx context.Context, st State, sm *StateMachine, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error) {

	s := st.State()

	logger.Info("state transition from state `%s` to `%s`", s.CurrentState.Name, a.Target)

	if s.CurrentState.OnExit != nil {
		logger.Debug("OnExit on state `%s`", s.CurrentState.Name)
		s.CurrentState.OnExit(ctx, logger, db, nk, dispatcher, tick, state, messages)
	}

	transition, has := sm.transitions[fromToKey{s.CurrentState.Name, a.Target}]
	if has {
		logger.Debug("OnTransition from: `%s` to: `%s`", s.CurrentState.Name, a.Target)
		transition.OnTransition(ctx, logger, db, nk, dispatcher, tick, state, messages)
	}

	target, has := sm.states[a.Target]
	if !has {
		panic(fmt.Sprintf("attempt transition to nonexistant state `%s`", a.Target))
	}

	s.CurrentState = target
	if s.CurrentState.Timeout != 0 {
		expireTime := time.Now().Add(s.CurrentState.Timeout)
		s.ExpireTime = &expireTime
	} else {
		s.ExpireTime = nil
	}

	if target.OnEnter != nil {
		logger.Debug("OnEnter on state `%s`", target.Name)
		target.OnEnter(ctx, logger, db, nk, dispatcher, tick, state, messages)
	}

	return state, nil
}

type terminateAction struct{}

func (a terminateAction) apply(ctx context.Context, s State, sm *StateMachine, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error) {
	return nil, nil
}

type stayAction struct{}

func (a stayAction) apply(ctx context.Context, s State, sm *StateMachine, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error) {
	return state, nil
}

var (
	// Terminate is a stateAction used to signal StateMachine to terminate execution
	Terminate = terminateAction{}

	// Stay is a stateAction used to signal StateMachine to remain in the current state
	Stay = stayAction{}
)
