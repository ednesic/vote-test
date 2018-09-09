package main

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/nats-io/go-nats-streaming"
	"github.com/stretchr/testify/mock"
)

type DataAccessLayerMock struct {
	mock.Mock
}

func (m *DataAccessLayerMock) Insert(collName string, doc interface{}) error {
	args := m.Called(collName, doc)
	return args.Error(0)
}
func (m *DataAccessLayerMock) FindOne(collName string, query interface{}, doc interface{}) error {
	args := m.Called(collName, query, doc)
	return args.Error(0)
}

func Example_procVote_election_not_found() {
	var s spec
	mgoDal := &DataAccessLayerMock{}
	s.mgoDal = mgoDal
	mgoDal.On("FindOne", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("Not found")).Once()

	msg := &stan.Msg{}
	m, _ := proto.Marshal(&pb.Vote{Id: 2, User: "Test2"})
	msg.Data = m
	err := msg.Unmarshal(m)
	if err != nil {
		log.Fatal("fail")
	}
	s.procVote(msg)
	// Output: Could not get election: Not found
}

func Example_procVote_fail_unmarshal() {
	var s spec
	mgoDal := &DataAccessLayerMock{}
	s.mgoDal = mgoDal
	mgoDal.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mgoDal.On("FindOne", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	msg := &stan.Msg{}
	msg.Unmarshal([]byte("test"))
	msg.Data = []byte("eee")
	s.procVote(msg)
	// Output:
	// Failed to process vote unexpected EOF
}

func Test_getElectionEnd(t *testing.T) {
	mgoDal := &DataAccessLayerMock{}

	tests := []struct {
		name     string
		queryRet interface{}
		wantErr  bool
	}{
		{"Get election", nil, false},
		{"Error on Find", errors.New("err"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgoDal.On("FindOne", mock.Anything, mock.Anything, mock.Anything).Return(tt.queryRet).Once()

			got, err := getElectionEnd(mgoDal, "test", 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("getElectionEnd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
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
	mgoDal := &DataAccessLayerMock{}
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
