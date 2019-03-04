package statemachine

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama/runtime"
)

type TimeoutExpiredError struct {
	TimeoutTime, ErrorInstant time.Time
	ExpiredState              *StateDef
}

func (e *TimeoutExpiredError) Error() string {
	return fmt.Sprintf(`state "%s" with timeout at %s expired at %s`, e.ExpiredState.Name, e.TimeoutTime, e.ErrorInstant)
}

type stateAction interface {
	apply(ctx context.Context, sm *StateMachine, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error)
}

type TransitionTo struct {
	Target string
}

func (a TransitionTo) apply(ctx context.Context, sm *StateMachine, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error) {

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
	if sm.CurrentState.Timeout != 0 {
		expireTime := time.Now().Add(sm.CurrentState.Timeout)
		sm.currentMeta.expireTime = &expireTime
	} else {
		sm.currentMeta.expireTime = nil
	}

	if target.OnEnter != nil {
		logger.Debug("OnEnter on state `%s`", target.Name)
		target.OnEnter(ctx, logger, db, nk, dispatcher, tick, state, messages)
	}

	return state, nil
}

type terminateAction struct{}

func (a terminateAction) apply(ctx context.Context, sm *StateMachine, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error) {
	return nil, nil
}

type stayAction struct{}

func (a stayAction) apply(ctx context.Context, sm *StateMachine, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error) {
	return state, nil
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
	Timeout time.Duration
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
	currentMeta  struct {
		expireTime *time.Time
	}
	states      map[string]*StateDef
	transitions map[fromToKey]*TransitionDef
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

	if sm.currentMeta.expireTime != nil {
		if now := time.Now(); sm.currentMeta.expireTime.Before(now) {
			return nil, &TimeoutExpiredError{
				ErrorInstant: now,
				TimeoutTime:  *sm.currentMeta.expireTime,
				ExpiredState: sm.CurrentState,
			}
		}
	}

	action := sm.CurrentState.OnLoop(ctx, logger, db, nk, dispatcher, tick, state, messages)

	return action.apply(ctx, sm, logger, db, nk, dispatcher, tick, state, messages)
}
