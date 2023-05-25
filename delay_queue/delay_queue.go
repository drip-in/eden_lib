package delay_queue

import (
	"context"
	"fmt"
	"github.com/drip-in/eden_lib/el_utils"
	"github.com/drip-in/eden_lib/godash/maps"
	"github.com/drip-in/eden_lib/logs"
	"github.com/go-redis/redis"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type PersistFn func(event *EventEntity) error

var (
	DelayQueueImpl *DelayQueue
)

type DelayQueue struct {
	namespace   string
	redisClient *redis.Client
	once        sync.Once
	wg          sync.WaitGroup
	isRunning   int32
	stop        chan struct{}
	persistFn   PersistFn
}

func init() {
	DelayQueueImpl = &DelayQueue{}
}

func GetDelayQueue() *DelayQueue {
	return DelayQueueImpl
}

func NewDelayQueue(namespace string, redisClient *redis.Client) *DelayQueue {
	DelayQueueImpl.namespace = namespace
	DelayQueueImpl.redisClient = redisClient
	DelayQueueImpl.stop = make(chan struct{})
	return DelayQueueImpl
}

func (q *DelayQueue) WithPersistForUnhandledEvent(fn PersistFn) {
	q.persistFn = fn
}

// gracefully shudown
func (q *DelayQueue) ShutDown() {
	if !atomic.CompareAndSwapInt32(&q.isRunning, 1, 0) {
		return
	}
	close(q.stop)
	q.wg.Wait()
}

func (q *DelayQueue) genBucketKey(topic string) string {
	return fmt.Sprintf("BUCKET_%v_%v", q.namespace, topic)
}

func (q *DelayQueue) genPoolKey(topic string) string {
	return fmt.Sprintf("POOL_%v_%v", q.namespace, topic)
}

func (q *DelayQueue) genQueueKey(topic string) string {
	return fmt.Sprintf("QUEUE_%v_%v", q.namespace, topic)
}

func (q *DelayQueue) InitOnce(subscriber IEventSubscriber, others ...IEventSubscriber) {
	if !atomic.CompareAndSwapInt32(&q.isRunning, 0, 1) {
		return
	}

	list := append([]IEventSubscriber{subscriber}, others...)
	topicConsumerMap := make(map[string][]IEventSubscriber)
	for _, s := range list {
		topicConsumerMap[s.Topic()] = append(topicConsumerMap[s.Topic()], s)
	}
	topicList := maps.Keys(topicConsumerMap).([]string)
	q.once.Do(func() {
		for _, t := range topicList {
			topic := t
			// 定时topic扫描到期的事件
			el_utils.GoSafe(func(ctx context.Context) {
				for {
					if atomic.LoadInt32(&q.isRunning) == 0 {
						break
					}
					count, err := q.carryEventToQueue(topic)
					if err == nil && count == 0 {
						time.Sleep(100 * time.Second)
					}
				}
			})

			// 消费topic队列的事件
			el_utils.GoSafe(func(ctx context.Context) {
				_ = q.runConsumer(topic, topicConsumerMap[topic])
			})
		}
	})
}

// 扫描zset中到期的任务，添加到对应topic的待消费队列里，并从Bucket中删除已进入待消费队列的事件;
// 每次都取指定数量,防止消息突增
func (q *DelayQueue) carryEventToQueue(topic string) (int64, error) {
	script := redis.NewScript(`
	  local members = redis.call('ZRangeByScore', KEYS[1], '0', ARGV[1], 'limit', 0, 20)
	  if(next(members) ~= nil) then
		redis.call('ZRem', KEYS[1], 0, unpack(members, 1, #members))
		redis.call('RPush', KEYS[2], unpack(members, 1, #members))
	  end
      return #members
	  `)
	ctx := context.Background()
	delayKey := q.genBucketKey(topic)
	readyKey := q.genQueueKey(topic)
	res, err := script.Run(q.redisClient.WithContext(ctx), []string{delayKey, readyKey}, el_utils.ToString(time.Now().Unix())).Result()
	if err != nil {
		logs.CtxError(ctx, "[carryEventToQueue] script.Run", logs.String("err", err.Error()))
		return 0, err
	}
	return res.(int64), nil
}

func (q *DelayQueue) runConsumer(topic string, subscriberList []IEventSubscriber) error {
	for {
		if atomic.LoadInt32(&q.isRunning) == 0 {
			break
		}
		q.wg.Add(1)
		ctx := context.Background()
		kvPair, err := q.redisClient.WithContext(ctx).BLPop(60*time.Second, q.genQueueKey(topic)).Result()
		if err != nil {
			logs.CtxWarn(ctx, "[runConsumer] BLPop", logs.String("err", err.Error()))
			q.wg.Done()
			continue
		}
		if len(kvPair) < 2 {
			q.wg.Done()
			continue
		}

		eventId := kvPair[1]
		data, err := q.redisClient.WithContext(ctx).HGet(q.genPoolKey(topic), eventId).Result()
		if err != nil && err != redis.Nil {
			logs.CtxWarn(ctx, "[runConsumer] HGet", logs.String("err", err.Error()))
			if q.persistFn != nil {
				_ = q.persistFn(&EventEntity{
					EventId: el_utils.String2Int64(eventId),
					Topic:   topic,
				})
			}
			q.wg.Done()
			continue
		}
		event := &EventEntity{}
		_ = jsoniter.UnmarshalFromString(data, event)

		for _, s := range subscriberList {
			subscriber := s
			el_utils.GoSafeWithCtx(ctx, func(ctx context.Context) {
				el_utils.Retry(3, 0, func() (success bool) {
					err = subscriber.Handle(ctx, event)
					if err != nil {
						logs.CtxWarn(ctx, "[runConsumer] subscriber.Handle", logs.String("err", err.Error()))
						return false
					}
					return true
				})
			})
		}

		err = q.redisClient.WithContext(ctx).HDel(q.genPoolKey(topic), eventId).Err()
		if err != nil {
			logs.CtxWarn(ctx, "[runConsumer] HDel", logs.String("err", err.Error()))
		}
		q.wg.Done()
	}
	return nil
}

// todo 原子性保证
func (q *DelayQueue) PublishEvent(ctx context.Context, event *EventEntity) error {
	pipeline := q.redisClient.WithContext(ctx).Pipeline()
	defer pipeline.Close()

	pipeline.HSet(q.genPoolKey(event.Topic), strconv.FormatInt(event.EventId, 10), el_utils.ToJsonString(event))
	pipeline.ZAdd(q.genBucketKey(event.Topic), redis.Z{
		Member: strconv.FormatInt(event.EventId, 10),
		Score:  float64(event.EffectTime.Unix()),
	})
	_, err := pipeline.Exec()
	if err != nil {
		logs.CtxWarn(ctx, "pipeline.Exec", logs.String("err", err.Error()))
		return err
	}
	logs.CtxInfo(ctx, "publish event success", logs.String("event", el_utils.ToJsonString(event)))
	return nil
}

//-- keys: pendingKey, readyKey
//-- argv: currentTime
//local msgs = redis.call('ZRangeByScore', KEYS[1], '0', ARGV[1])  -- 从 pending key 中找出已到投递时间的消息
//if (#msgs == 0) then return end
//local args2 = {'LPush', KEYS[2]} -- 将他们放入 ready key 中
//for _,v in ipairs(msgs) do
//table.insert(args2, v)
//end
//redis.call(unpack(args2))
//redis.call('ZRemRangeByScore', KEYS[1], '0', ARGV[1])  -- 从 pending key 中删除已投递的消息
//————————————————
//版权声明：本文为CSDN博主「十一技术斩」的原创文章，遵循CC 4.0 BY-SA版权协议，转载请附上原文出处链接及本声明。
//原文链接：https://blog.csdn.net/uuqaz/article/details/125916298
