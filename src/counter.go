package slapperx

import "sync/atomic"

type counter int64

func (c *counter) Add(v int64) int64 { return atomic.AddInt64((*int64)(c), v) }
func (c *counter) Load() int64       { return atomic.LoadInt64((*int64)(c)) }
func (c *counter) Store(v int64)     { atomic.StoreInt64((*int64)(c), v) }
