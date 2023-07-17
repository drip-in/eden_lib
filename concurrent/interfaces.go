package concurrent

import (
	"context"
	"github.com/drip-in/eden_lib/el_utils"
	"sync"

	"github.com/drip-in/eden_lib/gopool"
	"github.com/drip-in/eden_lib/logs"
)

var pool gopool.Pool

func init() {
	pool = gopool.NewPool(1000)
	pool.SetPanicHandler(func(ctx context.Context, e interface{}) {
		logs.CtxError(ctx, "panic", logs.String("err", el_utils.ToJsonString(e)))
	})
}

type ITask interface {
	Name() string

	ShouldDo(ctx context.Context, param, taskContext interface{}) bool
	Load(ctx context.Context, param, taskContext interface{}) error
}

type IPack interface {
	Pack(ctx context.Context, param, taskContext interface{}) error
}

type packComponent struct {
	IPack
	prior int
}

var taskMap sync.Map

func registerTask(task ITask) {
	taskMap.Store(task.Name(), task)
}
