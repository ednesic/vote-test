package main

import (
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
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/mgo.v2/dbtest"
)

var Server dbtest.DBServer

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

func Test_containsString(t *testing.T) {
	type args struct {
		s   string
		arr []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Contains", args{s: "test1", arr: []string{"test1", "test2"}}, true},
		{"Not Contains", args{s: "test3", arr: []string{"test1", "test2"}}, false},
		{"Empty arr", args{s: "test3", arr: nil}, false},
		{"Empty s", args{s: "", arr: []string{"test1", "test2"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsString(tt.args.s, tt.args.arr); got != tt.want {
				t.Errorf("containsCandidate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getElectionEnd(t *testing.T) {
	const server = "http://localhost"
	defer gock.Off()
	const id = 1

	tests := []struct {
		name              string
		reply             int
		errorReply        error
		jsonRes           string
		expectedErr       string
		containsStringRes bool
		isElectionOverRes bool
	}{
		{"Election valid", 200, nil, "{}", "", true, false},
		{"Election over", 200, nil, "{}", errElectionEnded, true, true},
		{"Candidate not found", 200, nil, "{}", errCandidateNotFound, false, false},
		{"Election invalid", 404, nil, errCandidateNotFound, errCandidateNotFound, true, true},
		{"Err on request", 0, errors.New("internal error"), "", "internal error", true, false},
		{"Unmarshal fail", 200, nil, "{", "unexpected end of JSON input", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			containsString = func(s string, arrs []string) bool {
				return tt.containsStringRes
			}

			isElectionOver = func(end *timestamp.Timestamp) bool {
				return tt.isElectionOverRes
			}

			if tt.errorReply != nil {
				gock.New(server).
					Get("/election/" + fmt.Sprint(id)).
					ReplyError(tt.errorReply)
			} else {
				gock.New(server).
					Get("/election/" + fmt.Sprint(id)).
					Reply(tt.reply).
					BodyString(tt.jsonRes)
			}

			err := verifyElection(server, &pb.Vote{Candidate: "", ElectionId: id})
			if tt.expectedErr != "" {
				assert.Contains(t, err.Error(), tt.expectedErr)
				return
			}

			assert.Nil(t, err)
		})
	}
}
