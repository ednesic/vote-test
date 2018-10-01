package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/pb"
	"github.com/ednesic/vote-test/tests"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gock "gopkg.in/h2non/gock.v1"
	"gopkg.in/mgo.v2/dbtest"
)

var Server dbtest.DBServer

func Test_getElectionEnd(t *testing.T) {
	const server = "http://localhost"
	defer gock.Off()

	tests := []struct {
		name       string
		reply      int
		errorReply error
		jsonRes    string
		wantErr    bool
		err        string
		id         int32
	}{
		{"Election termination found", 200, nil, `{"id": 1, "end": { "seconds": 1536525322 }}`, false, "", 1},
		{"Error on get", 500, errors.New("fail on get"), "nil", true, "fail on get", 2},
		{"Election not ok", 404, nil, "election not found", true, "election not found", 4},
		{"Election unmarshal error", 200, nil, `{city: 3, "end": { "seconds": 1536525322 }}`, true, "invalid character", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.errorReply != nil {
				gock.New(server).
					Get("/election/" + fmt.Sprint(tt.id)).
					ReplyError(tt.errorReply)
			} else {
				gock.New(server).
					Get("/election/" + fmt.Sprint(tt.id)).
					Reply(tt.reply).
					BodyString(tt.jsonRes)
			}

			termino, err := getElectionEnd(server, tt.id)
			if !tt.wantErr {
				var e pb.Election
				json.Unmarshal([]byte(tt.jsonRes), &e)
				assert.Nil(t, err)
				assert.Equal(t, termino, e.End)
				return
			}

			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

func Test_isElectionOver(t *testing.T) {
	tbefore := ptypes.TimestampNow()
	tafter := ptypes.TimestampNow()
	tinvalid := ptypes.TimestampNow()

	tbefore.Seconds = tbefore.GetSeconds() - 1000
	tafter.Seconds = tafter.GetSeconds() + 1000
	tinvalid.Seconds = -1000
	tinvalid.Nanos = -1000

	type args struct {
		end *timestamp.Timestamp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Time before now", args{end: tbefore}, true},
		{"Time after now", args{end: tafter}, false},
		{"Invalid time", args{end: tinvalid}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isElectionOver(tt.args.end); got != tt.want {
				t.Errorf("isElectionOver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_vote(t *testing.T) {
	mgoDal := &tests.DataAccessLayerMock{}
	type args struct {
		dal  db.DataAccessLayer
		coll string
		vote *pb.Vote
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		queryRet error
	}{
		{"Insert work", args{dal: mgoDal, coll: "test", vote: &pb.Vote{}}, false, nil},
		{"Insert fail", args{dal: mgoDal, coll: "test", vote: &pb.Vote{}}, true, errors.New("err")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgoDal.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(tt.queryRet).Once()
			if err := vote(tt.args.dal, tt.args.coll, tt.args.vote); (err != nil) != tt.wantErr {
				t.Errorf("vote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
