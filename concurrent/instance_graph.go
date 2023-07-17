package concurrent

import (
	"context"
	"github.com/drip-in/eden_lib/logs"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

var graphPool sync.Pool

func init() {
	graphPool.New = newGraphInstance
}

type graphInstance struct {
	nodeMap map[string]*node
	g       *graph
	counter int32
	signal  chan struct{}

	name        string
	earlyReturn bool
}

func newGraphInstance() interface{} {
	return &graphInstance{}
}

func (this *graphInstance) zero() {
	this.nodeMap = nil
	this.g = nil
	this.counter = 0
	this.signal = nil
	this.name = ""
	this.earlyReturn = false
}

func (this *graphInstance) recycle() {
	for _, n := range this.nodeMap {
		n.recycle()
	}

	this.zero()
	graphPool.Put(this)
}

func (this *graphInstance) Execute(ctx context.Context, param interface{}, taskContext interface{}) error {
	this.plan()

	var e error

	for _, n := range this.nodeMap {
		taskName := n.taskName
		n := n
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logs.CtxWarn(ctx, "", logs.String("graph", this.name), logs.String("execute node", taskName), logs.ByteString("stack", debug.Stack()))
				}
			}()
			err := n.execute(ctx, param, taskContext)
			if err != nil {
				if n.core {
					e = err
					this.earlyReturn = true
					this.signal <- struct{}{}
				}
			}
		}()
	}

	<-this.signal

	if this.earlyReturn {
		for _, n := range this.nodeMap {
			if n.cond != nil {
				n.cond.Notify()
			}
		}
		return e
	}

	//加载loader
	for _, p := range this.g.packers {
		err := p.Pack(ctx, param, taskContext)
		if err != nil {
			return err
		}
	}
	return e
}

// 构建cond
func (this *graphInstance) plan() {
	/*	for taskName := range this.g.taskSet {
		val, _ := taskMap.Load(taskName)
		if !val.(ITask).ShouldDo(ctx, param) {
			this.deleteNode(taskName)
		}
	}*/

	for taskName, node := range this.nodeMap {
		atomic.AddInt32(&this.counter, 1)
		//当前节点有依赖，则cond需要所有依赖通知
		if depList, ok := this.g.depMap[taskName]; ok {
			node.cond = NewCond(int32(len(depList)))
		}
	}
}

/*func (this *graphInstance) deleteNode(taskName string) {
	// 删除节点
	if this.nodeMap[taskName] == nil {
		return
	}
	this.nodeMap[taskName].recycle()
	delete(this.nodeMap, taskName)

	// 被删节点的所依赖关系删掉
	delete(this.depMap, taskName)

	for main, subList := range this.reverseDepMap {
		for sub := range subList {
			if sub == taskName {
				// 被依赖关系中，删除taskName
				delete(this.reverseDepMap[main], taskName)
			}
		}
	}

	// 依赖被删节点的节点，也都删掉
	for sub := range this.reverseDepMap[taskName] {
		this.deleteNode(sub)
	}
}*/
