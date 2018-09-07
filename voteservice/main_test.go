package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ednesic/vote-test/pb"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type stanConnMock struct {
	mock.Mock
}

func (s stanConnMock) Publish(subject string, data []byte) error {
	args := s.Called(subject, data)
	return args.Error(0)
}

func (s stanConnMock) PublishAsync(subject string, data []byte, ah stan.AckHandler) (string, error) {
	return "", nil
}
func (s stanConnMock) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return nil, nil
}
func (s stanConnMock) QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return nil, nil
}

func (s stanConnMock) Close() error {
	return nil
}

func (s stanConnMock) NatsConn() *nats.Conn {
	return nil
}

func Test_server_publishEvent(t *testing.T) {
	stanMock := new(stanConnMock)
	stanMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	type args struct {
		vote *pb.Vote
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Empty vote", args{&pb.Vote{}}, false},
		{"Empty user on vote", args{&pb.Vote{User: "", Id: 12}}, false},
		{"Empty id on vote", args{&pb.Vote{User: "1", Id: 0}}, false},
		{"Normal vote", args{&pb.Vote{User: "1", Id: 90}}, false},
		{"Negative id on vote", args{&pb.Vote{User: "1", Id: -1}}, false},
		{"High value on vote", args{&pb.Vote{User: "1", Id: 1000000000}}, false},
		{"Big string on vote", args{&pb.Vote{User: "sdgnasodgsdgsino", Id: 1}}, false},
		{"Nil vote must fail", args{nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{stanConn: stanMock}
			if err := s.publishEvent(tt.args.vote); (err != nil) != tt.wantErr {
				t.Errorf("server.publishEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_server_createVote(t *testing.T) {
	stanMock := new(stanConnMock)
	tests := []struct {
		name         string
		body         string
		statusCode   int
		responseBody string
		pubRes       error
	}{
		{"Could not process message", `{"id":12,"user":"abc"}`, http.StatusUnprocessableEntity, errFailPubVote, errors.New("err")},
		{"Creation successful", `{"id":12,"user":"abc"}`, http.StatusCreated, `{"id":12,"user":"abc"}`, nil},
		{"Wrong user type", `{"id":"12"}`, http.StatusBadRequest, errInvalidData, nil},
		{"Wrong id type", `{"user":12}`, http.StatusBadRequest, errInvalidData, nil},
		{"Missing user", `{"id":12}`, http.StatusBadRequest, errInvalidUser, nil},
		{"Missing id", `{"user":"abc"}`, http.StatusBadRequest, errInvalidID, nil},
		{"Missing all", `{}`, http.StatusBadRequest, errInvalidID, nil},
		{"Wrong parameters", `{"id":12,"user":"abc","Home: 5"}`, http.StatusBadRequest, errInvalidData, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stanMock.On("Publish", mock.Anything, mock.Anything).Return(tt.pubRes).Once()
			s := &server{
				stanConn: stanMock,
			}
			req, err := http.NewRequest("POST", "localhost:9222/vote", strings.NewReader(tt.body))
			assert.Nil(t, err, "could not create request")

			rec := httptest.NewRecorder()
			s.createVote(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.statusCode, res.StatusCode, "Did not get the same response code")
			assert.Nil(t, err, "could not read response body")
			assert.Equal(t, tt.responseBody, strings.TrimSuffix(rec.Body.String(), "\n"))
		})
	}
}

func Test_server_initRoutes(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Return non nil object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{}
			assert.NotNil(t, s.initRoutes())
		})
	}
}
