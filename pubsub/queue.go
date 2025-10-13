package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	defaultRetry       = 12
	defaultConcurrency = 5
)

func NewAsyncQueue(cfg AsyncQueueConfig) Queue {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisURL})
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: cfg.RedisURL})
	return &AsyncQueue{client: client, inspector: inspector}
}

type AsyncQueueConfig struct {
	RedisURL    string `mapstructure:"aq_redis_url"`
	Concurrency int    `mapstructure:"concurrency"`
	Pool        int    `mapstructure:"pool"`
	TimeZone    string `mapstructure:"aq_timezone"`
	QueuesConfig
}

type QueuesConfig map[string]int

type Job[T JobArg] interface {
	Name() string
	Arg() JobArg
	Config() JobConfig
}

type AsyncJob[T JobArg] struct {
	Argument T
	JobName  string
	ID       string
	JobConfig
}

func (aj AsyncJob[T]) Name() string {
	return aj.JobName
}

func (aj AsyncJob[T]) Arg() JobArg {
	return aj.Argument
}

func (aj AsyncJob[T]) Config() JobConfig {
	if len(aj.ID) != 0 {
		aj.JobConfig.JobID = aj.ID
	}
	return aj.JobConfig
}

type Queue interface {
	Enqueue(Job[JobArg]) error
	EnqueueAfterTime(job Job[JobArg], duration time.Duration) error
	EnqueueAt(Job Job[JobArg], executedAt time.Time) error
	DeleteTask(queue, taskID string) error
}

type QueueType string

func (q QueueType) String() string {
	return string(q)
}

const (
	Critical  QueueType = "critical"
	Default   QueueType = "default"
	Migration QueueType = "migration"
)

type AsyncQueue struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

type JobArg interface {
	JSON() ([]byte, error)
}

func (aq AsyncQueue) DeleteTask(queue, taskID string) error {
	return aq.inspector.DeleteTask(queue, taskID)
}

func (aq AsyncQueue) Enqueue(job Job[JobArg]) error {
	return aq.EnqueueAfterTime(job, 0*time.Second)
}

func (aq AsyncQueue) EnqueueAfterTime(job Job[JobArg], duration time.Duration) error {
	payload, err := job.Arg().JSON()
	if err != nil {
		return err
	}
	task := asynq.NewTask(job.Name(), payload)
	opts := newAsynqOptions(job.Config())
	opts = append(opts, asynq.ProcessIn(duration))
	_, err = aq.client.Enqueue(task, opts...)
	return err
}

func (aq AsyncQueue) EnqueueAt(job Job[JobArg], t time.Time) error {
	payload, err := job.Arg().JSON()
	if err != nil {
		return err
	}
	task := asynq.NewTask(job.Name(), payload)
	opts := newAsynqOptions(job.Config())
	opts = append(opts, asynq.ProcessAt(t))
	_, err = aq.client.Enqueue(task, opts...)
	return err
}

func newAsynqOptions(cfg JobConfig) []asynq.Option {
	opts := make([]asynq.Option, 0)
	if cfg.Retry != nil {
		opts = append(opts, asynq.MaxRetry(*cfg.Retry))
	} else {
		opts = append(opts, asynq.MaxRetry(defaultRetry))
	}
	if len(cfg.Queue) > 0 {
		queueOpt := asynq.Queue(cfg.Queue.String())
		opts = append(opts, queueOpt)
	} else {
		queueOpt := asynq.Queue(Default.String())
		opts = append(opts, queueOpt)
	}
	if len(cfg.JobID) > 0 {
		opts = append(opts, asynq.TaskID(cfg.JobID))
	}
	if cfg.Timeout != 0 {
		opts = append(opts, asynq.Timeout(cfg.Timeout))
	}
	if cfg.Unique > 0 {
		opts = append(opts, asynq.Unique(cfg.Unique))
	}
	if cfg.Retention > 0 {
		opts = append(opts, asynq.Retention(cfg.Retention))
	}
	return opts
}

type JobConfig struct {
	Queue     QueueType
	Retry     *int // retry N times, nil for default of 12 times
	JobID     string
	Timeout   time.Duration
	Unique    time.Duration
	Retention time.Duration
}

type Worker[T JobArg] interface {
	JobName() string
	Process(ctx context.Context, args T) error
}

func NewAsyncServer(cfg AsyncQueueConfig) AsyncServer {
	redisCfg := asynq.RedisClientOpt{Addr: cfg.RedisURL}
	asynqCfg := asynq.Config{
		Concurrency: cfg.Concurrency,
		Queues:      cfg.QueuesConfig,
	}
	if cfg.Concurrency == 0 {
		asynqCfg.Concurrency = defaultConcurrency
	}
	if len(cfg.QueuesConfig) == 0 {
		// default queues config
		asynqCfg.Queues = map[string]int{
			Migration.String(): 1,
			Default.String():   10,
			Critical.String():  10,
		}
	}
	server := asynq.NewServer(redisCfg, asynqCfg)
	mux := asynq.NewServeMux()
	return AsyncServer{
		mux:      mux,
		server:   server,
		handlers: map[string]func(context.Context, *asynq.Task) error{},
	}
}

type AsyncServer struct {
	mux         *asynq.ServeMux
	server      *asynq.Server
	handlers    map[string]func(context.Context, *asynq.Task) error
	middlewares []asynq.MiddlewareFunc
}

func (app *AsyncServer) AddMiddleware(traceMiddleware asynq.MiddlewareFunc) {
	app.middlewares = append(app.middlewares, traceMiddleware)
}

func (app *AsyncServer) Run() {
	if len(app.middlewares) >= 0 {
		for _, mid := range app.middlewares {
			app.mux.Use(mid)
		}
	}
	app.registerWorkers(app.mux)
	if err := app.server.Run(app.mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (app *AsyncServer) Close() {}

func (app *AsyncServer) registerWorkers(mux *asynq.ServeMux) {
	for jobName, handler := range app.handlers {
		mux.HandleFunc(jobName, handler)
	}
}

func AddWorker[T JobArg](app *AsyncServer, wk Worker[T]) {
	app.handlers[wk.JobName()] = func(ctx context.Context, t *asynq.Task) error {
		var arg T
		if err := json.Unmarshal(t.Payload(), &arg); err != nil {
			return fmt.Errorf("failed to unmarshal task payload: %v", err)
		}
		return wk.Process(ctx, arg)
	}
}

func addMiddleware(app *AsyncServer, mw asynq.MiddlewareFunc) {
	app.mux.Use(mw)
}

func AddTraceMiddleware(app *AsyncServer) {
	addMiddleware(app, traceMiddleware)
}

var traceMiddleware = func(hlr asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		ddSpan, ctx := tracer.StartSpanFromContext(ctx, t.Type())
		defer ddSpan.Finish()
		err := hlr.ProcessTask(ctx, t)
		if err != nil {
			return err
		}
		return nil
	})
}
