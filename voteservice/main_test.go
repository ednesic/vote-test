package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		body   io.Reader
		err    string
	}{
		// TODO: Add test cases.
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
			req, err := http.NewRequest("POST", "localhost:9222/vote", tt.body)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			rec := httptest.NewRecorder()
			s.createVote(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			//do tests
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			if tt.err != "" {
				t.Fatalf("todo: %v", err)
			}
			fmt.Println(b)
			fmt.Println(s)
			fmt.Println(req)
			fmt.Println(err)
		})
	}
}
