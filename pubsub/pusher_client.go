package pubsub

import (
	"context"
	"github.com/drip-in/eden_lib/logs"
	"net/http"

	"github.com/pusher/pusher-http-go/v5"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	PusherTracerName = "pusher"
)

type PusherConfig struct {
	AppID   string `mapstructure:"app_id"`
	Key     string `mapstructure:"key"`
	Secret  string `mapstructure:"secret"`
	Cluster string `mapstructure:"cluster"`
	Secure  bool   `mapstructure:"secure"`
}

type PusherClientWithCtx func(ctx context.Context) PusherClient

type PusherClient interface {
	Trigger(channel, eventName string, data interface{}) error
	AuthenticatePrivateChannel(params []byte) (response []byte, err error)
	Webhook(header http.Header, body []byte) (*pusher.Webhook, error)
}

type PusherClientImpl struct {
	ctx    context.Context
	client pusher.Client
}

func (p *PusherClientImpl) AuthenticatePrivateChannel(params []byte) (response []byte, err error) {
	ddSpan, _ := tracer.StartSpanFromContext(p.ctx, PusherTracerName)
	response, err = p.client.AuthorizePrivateChannel(params)
	ddSpan.Finish(tracer.WithError(err))
	return
}

func (p *PusherClientImpl) Webhook(header http.Header, body []byte) (webhook *pusher.Webhook, err error) {
	ddSpan, _ := tracer.StartSpanFromContext(p.ctx, PusherTracerName)
	webhook, err = p.client.Webhook(header, body)
	ddSpan.Finish(tracer.WithError(err))
	return
}

func (p *PusherClientImpl) Trigger(channel string, eventName string, data interface{}) (err error) {
	requestCtx := p.ctx
	ddSpan, _ := tracer.StartSpanFromContext(requestCtx, PusherTracerName)
	err = p.client.Trigger(channel, eventName, data)
	if err != nil {
		logs.CtxError(requestCtx, "trigger", logs.String("channel", channel), logs.String("event_name", eventName), logs.Any("data", data), logs.String("err", err.Error()))
	}
	ddSpan.Finish(tracer.WithError(err))
	return

}

func NewPusherClient(config PusherConfig) PusherClientWithCtx {
	return func(ctx context.Context) PusherClient {
		return &PusherClientImpl{
			ctx:    ctx,
			client: initPusher(config),
		}
	}
}

func initPusher(config PusherConfig) pusher.Client {
	return pusher.Client{
		AppID:   config.AppID,
		Key:     config.Key,
		Secret:  config.Secret,
		Cluster: config.Cluster,
		Secure:  config.Secure,
	}
}
