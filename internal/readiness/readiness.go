package readiness

import "sync/atomic"

type State struct{ ready atomic.Bool }

func (s *State) MarkReady()    { s.ready.Store(true) }
func (s *State) BeginDrain()   { s.ready.Store(false) }
func (s *State) IsReady() bool { return s.ready.Load() }
