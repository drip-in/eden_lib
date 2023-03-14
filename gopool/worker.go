package gopool

import (
	"sync"
	"sync/atomic"
)

var workerPool sync.Pool

func init() {
	workerPool.New = newWorker
}

type worker struct {
	pool *pool
}

func newWorker() interface{} {
	return &worker{}
}

func (w *worker) run() {
	go func() {
		for t := range w.pool.taskCh {
			if t == nil {
				// 如果没有任务要做了，就释放资源，退出
				w.close()
				w.Recycle()
				return
			}

			atomic.AddInt32(&w.pool.taskCount, -1)
			func() {
				defer func() {
					if r := recover(); r != nil {
						w.pool.panicHandler(t.ctx, r)
					}
				}()
				t.f()
			}()
			t.Recycle()
		}
	}()
}

func (w *worker) close() {
	w.pool.decWorkerCount()
}

func (w *worker) zero() {
	w.pool = nil
}

func (w *worker) Recycle() {
	w.zero()
	workerPool.Put(w)
}
