package gopool

import (
	"context"
	"sync"
	"sync/atomic"
)

// Option represents the optional function.
type Option func(opts *Options)

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}

type Pool interface {
	// 更新 goroutine pool 的容量
	SetCap(cap int32)
	// 执行 f
	Go(f func())
	// 传入 ctx 和 f，panic 打日志时带上 logid
	CtxGo(ctx context.Context, f func())
	// panic 的时候调用额外的 handler
	SetPanicHandler(f func(context.Context, interface{}))
	// 获取当前正在运行的 goroutine 数量
	WorkerCount() int32
	// Close 会停止接收新的任务，等到旧的任务全部执行完成之后，所有的 worker 会自动退出
	Close()
}

var taskPool sync.Pool

func init() {
	taskPool.New = newTask
}

type task struct {
	ctx context.Context
	f   func()
}

func (t *task) zero() {
	t.ctx = nil
	t.f = nil
}

func (t *task) Recycle() {
	t.zero()
	taskPool.Put(t)
}

func newTask() interface{} {
	return &task{}
}

type pool struct {
	// pool 的容量
	cap int32
	// 配置信息
	options *Options
	// 任务管道
	taskCh    chan *task
	taskLock  sync.Mutex
	taskCount int32

	// 记录正在运行的 worker 数量
	workerCount int32

	// 用来标记是否关闭
	closed int32

	// worker panic 的时候会调用这个方法
	panicHandler func(context.Context, interface{})
}

func NewPool(cap int32, options ...Option) Pool {
	p := &pool{
		cap:     cap,
		taskCh:  make(chan *task, 1),
		options: loadOptions(options...),
	}
	return p
}

func (p *pool) SetCap(cap int32) {
	atomic.StoreInt32(&p.cap, cap)
}

func (p *pool) Go(f func()) {
	p.CtxGo(context.Background(), f)
}

func (p *pool) CtxGo(ctx context.Context, f func()) {
	t := taskPool.Get().(*task)
	t.ctx = ctx
	t.f = f
	p.taskCh <- t
	//fmt.Printf("task: %+v", t)
	atomic.AddInt32(&p.taskCount, 1)
	// 如果 pool 已经被关闭了，就 panic
	if atomic.LoadInt32(&p.closed) == 1 {
		panic("use closed pool")
	}
	// 满足以下任意一个条件：
	// 1. 目前的 worker 数量小于上限 p.cap
	// 2. 目前没有 worker
	if p.WorkerCount() < atomic.LoadInt32(&p.cap) || p.WorkerCount() == 0 {
		p.incWorkerCount()
		w := workerPool.Get().(*worker)
		w.pool = p
		w.run()
	}
}

func (p *pool) SetPanicHandler(f func(context.Context, interface{})) {
	p.panicHandler = f
}

func (p *pool) WorkerCount() int32 {
	return atomic.LoadInt32(&p.workerCount)
}

// Close 会停止接收新的任务，等到旧的任务全部执行完成之后，所有的 worker 会自动退出
func (p *pool) Close() {
	atomic.StoreInt32(&p.closed, 1)
	p.taskCh <- nil
}

func (p *pool) incWorkerCount() {
	atomic.AddInt32(&p.workerCount, 1)
}

func (p *pool) decWorkerCount() {
	atomic.AddInt32(&p.workerCount, -1)
}

// Options contains all options which will be applied when instantiating an ants pool.
type Options struct {
	// PanicHandler is used to handle panics from each worker goroutine.
	// if nil, panics will be thrown out again from worker goroutines.
	PanicHandler func(interface{})

	// Logger is the customized logger for logging info, if it is not set,
	// default standard logger from log package is used.
	Logger Logger
}
