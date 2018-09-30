package main

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/pb"
	"github.com/ednesic/vote-test/tests"
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
		candidate  string
		id         int32
		wantErr    bool
	}{
		{"Request error", 0, errors.New("Error request"), "candidateMock1", 1, true},
		{"Request Ok", http.StatusOK, nil, "candidateMock1", 1, false},
		{"Request fail", http.StatusGone, nil, "candidateMock1", 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.errorReply != nil {
				gock.New(server).
					Get("/election/"+fmt.Sprint(tt.id)+"/validate").
					MatchParam("candidate", tt.candidate).
					ReplyError(tt.errorReply)
			} else {
				gock.New(server).
					Get("/election/"+fmt.Sprint(tt.id)+"/validate").
					MatchParam("candidate", tt.candidate).
					Reply(tt.reply)
			}
			err := validateVote(server, &pb.Vote{Candidate: tt.candidate, ElectionId: tt.id})
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
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
