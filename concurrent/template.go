package concurrent

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/drip-in/eden_lib/logs"
)

// graph
type graph struct {
	name string

	taskSet       map[string]bool
	depMap        map[string]map[string]bool
	reverseDepMap map[string]map[string]bool
	packers       []*packComponent

	checkPass       bool
	metricsTaskCost bool
	loggerTaskCost  bool
}

func NewGraph(name string) *graph {
	return &graph{
		name:          name,
		taskSet:       make(map[string]bool),
		depMap:        make(map[string]map[string]bool),
		reverseDepMap: make(map[string]map[string]bool),
	}
}

func (this *graph) AddNode(task ITask, isCore bool, dependsOn ...ITask) *graph {
	if this.checkPass {
		panic("graph is checked")
	}

	taskName := task.Name()
	if _, ok := this.taskSet[taskName]; ok {
		panic(fmt.Errorf("node:%v already exist", taskName))
	}

	registerTask(task)

	for _, dep := range dependsOn {
		registerTask(dep)
		this.addDep(taskName, dep.Name())
	}

	this.taskSet[taskName] = isCore
	return this
}

func (this *graph) addDep(sub string, main string) {
	// sub依赖main
	if _, ok := this.depMap[sub]; !ok {
		this.depMap[sub] = make(map[string]bool)
	}
	this.depMap[sub][main] = true

	// main被sub依赖
	if _, ok := this.reverseDepMap[main]; !ok {
		this.reverseDepMap[main] = make(map[string]bool)
	}
	this.reverseDepMap[main][sub] = true
}

func (this *graph) AddPacker(prior int, packer IPack) *graph {
	if this.checkPass {
		panic("graph is checked")
	}

	this.packers = append(this.packers, &packComponent{
		IPack: packer,
		prior: prior,
	})

	return this
}

func (this *graph) MetricsTaskCost() *graph {
	if this.checkPass {
		panic("graph is checked")
	}

	this.metricsTaskCost = true
	return this
}

func (this *graph) LoggerTaskCost() *graph {
	if this.checkPass {
		panic("graph is checked")
	}

	this.loggerTaskCost = true
	return this
}

func (this *graph) Build() *graph {
	this.checkDeclare()
	// 验证是否有环形依赖
	for n := range this.depMap {
		this.dfs(n, make(map[string]struct{}))
	}

	this.checkPass = true
	return this
}

func (this *graph) checkDeclare() {
	if len(this.packers) != 0 {
		sort.Slice(this.packers, func(i, j int) bool {
			return this.packers[i].prior < this.packers[j].prior
		})
	}

	for _, vList := range this.depMap {
		for v := range vList {
			if _, ok := this.taskSet[v]; !ok {
				panic(fmt.Sprintf("dependency:%v not declared in graph", v))
			}
		}
	}

}

func (this *graph) dfs(name string, visited map[string]struct{}) {
	if _, ok := visited[name]; ok {
		panic(fmt.Sprintf("graph:%v cycle detected", this.name))
	}
	visited[name] = struct{}{}
	for next := range this.depMap[name] {
		visitedSnapshot := make(map[string]struct{})
		for k, v := range visited {
			visitedSnapshot[k] = v
		}

		this.dfs(next, visitedSnapshot)
	}
}

func (this *graph) Execute(ctx context.Context, param interface{}, taskContext interface{}) error {
	if !this.checkPass {
		return errors.New("graph not checked")
	}

	now := time.Now()
	instance := this.newInstance()
	err := instance.Execute(ctx, param, taskContext)
	if !instance.earlyReturn {
		instance.recycle()
	}

	costMs := time.Since(now).Milliseconds()

	logs.CtxDebug(ctx, "graph", logs.String("name", this.name), logs.String("cost", fmt.Sprintf("%v ms", costMs)))

	return err
}

func (this *graph) newInstance() *graphInstance {
	instance := graphPool.Get().(*graphInstance)
	instance.nodeMap = make(map[string]*node)
	instance.g = this
	instance.signal = make(chan struct{}, 2)
	instance.name = this.name

	// copy node
	for taskName, isCore := range this.taskSet {
		n := nodePool.Get().(*node)
		n.gi = instance
		n.taskName = taskName
		n.graphName = this.name
		n.core = isCore
		n.metricsTaskCost = this.metricsTaskCost
		n.loggerTaskCost = this.loggerTaskCost

		instance.nodeMap[taskName] = n
	}

	return instance
}
