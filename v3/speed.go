package pb

import (
	"fmt"
	"math"
	"time"

	"github.com/VividCortex/ewma"
)

var speedAddLimit = time.Second / 2

type Speed struct {
	ewma                  ewma.MovingAverage
	lastStateId           uint64
	prevValue, startValue int64
	prevTime, startTime   time.Time
}

func (s *Speed) value(state *State) float64 {
	if s.ewma == nil {
		s.ewma = ewma.NewMovingAverage()
	}
	if state.IsFirst() || state.Id() < s.lastStateId {
		s.reset(state)
		return 0
	}
	if state.Id() == s.lastStateId {
		return s.ewma.Value()
	}
	if state.IsFinished() {
		return s.absValue(state)
	}
	dur := state.Time().Sub(s.prevTime)
	if dur < speedAddLimit {
		return s.ewma.Value()
	}
	diff := math.Abs(float64(state.Value() - s.prevValue))
	lastSpeed := diff / dur.Seconds()
	s.prevTime = state.Time()
	s.prevValue = state.Value()
	s.lastStateId = state.Id()
	s.ewma.Add(lastSpeed)
	return s.ewma.Value()
}

func (s *Speed) reset(state *State) {
	s.lastStateId = state.Id()
	s.startTime = state.Time()
	s.prevTime = state.Time()
	s.startValue = state.Value()
	s.prevValue = state.Value()
	s.ewma = ewma.NewMovingAverage()
}

func (s *Speed) absValue(state *State) float64 {
	if dur := state.Time().Sub(s.startTime); dur > 0 {
		return float64(state.Value()) / dur.Seconds()
	}
	return 0
}

func GetSpeedObj(state *State) (s *Speed) {
	if sObj, ok := state.Get(speedObj).(*Speed); ok {
		return sObj
	}
	s = new(Speed)
	state.Set(speedObj, s)
	return
}

// ElementSpeed calculates current speed by EWMA
// Optionally can take one or two string arguments.
// First string will be used as value for format speed, default is "%s/s".
// Second string will be used when speed not available, default is "?/s"
// In template use as follows: {{speed .}} or {{speed . "%s per second"}} or {{speed . "%s ps" "..."}
var ElementSpeed ElementFunc = func(state *State, args ...string) string {
	sp := GetSpeedObj(state).value(state)
	if sp == 0 {
		return argsHelper(args).getNotEmptyOr(1, "?/s")
	}
	return fmt.Sprintf(argsHelper(args).getNotEmptyOr(0, "%s/s"), state.Format(int64(round(sp))))
}

var ElementFloatSpeed ElementFunc = func(state *State, args ...string) string {
	sp := GetSpeedObj(state).value(state)
	return fmt.Sprintf(argsHelper(args).getOr(0, "%f"), sp)
}
