package concurrent

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNode(t *testing.T) {

	g := NewGraph("TestNode").
		AddNode(taskA, true).
		AddNode(taskE, true).
		AddNode(taskB, true, taskA).
		AddNode(taskC, true, taskA).
		AddNode(taskD, true, taskB, taskC).
		AddNode(taskF, true, taskE).
		AddNode(taskG, true, taskD, taskF).
		AddNode(taskH, true, taskD).
		AddPacker(1, &Pack1{}).
		AddPacker(3, &Pack2{}).
		AddPacker(2, &Pack3{}).
		Build()

	entry := &TaskContext{}
	err := g.Execute(context.Background(), 2, entry)
	if err != nil {
		t.Logf("err:%s\n", err)
	}

	t.Logf("entry:%s\n", entry)
}

func TestCheck(t *testing.T) {
	g := NewGraph("TestCheck").
		AddNode(taskA, true).
		AddNode(taskB, true, taskA).
		AddNode(taskC, true).
		AddNode(taskD, true, taskC).
		AddNode(taskF, true, taskE)

	g.Build()
}

func TestCheckCycle(t *testing.T) {
	g := NewGraph("TestCheckCycle").
		AddNode(taskA, true, taskH).
		AddNode(taskE, true).
		AddNode(taskB, true, taskA).
		AddNode(taskC, true, taskA).
		AddNode(taskD, true, taskB, taskC).
		AddNode(taskF, true, taskE).
		AddNode(taskG, true, taskD, taskF).
		AddNode(taskH, true, taskD)

	g.Build()
}

func TestLayer(t *testing.T) {
	g := NewGraph("TestNode").
		AddNode(taskA, true).
		AddNode(taskB, true, taskA).
		AddNode(taskC, true).
		AddNode(taskD, true, taskC).
		AddNode(taskE, true).
		AddNode(taskF, true, taskE).LoggerTaskCost().Build()

	entry := &TaskContext{}
	now := time.Now()
	err := g.Execute(context.Background(), 2, entry)
	if err != nil {
		panic(err)
	}

	cost := time.Now().Sub(now).Milliseconds()
	t.Logf("entry:%s\n", entry)
	t.Logf("cost:%v Ms\n", cost)
}

func TestConcurrentNode(t *testing.T) {

	g := NewGraph("TestConcurrentNode").
		AddNode(taskA, true).
		AddNode(taskE, true).
		AddNode(taskB, true, taskA).
		AddNode(taskC, true, taskA).
		AddNode(taskD, true, taskB, taskC).
		AddNode(taskF, true, taskE).
		AddNode(taskG, true, taskD, taskF).
		AddNode(taskH, true, taskD).
		Build()

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			entry := &TaskContext{}
			err := g.Execute(context.Background(), i, entry)
			if err != nil {
				t.Logf("err:%s\n", err)
			}
			t.Logf("[%v] entry:%s\n", i, entry)

		}(i)
	}

	wg.Wait()
}

func TestSingleNode(t *testing.T) {
	g := NewGraph("TestSingleNode").
		AddNode(taskB, true).
		Build()

	entry := &TaskContext{}
	now := time.Now()
	err := g.Execute(context.Background(), 1, entry)
	if err != nil {
		panic(err)
	}

	cost := time.Now().Sub(now).Milliseconds()
	t.Logf("entry:%s\n", entry)
	t.Logf("cost:%v Ms\n", cost)
}

func TestBts(t *testing.T) {
	//bts := []byte{38, 123, 91, 55, 48, 52, 50, 50, 51, 55, 52, 50, 50, 50, 56, 49, 49, 54, 52, 56, 48, 55, 93, 32, 91, 93, 32, 91, 93, 125}
	//t.Log(string(bts))

	location, err := time.ParseInLocation("2006-01-02 15:04:05", "0000-00-00 00:00:00", time.Local)
	if err != nil {
		panic(err)
	}

	t.Log(location)

}
