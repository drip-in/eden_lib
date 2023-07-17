package concurrent

import (
	"context"
	"fmt"
	"github.com/drip-in/eden_lib/el_utils"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/drip-in/eden_lib/logs"
)

var nodePool sync.Pool

func init() {
	nodePool.New = newNode
}

// node
type node struct {
	taskName   string
	graphName  string
	cond       *Cond
	gi         *graphInstance
	core       bool
	shouldSkip int32

	metricsTaskCost bool
	loggerTaskCost  bool
}

func newNode() interface{} {
	return &node{}
}

func (this *node) zero() {
	this.taskName = ""
	this.graphName = ""
	this.cond = nil
	this.gi = nil
	this.core = false
	this.shouldSkip = 0
	this.metricsTaskCost = false
	this.loggerTaskCost = false
}

func (this *node) needStopWatch() bool {
	return this.metricsTaskCost || this.loggerTaskCost
}

func (this *node) stopWatchIfNeed(ctx context.Context, start time.Time, err error, accumulate bool) {
	if !this.needStopWatch() {
		return
	}

	costMs := time.Since(start).Milliseconds()

	if this.loggerTaskCost {
		logs.CtxInfo(ctx, "", logs.String("graph", this.graphName), logs.String("task", this.taskName), logs.String("cost", fmt.Sprintf("%v ms", costMs)), logs.String("err", err.Error()))
	}
}

func (this *node) recycle() {
	if this.cond != nil {
		this.cond.Recycle()
	}

	this.zero()
	nodePool.Put(this)
}

func (this *node) execute(ctx context.Context, param interface{}, taskContext interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logs.CtxWarn(ctx, "", logs.String("graph", this.gi.g.name), logs.String("execute node", this.taskName), logs.ByteString("stack", debug.Stack()))
			err = fmt.Errorf("panic occurs in task:%s", this.taskName)
		}

		if atomic.AddInt32(&this.gi.counter, -1) == 0 {
			this.gi.signal <- struct{}{}
		}
	}()

	var start time.Time
	if this.needStopWatch() {
		start = time.Now()
	}

	//等待前置满足
	if this.cond != nil {
		logs.CtxDebug(ctx, "node wait", logs.String("node", this.taskName))
		this.cond.Wait()
	}

	err = this.safeExecute(ctx, param, taskContext)

	this.stopWatchIfNeed(ctx, start, err, true)

	//通知后置
	for waiter := range this.gi.g.reverseDepMap[this.taskName] {
		waiterNode := this.gi.nodeMap[waiter]
		if waiterNode == nil || waiterNode.cond == nil {
			continue
		}
		waiterNode.cond.Notify()
	}

	return err

}

func (this *node) safeExecute(ctx context.Context, param interface{}, taskContext interface{}) (err error) {
	if this.gi.earlyReturn {
		return nil
	}
	logs.CtxDebug(ctx, "execute", logs.String(fmt.Sprintf("task"), this.taskName))

	defer func() {
		if e := recover(); e != nil {
			logs.CtxDebug(ctx, "execute", logs.String(fmt.Sprintf("task"), this.taskName), logs.String("err", el_utils.ToJsonString(e)))
			err = fmt.Errorf("%v", e)
			return
		}
	}()

	if !this.shouldDo(ctx, param, taskContext) {
		return nil
	}

	var start time.Time
	if this.needStopWatch() {
		start = time.Now()
	}

	val, _ := taskMap.Load(this.taskName)
	iTask := val.(ITask)
	err = iTask.Load(ctx, param, taskContext)

	this.stopWatchIfNeed(ctx, start, err, false)

	return err
}

func (this *node) shouldDo(ctx context.Context, param interface{}, taskContext interface{}) bool {
	//递归检查依赖的节点是否被skip
	if this.skipByParent() {
		return false
	}

	val, _ := taskMap.Load(this.taskName)
	iTask := val.(ITask)
	if !iTask.ShouldDo(ctx, param, taskContext) {
		atomic.StoreInt32(&this.shouldSkip, 1)
		return false
	}

	return true
}

func (this *node) skipByParent() bool {
	depMap := this.gi.g.depMap[this.taskName]
	if len(depMap) == 0 {
		return atomic.LoadInt32(&this.shouldSkip) == 1
	}

	for parent, ok := range depMap {
		if !ok {
			continue
		}

		if this.gi.nodeMap[parent].skipByParent() {
			return true
		}
	}

	return false
}
