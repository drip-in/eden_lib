package concurrent

import (
	"sync"
	"sync/atomic"
)

var condPool sync.Pool

func init() {
	condPool.New = newCond
}

func newCond() interface{} {
	return &Cond{}
}

type Cond struct {
	threshold int32
	counter   int32
	signal    chan struct{}
}

func NewCond(threshold int32) *Cond {
	cond := condPool.Get().(*Cond)

	cond.threshold = threshold
	cond.signal = make(chan struct{}, 1)

	return cond
}

func (receiver *Cond) Notify() {
	if atomic.AddInt32(&receiver.counter, 1) == receiver.threshold {
		receiver.signal <- struct{}{}
	}
}

func (receiver *Cond) Wait() {
	<-receiver.signal
}

func (this *Cond) zero() {
	this.threshold = 0
	this.counter = 0
	this.signal = nil
}

func (this *Cond) Recycle() {
	this.zero()
	condPool.Put(this)
}
