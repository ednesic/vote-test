package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ednesic/vote-test/natsutil"
	"github.com/ednesic/vote-test/pb"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
)

type stanConnMock struct{}

func (s stanConnMock) Publish(subject string, data []byte) error {
	return nil
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

type strmCmpMock struct{}

func (s strmCmpMock) ConnectToNATSStreaming(clusterID string, options ...stan.Option) error {
	return nil
}

func (s strmCmpMock) NATS() stan.Conn {
	return stanConnMock{}
}

func (s strmCmpMock) ID() string {
	return ""
}

func (s strmCmpMock) Name() string {
	return ""
}

func (s strmCmpMock) Shutdown() error {
	return nil
}

func Test_server_publishEvent(t *testing.T) {
	type fields struct {
		Port          string
		NatsClusterID string
		VoteChannel   string
		NatsServer    string
		ClientID      string
		srv           *http.Server
		strmCmp       natsutil.StreamingComponent
	}
	type args struct {
		component natsutil.StreamingComponent
		vote      *pb.Vote
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Empty vote", fields{}, args{strmCmpMock{}, &pb.Vote{}}, false},
		{"Empty user on vote", fields{}, args{strmCmpMock{}, &pb.Vote{User: "", Id: 12}}, false},
		{"Empty id on vote", fields{}, args{strmCmpMock{}, &pb.Vote{User: "1", Id: 0}}, false},
		{"Normal vote", fields{}, args{strmCmpMock{}, &pb.Vote{User: "1", Id: 90}}, false},
		{"Negative id on vote", fields{}, args{strmCmpMock{}, &pb.Vote{User: "1", Id: -1}}, false},
		{"High value on vote", fields{}, args{strmCmpMock{}, &pb.Vote{User: "1", Id: 1000000000}}, false},
		{"Big string on vote", fields{}, args{strmCmpMock{}, &pb.Vote{User: "sdgnasodgsdgsino", Id: 1}}, false},
		{"Nil vote must fail", fields{}, args{strmCmpMock{}, nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				Port:          tt.fields.Port,
				NatsClusterID: tt.fields.NatsClusterID,
				VoteChannel:   tt.fields.VoteChannel,
				NatsServer:    tt.fields.NatsServer,
				ClientID:      tt.fields.ClientID,
				srv:           tt.fields.srv,
				strmCmp:       tt.fields.strmCmp,
			}
			if err := s.publishEvent(tt.args.component, tt.args.vote); (err != nil) != tt.wantErr {
				t.Errorf("server.publishEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_server_createVote(t *testing.T) {
	type fields struct {
		Port          string
		NatsClusterID string
		VoteChannel   string
		NatsServer    string
		ClientID      string
		srv           *http.Server
		strmCmp       natsutil.StreamingComponent
	}
	tests := []struct {
		name         string
		fields       fields
		body         string
		statusCode   int
		responseBody string
	}{
		{"Creation successful", fields{strmCmp: strmCmpMock{}}, `{"id":12,"user":"abc"}`, 201, `{"id":12,"user":"abc"}`},
		{"Missing user", fields{strmCmp: strmCmpMock{}}, `{"id":12}`, 400, ErrInvalidUser},
		{"Missing id", fields{strmCmp: strmCmpMock{}}, `{"user":"abc"}`, 400, ErrInvalidId},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				Port:          tt.fields.Port,
				NatsClusterID: tt.fields.NatsClusterID,
				VoteChannel:   tt.fields.VoteChannel,
				NatsServer:    tt.fields.NatsServer,
				ClientID:      tt.fields.ClientID,
				srv:           tt.fields.srv,
				strmCmp:       tt.fields.strmCmp,
			}
			req, err := http.NewRequest("POST", "localhost:9222/vote", strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			rec := httptest.NewRecorder()
			s.createVote(rec, req)

			res := rec.Result()

			if res.StatusCode != tt.statusCode {
				t.Fatalf("Did not get the same response code: %v", err)
			}

			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response body: %v", err)
			}

			fmt.Println(tt.responseBody)
			fmt.Println(string(b))
			if tt.responseBody != string(b) {
				t.Fatalf("POST did not return the same object passed on the request body ")
			}
		})
	}
}
