package main

import (
	"net/http"
	"testing"

	"github.com/ednesic/vote-test/natsutil"
	"github.com/ednesic/vote-test/pb"
)

func Test_server_publishEvent(t *testing.T) {
	mckhttp := &http.Server{}
	mckStrmCmp := natsutil.NewStreamingComponent("mock")
	type fields struct {
		Port          string
		NatsClusterID string
		VoteChannel   string
		NatsServer    string
		ClientID      string
		srv           *http.Server
		strmCmp       *natsutil.StreamingComponent
	}
	type args struct {
		component *natsutil.StreamingComponent
		vote      *pb.Vote
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"test1", fields{"", "", "", "", "", mckhttp, mckStrmCmp}, args{mckStrmCmp, &pb.Vote{}}, false},
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
