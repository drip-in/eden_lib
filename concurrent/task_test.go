package concurrent

import (
	"context"
	"fmt"
	"time"
)

type TaskContext struct {
	a string
	b string
	c string
	d string
	e string
	f string
	g string
	h string

	res1 string
	res2 string
	res3 string
}

func (tc *TaskContext) String() string {
	return fmt.Sprintf("{\na:(%v)\nb:(%v)\nc:(%v)\nd:(%v)\ne:(%v)\nf:(%v)\ng:(%v)\nh:(%v)\nres1:(%v)\nres2:(%v)\nres3:(%v)\n}",
		tc.a, tc.b, tc.c, tc.d, tc.e, tc.f, tc.g, tc.h, tc.res1, tc.res2, tc.res3)
}

var (
	taskA *TaskA
	taskB *TaskB
	taskC *TaskC
	taskD *TaskD
	taskE *TaskE
	taskF *TaskF
	taskG *TaskG
	taskH *TaskH
)

func init() {
	taskA = &TaskA{}
	taskB = &TaskB{}
	taskC = &TaskC{}
	taskD = &TaskD{}
	taskE = &TaskE{}
	taskF = &TaskF{}
	taskG = &TaskG{}
	taskH = &TaskH{}

}

// a
type TaskA struct {
}

func (this *TaskA) Name() string {
	return "a"
}

func (this *TaskA) ShouldDo(ctx context.Context, param, taskContext interface{}) bool {
	return true
}

func (this *TaskA) Load(ctx context.Context, dataCtx, taskContext interface{}) error {
	time.Sleep(time.Second)
	taskContext.(*TaskContext).a = "a"
	//return fmt.Errorf("fuck")
	//panic("fuck")
	return nil
}

// b
type TaskB struct {
}

func (this *TaskB) Name() string {
	return "b"
}

func (this *TaskB) ShouldDo(ctx context.Context, param, taskContext interface{}) bool {
	return true
}

func (this *TaskB) Load(ctx context.Context, dataCtx, taskContext interface{}) error {
	taskContext.(*TaskContext).b = "b"
	time.Sleep(time.Second)
	return nil
}

// c
type TaskC struct {
}

func (this *TaskC) Name() string {
	return "c"
}

func (this *TaskC) ShouldDo(ctx context.Context, param, taskContext interface{}) bool {
	return true
}

func (this *TaskC) Load(ctx context.Context, dataCtx, taskContext interface{}) error {
	taskContext.(*TaskContext).c = "c"
	time.Sleep(time.Second)
	return nil
}

// d
type TaskD struct {
}

func (this *TaskD) Name() string {
	return "d"
}

func (this *TaskD) ShouldDo(ctx context.Context, param, taskContext interface{}) bool {
	return param.(int)%2 == 0
}

func (this *TaskD) Load(ctx context.Context, dataCtx, taskContext interface{}) error {
	taskContext.(*TaskContext).d = "d"
	time.Sleep(time.Second)
	return nil
}

// e
type TaskE struct {
}

func (this *TaskE) Name() string {
	return "e"
}

func (this *TaskE) ShouldDo(ctx context.Context, param, taskContext interface{}) bool {
	return true
}

func (this *TaskE) Load(ctx context.Context, dataCtx, taskContext interface{}) error {
	taskContext.(*TaskContext).e = "e"
	time.Sleep(time.Second)
	//panic("fuck")
	return nil
}

// f
type TaskF struct {
}

func (this *TaskF) Name() string {
	return "f"
}

func (this *TaskF) ShouldDo(ctx context.Context, param, taskContext interface{}) bool {
	return true
}

func (this *TaskF) Load(ctx context.Context, dataCtx, taskContext interface{}) error {
	taskContext.(*TaskContext).f = "f"
	time.Sleep(time.Second)
	return nil
}

// g
type TaskG struct {
}

func (this *TaskG) Name() string {
	return "g"
}

func (this *TaskG) ShouldDo(ctx context.Context, param, taskContext interface{}) bool {
	return true
}

func (this *TaskG) Load(ctx context.Context, dataCtx, taskContext interface{}) error {
	taskContext.(*TaskContext).g = "g"
	time.Sleep(time.Second)
	return nil
}

// h
type TaskH struct {
}

func (this *TaskH) Name() string {
	return "h"
}

func (this *TaskH) ShouldDo(ctx context.Context, param, taskContext interface{}) bool {
	return true
}

func (this *TaskH) Load(ctx context.Context, dataCtx, taskContext interface{}) error {
	taskContext.(*TaskContext).h = "h"
	time.Sleep(time.Second)
	return nil
}
