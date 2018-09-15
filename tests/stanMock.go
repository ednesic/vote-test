package tests

import (
	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/stretchr/testify/mock"
)

type StanConnMock struct {
	mock.Mock
}

func (s StanConnMock) Publish(subject string, data []byte) error {
	args := s.Called(subject, data)
	return args.Error(0)
}

func (s StanConnMock) PublishAsync(subject string, data []byte, ah stan.AckHandler) (string, error) {
	return "", nil
}
func (s StanConnMock) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return nil, nil
}
func (s StanConnMock) QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return nil, nil
}

func (s StanConnMock) Close() error {
	return nil
}

func (s StanConnMock) NatsConn() *nats.Conn {
	return nil
}
