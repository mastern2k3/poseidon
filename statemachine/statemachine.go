package statemachine

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama/runtime"
)

const (
	InitialStateName = "init"
)

// TimeoutExpiredError is returned from StateMachine's Loop when the timeout duration defined on a StateDef expires
type TimeoutExpiredError struct {
	TimeoutTime, ErrorInstant time.Time
	ExpiredState              *StateDef
}

func (e *TimeoutExpiredError) Error() string {
	return fmt.Sprintf(`state "%s" with timeout at %s expired at %s`, e.ExpiredState.Name, e.TimeoutTime, e.ErrorInstant)
}

type stateAction interface {
	apply(ctx context.Context, s State, sm *StateMachine, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error)
}

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

type State interface {
	State() *StateData
}

type StateData struct {
	CurrentState *StateDef
	ExpireTime   *time.Time
}

func (s *StateData) State() *StateData {
	return s
}

type StateMachine struct {
	states      map[string]*StateDef
	transitions map[fromToKey]*TransitionDef
}

// NewStateMachine creates a StateMachine
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

	if _, has := new.states[InitialStateName]; !has {
		panic(fmt.Sprintf(`state machine must include StateDef with name "%s"`, InitialStateName))
	}

	return new
}

func (sm *StateMachine) Loop(ctx context.Context, st State, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, error) {

	s := st.State()

	if s.CurrentState == nil {
		s.CurrentState = sm.states[InitialStateName]
	}

	if s.ExpireTime != nil {
		if now := time.Now(); s.ExpireTime.Before(now) {
			return nil, &TimeoutExpiredError{
				ErrorInstant: now,
				TimeoutTime:  *s.ExpireTime,
				ExpiredState: s.CurrentState,
			}
		}
	}

	action := s.CurrentState.OnLoop(ctx, logger, db, nk, dispatcher, tick, state, messages)

	return action.apply(ctx, s, sm, logger, db, nk, dispatcher, tick, state, messages)
}
