package pubsub

import (
	"context"
	"fmt"
	"github.com/drip-in/eden_lib/errcode"
	"github.com/drip-in/eden_lib/logs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

const Enabled = "enabled"

type KafkaConfig struct {
	SASLEnabled         string
	Brokers             string
	Key                 string
	Secret              string
	ProducerID          string
	ProducerName        string
	ConsumerServiceName string
}

func (config KafkaConfig) IsSASLEnabled() bool {
	return config.SASLEnabled == Enabled
}

func (config KafkaConfig) NewKafkaClientConfig() *sarama.Config {
	ver, err := sarama.ParseKafkaVersion("2.2.1")
	if err != nil {
		panic(fmt.Sprint("Error parsing Kafka version", err))
	}
	conf := sarama.NewConfig()
	conf.Consumer.MaxProcessingTime = time.Second * 30
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	conf.Version = ver
	conf.Net.SASL.User = config.Key
	conf.Net.SASL.Password = config.Secret
	if config.IsSASLEnabled() {
		conf.Net.SASL.Handshake = true
		conf.Net.SASL.Enable = true
		conf.Net.TLS.Enable = true
	}

	saslMechanism, ok := os.LookupEnv("KAFKA_SASL_MECHANISM")
	if !ok {
		saslMechanism = "PLAIN"
	}
	if saslMechanism != "PLAIN" {
		if saslMechanism == "SCRAM_SHA512" {
			conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
			}
			conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		} else if saslMechanism == "SCRAM_SHA256" {
			conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
			}
			conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		} else {
			panic(fmt.Sprint("Invalid or unsupported KAFKA_SASL_MECHANISM", saslMechanism, "only PLAIN (default) or SCRAM_SHA512 or SCRAM_SHA256 are supported"))
		}
	}
	return conf
}

type KafkaClient interface {
	AsyncPublish(topic string, key string, message proto.Message) error
	SyncPublish(topic string, key string, message proto.Message) (partition int32, offset int64, err error)
	Subscribe(topic string, handler sarama.ConsumerGroupHandler) error
	SubscribeWithGroup(topic string, groupID string, handler sarama.ConsumerGroupHandler) error
	Unsubscribe(topic string) error
	UnsubscribeWithGroup(topic string, groupID string) error
	RegisterAsyncProducerErrCallback(topic string, callback ProducerErrorCallback)
	GetConfig() KafkaConfig
}

type ProducerErrorCallback func(producerError *sarama.ProducerError)

type consumerGroupKey struct {
	topic   string
	groupID string
}

type KafkaClientImpl struct {
	admin               sarama.ClusterAdmin
	asyncProducer       sarama.AsyncProducer
	syncProducer        sarama.SyncProducer
	consumerGroups      map[consumerGroupKey]sarama.ConsumerGroup
	cancelMap           map[consumerGroupKey]context.CancelFunc
	producerErrCallback sync.Map
	config              KafkaConfig
	mu                  sync.Mutex
}

func (kafka *KafkaClientImpl) SubscribeWithGroup(topic string, groupID string, handler sarama.ConsumerGroupHandler) error {
	consumerConfig := kafka.config.NewKafkaClientConfig()
	consumerGroup, err := sarama.NewConsumerGroup(strings.Split(kafka.config.Brokers, ","), groupID, consumerConfig)
	if err != nil {
		return err
	}
	key := consumerGroupKey{topic: topic, groupID: groupID}
	kafka.mu.Lock()
	defer kafka.mu.Unlock()
	kafka.consumerGroups[key] = consumerGroup
	go func() {
		ctx, cancelFunc := context.WithCancel(context.Background())
		kafka.cancelMap[key] = cancelFunc
		for {
			if err = kafka.consumerGroups[key].Consume(ctx, []string{topic}, handler); err != nil {
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	return nil
}

func (kafka *KafkaClientImpl) Subscribe(topic string, handler sarama.ConsumerGroupHandler) error {
	if kafka.config.ConsumerServiceName == "" {
		panic(fmt.Sprint("Kafka consumer service name is empty"))
	}
	groupID := fmt.Sprintf("%s-%s-consumer-group", kafka.config.ConsumerServiceName, topic)
	return kafka.SubscribeWithGroup(topic, groupID, handler)
}

func (kafka *KafkaClientImpl) Unsubscribe(topic string) error {
	if kafka.config.ConsumerServiceName == "" {
		panic(fmt.Sprint("Kafka consumer service name is empty"))
	}
	groupID := fmt.Sprintf("%s-%s-consumer-group", kafka.config.ConsumerServiceName, topic)
	return kafka.UnsubscribeWithGroup(topic, groupID)
}

func (kafka *KafkaClientImpl) UnsubscribeWithGroup(topic string, groupID string) error {
	key := consumerGroupKey{topic: topic, groupID: groupID}

	kafka.mu.Lock()
	defer kafka.mu.Unlock()

	if group, ok := kafka.consumerGroups[key]; ok {
		if err := group.Close(); err != nil {
			return err
		}
	}
	delete(kafka.consumerGroups, key)

	if cancelFunc, ok := kafka.cancelMap[key]; ok {
		cancelFunc()
		delete(kafka.cancelMap, key)
	}
	return nil
}

func (kafka *KafkaClientImpl) AsyncPublish(topic string, key string, message proto.Message) error {
	if kafka.asyncProducer == nil {
		return errcode.KafkaProducerInitFailed
	}
	bytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	producerMessage := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(bytes),
		Timestamp: time.Now(),
	}
	select {
	case kafka.asyncProducer.Input() <- producerMessage:
		return nil
	case <-time.After(time.Second * 3):
		return errcode.KafkaPublishTimeout
	}
}

func (kafka *KafkaClientImpl) SyncPublish(topic string, key string, message proto.Message) (partition int32, offset int64, err error) {
	if kafka.syncProducer == nil {
		err = errcode.KafkaProducerInitFailed
		return
	}
	bytes, err := proto.Marshal(message)
	if err != nil {
		err = errcode.ErrInvalidParam
		return
	}
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(bytes),
		Timestamp: time.Now(),
	}
	partition, offset, err = kafka.syncProducer.SendMessage(msg)
	return
}

func (kafka *KafkaClientImpl) RegisterAsyncProducerErrCallback(topic string, callback ProducerErrorCallback) {
	kafka.producerErrCallback.Store(topic, callback)
}

func (kafka *KafkaClientImpl) GetConfig() KafkaConfig {
	return kafka.config
}

func NewKafkaClient(config KafkaConfig) (KafkaClient, error) {
	brokers := strings.Split(config.Brokers, ",")
	kafkaConfig := config.NewKafkaClientConfig()
	admin, err := sarama.NewClusterAdmin(brokers, kafkaConfig)
	if err != nil {
		logs.CtxError(context.Background(), "create_kafka_admin_failed", logs.String("err", err.Error()))
		return nil, err
	}
	producer, err := sarama.NewAsyncProducer(brokers, kafkaConfig)
	if err != nil {
		logs.CtxError(context.Background(), "create_kafka_async_producer_failed", logs.String("err", err.Error()))
		return nil, err
	}

	syncProducer, err := sarama.NewSyncProducer(brokers, kafkaConfig)
	if err != nil {
		logs.CtxError(context.Background(), "create_kafka_sync_producer_failed", logs.String("err", err.Error()))
		return nil, err
	}

	client := &KafkaClientImpl{
		config:              config,
		asyncProducer:       producer,
		syncProducer:        syncProducer,
		admin:               admin,
		consumerGroups:      make(map[consumerGroupKey]sarama.ConsumerGroup),
		cancelMap:           make(map[consumerGroupKey]context.CancelFunc),
		producerErrCallback: sync.Map{},
	}

	go func() {
		for e := range producer.Errors() {
			keyBytes, _ := e.Msg.Key.Encode()
			key := string(keyBytes)
			valBytes, _ := e.Msg.Value.Encode()
			logs.CtxError(context.Background(), "kafka_publish_failed", logs.String("key", key), logs.String("val", string(valBytes)), logs.String("topic", e.Msg.Topic), logs.String("err", e.Err.Error()))

			if v, ok := client.producerErrCallback.Load(e.Msg.Topic); ok {
				v.(ProducerErrorCallback)(e)
			}
		}
	}()

	return client, nil
}
