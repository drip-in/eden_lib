package delay_queue

import (
	"context"
	"fmt"
	"github.com/drip-in/eden_lib/el_utils"
	"github.com/drip-in/eden_lib/logs"
	"github.com/go-redis/redis"
	"strconv"
	"sync"
	"time"
)

type DelayQueue struct {
	namespace   string
	redisClient *redis.Client
	once        sync.Once
}

func NewDelayQueue(namespace string, redisClient *redis.Client) *DelayQueue {
	initQueue()
	return &DelayQueue{
		namespace:   namespace,
		redisClient: redisClient,
	}
}

func initQueue() {

}

func (q *DelayQueue) InitOnce(subscriber IEventSubscriber, others ...IEventSubscriber) {
	list := append([]IEventSubscriber{subscriber}, others...)
	q.once.Do(func() {
		for _, s := range list {
			subscriber := s
			el_utils.GoSafe(func() {
				ticker := time.NewTicker(time.Second)
				for {
					select {
					case <-ticker.C:
						_ = q.consumeWithSubscriber(subscriber)
						return
					}
				}
			})
		}
	})
}

func (q *DelayQueue) consumeWithSubscriber(subscriber IEventSubscriber) error {
	ctx := context.Background()
	members, err := q.redisClient.WithContext(ctx).ZRangeWithScores(q.genBucketKey(subscriber.Topic()), 0, time.Now().Unix()).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	group := &sync.WaitGroup{}
	lock := &sync.Mutex{}
	errMap := make(map[string]error)
	for _, m := range members {
		group.Add(1)
		eventId := m.Member.(string)
		data, err := q.redisClient.WithContext(ctx).HGet(q.genPoolKey(subscriber.Topic()), eventId).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		event := &EventEntity{
			EventId: el_utils.String2Int64(eventId),
			Body:    data,
		}
		el_utils.GoSafe(func() {
			err = subscriber.Handle(ctx, event)
			if err != nil {
				lock.Lock()
				errMap[eventId] = err
				lock.Unlock()
			}
		}, group.Done)
	}
	group.Wait()

	// 从JobPool和Bucket中删除事件
	var fields []string
	var doneMembers []interface{}
	for _, m := range members {
		eventId := m.Member.(string)
		if _, ok := errMap[eventId]; !ok {
			fields = append(fields, eventId)
			doneMembers = append(doneMembers, eventId)
		}
	}
	pipeline := q.redisClient.WithContext(ctx).Pipeline()
	defer pipeline.Close()

	pipeline.HDel(q.genPoolKey(subscriber.Topic()), fields...)
	pipeline.ZRem(q.genBucketKey(subscriber.Topic()), doneMembers...)
	_, err = pipeline.Exec()
	if err != nil {
		//重试一次
		_, err = pipeline.Exec()
	}
	return err
}

func (q *DelayQueue) genBucketKey(topic string) string {
	return fmt.Sprintf("BUCKET_%v_%v", q.namespace, topic)
}

func (q *DelayQueue) genPoolKey(topic string) string {
	return fmt.Sprintf("POOL_%v_%v", q.namespace, topic)
}

func (q *DelayQueue) PublishEvent(ctx context.Context, event EventEntity) error {
	pipeline := q.redisClient.WithContext(ctx).Pipeline()
	defer pipeline.Close()

	pipeline.HSet(q.genPoolKey(event.Topic), strconv.FormatInt(event.EventId, 10), event.Body)
	pipeline.ZAdd(q.genBucketKey(event.Topic), redis.Z{
		Member: strconv.FormatInt(event.EventId, 10),
		Score:  float64(event.EffectTime.Unix()),
	})
	_, err := pipeline.Exec()
	if err != nil { //报错后进行一次额外尝试
		logs.CtxWarn(ctx, "pipeline.Exec", logs.String("err", err.Error()))
		_, err = pipeline.Exec()
	}
	return err
}
