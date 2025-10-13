package pubsub

import (
	"context"
	"errors"
	"fmt"
	"github.com/drip-in/eden_lib/logs"
	"runtime/debug"
	"sync"

	"github.com/IBM/sarama"
)

type KafkaMessageHandler func(key, value []byte) error
type KafkaErrorHandler func(key, value []byte, err error) error

type KafkaMessageDispatcher interface {
	RegisterHandler(topic string, handler KafkaMessageHandler, fallbacks ...KafkaErrorHandler)
	RegisterHandlerWithGroup(topic string, groupID string, handler KafkaMessageHandler, fallbacks ...KafkaErrorHandler)
}

type handlers struct {
	handler   KafkaMessageHandler
	fallbacks []KafkaErrorHandler
}

// 为每个消费者组创建独立的处理器
type groupHandler struct {
	dispatcher *kafkaMessageDispatcher
	groupID    string
	topic      string
}

type kafkaMessageDispatcher struct {
	kafkaClient KafkaClient
	// key: topic-groupID, value: handlers
	handlers map[string]handlers
	// 保存每个 topic 对应的所有 groupIDs
	topicGroups map[string][]string
	mu          sync.RWMutex
}

func (dispatcher *kafkaMessageDispatcher) RegisterHandler(topic string, handler KafkaMessageHandler, fallbacks ...KafkaErrorHandler) {
	// 生成默认的消费者组ID
	groupID := fmt.Sprintf("%s-%s-consumer-group", dispatcher.kafkaClient.GetConfig().ConsumerServiceName, topic)
	dispatcher.registerHandlerWithGroup(topic, groupID, handler, fallbacks...)
}

func (dispatcher *kafkaMessageDispatcher) RegisterHandlerWithGroup(topic string, groupID string, handler KafkaMessageHandler, fallbacks ...KafkaErrorHandler) {
	dispatcher.registerHandlerWithGroup(topic, groupID, handler, fallbacks...)
}

// 内部注册方法
func (dispatcher *kafkaMessageDispatcher) registerHandlerWithGroup(topic string, groupID string, handler KafkaMessageHandler, fallbacks ...KafkaErrorHandler) {
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()

	// 创建 topic-groupID 组合的 key
	key := fmt.Sprintf("%s-%s", topic, groupID)

	// 保存处理器
	dispatcher.handlers[key] = handlers{
		handler:   handler,
		fallbacks: fallbacks,
	}

	// 更新 topic 对应的 groupIDs 列表
	groupExists := false
	for _, existingGroup := range dispatcher.topicGroups[topic] {
		if existingGroup == groupID {
			groupExists = true
			break
		}
	}

	if !groupExists {
		dispatcher.topicGroups[topic] = append(dispatcher.topicGroups[topic], groupID)
		// 为这个消费者组创建独立的处理器并订阅
		groupHandler := &groupHandler{
			dispatcher: dispatcher,
			groupID:    groupID,
			topic:      topic,
		}
		err := dispatcher.kafkaClient.SubscribeWithGroup(topic, groupID, groupHandler)
		if err != nil {
			logs.CtxError(context.TODO(), "kafka_failed_to_register_handler", logs.String("topic", topic), logs.String("groupID", groupID), logs.String("err", err.Error()))
		}
	}
}

// groupHandler 实现 sarama.ConsumerGroupHandler 接口
func (gh *groupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (gh *groupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (gh *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			// 直接使用已知的 groupID 处理消息
			err := gh.dispatcher.handleMessageWithGroup(gh.groupID, message, nil)
			if err != nil {
				err = gh.dispatcher.handleMessageWithGroup(gh.groupID, message, err)
				if err != nil {
					logs.CtxError(session.Context(), "kafka_failed_to_handle_error", logs.String("key", string(message.Key)), logs.String("err", err.Error()))
				}
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

// 使用指定 groupID 处理消息
func (dispatcher *kafkaMessageDispatcher) handleMessageWithGroup(groupID string, message *sarama.ConsumerMessage, err error) (returnErr error) {
	if message == nil {
		return errors.New("message is null")
	}

	defer func() {
		if p := recover(); p != nil {
			logs.CtxError(context.Background(), "kafka_consumer_panic", logs.String("stack", string(debug.Stack())), logs.String("err", fmt.Sprint(p)))
			returnErr = errors.New(fmt.Sprint(p))
		}
	}()

	// 构造 key 查找对应的处理器
	key := fmt.Sprintf("%s-%s", message.Topic, groupID)

	handlers, ok := dispatcher.handlers[key]

	if ok {
		if err != nil {
			// 处理错误情况
			for index := range handlers.fallbacks {
				err := handlers.fallbacks[index](message.Key, message.Value, err)
				if err != nil {
					return err
				}
			}
			return nil
		}
		// 正常处理消息
		return handlers.handler(message.Key, message.Value)
	} else {
		return fmt.Errorf("no consumer for topic %s with group %s", message.Topic, groupID)
	}
}

func NewKafkaMessageDispatcher(kafkaClient KafkaClient) KafkaMessageDispatcher {
	return &kafkaMessageDispatcher{
		kafkaClient: kafkaClient,
		handlers:    make(map[string]handlers),
		topicGroups: make(map[string][]string),
	}
}

// 确保接口实现
var _ KafkaMessageDispatcher = &kafkaMessageDispatcher{}
