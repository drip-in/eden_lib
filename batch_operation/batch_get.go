package batch_operation

import (
	"context"
	"github.com/drip-in/eden_lib/el_utils"
	"github.com/drip-in/eden_lib/logs"
	"go.uber.org/atomic"
	"math"
	"sync"
	"time"
)

type ConcurrentHandlerInBatch struct {
	BatchHandler   BatchHandler
	TimeOut        time.Duration
	BatchSize      int
	MaxConcurrency int
	ExitWhenError  bool
}

type BatchHandler func(ctx context.Context, batchInParams []interface{}, dataMap *sync.Map) error

// ConRun 并发运行
func (c *ConcurrentHandlerInBatch) ConRun(ctx context.Context, inParams []interface{}) (*sync.Map, error) {
	batchSize := c.GetBatchSize()
	var ch chan struct{}
	if c.MaxConcurrency > 0 {
		ch = make(chan struct{}, c.MaxConcurrency)
	}
	count := int(math.Ceil(float64(len(inParams)) / float64(batchSize)))
	group := &sync.WaitGroup{}
	dataMap := &sync.Map{}
	errMap := &sync.Map{}
	exitSignal := atomic.NewBool(false)
	for i := 0; i < count; i++ {
		if c.ExitWhenError && exitSignal.Load() {
			break
		}
		start, end := i*batchSize, batchSize*(i+1)
		if end > len(inParams) {
			end = len(inParams)
		}
		group.Add(1)
		if c.MaxConcurrency > 0 {
			ch <- struct{}{}
		}
		go func(start, end int) {
			defer func() {
				if c.MaxConcurrency > 0 {
					<-ch
				}
				group.Done()
				if r := recover(); r != nil {
					logs.CtxError(ctx, "ConcurrentGetPoiData panic")
				}
			}()
			if c.ExitWhenError && exitSignal.Load() {
				return
			}
			err := c.handlerWithRetry(ctx, inParams[start:end], dataMap)
			if err != nil {
				errMap.Store(start, err)
				if c.ExitWhenError {
					exitSignal.Store(true)
				}
				return
			}
		}(start, end)
	}
	group.Wait()
	var err error
	errMap.Range(func(key, value interface{}) bool {
		err = value.(error)
		return false
	})
	return dataMap, err
}

// GetTimeOut 返回超时时间
func (c *ConcurrentHandlerInBatch) GetTimeOut() time.Duration {
	if c.TimeOut != 0 {
		return c.TimeOut
	}
	return TimeOut
}

// GetBatchSize 返回批量大小
func (c *ConcurrentHandlerInBatch) GetBatchSize() int {
	if c.BatchSize > 0 {
		return c.BatchSize
	}
	return 50
}

const (
	TimeOut = 5 * time.Second
)

func (c *ConcurrentHandlerInBatch) handlerWithRetry(ctx context.Context, batchInParams []interface{}, dataMap *sync.Map) error {
	var err error
	success := el_utils.Retry(3, 0, func() (success bool) {
		err = c.BatchHandler(ctx, batchInParams, dataMap)
		if err != nil {
			logs.CtxError(ctx, "BatchHandler fail", logs.String("batchInParams", el_utils.ToJsonString(batchInParams)), logs.String("err", err.Error()))
			return false
		}
		return true
	})
	if !success {
		return err
	}
	return nil
}
