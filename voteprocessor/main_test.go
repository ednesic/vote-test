package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"testing"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/pb"
	"github.com/ednesic/vote-test/tests"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/nats-io/go-nats-streaming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gock "gopkg.in/h2non/gock.v1"
	"gopkg.in/mgo.v2/dbtest"
)

var Server dbtest.DBServer

func Example_procVote_election_not_found() {
	var s spec
	mgoDal := &tests.DataAccessLayerMock{}
	s.mgoDal = mgoDal
	mgoDal.On("FindOne", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("Not found")).Once()

	msg := &stan.Msg{}
	m, _ := proto.Marshal(&pb.Vote{ElectionId: 2, User: "Test2"})
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
	mgoDal := &tests.DataAccessLayerMock{}
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
	const server = "http://localhost"
	defer gock.Off()

	tests := []struct {
		name       string
		reply      int
		errorReply error
		jsonRes    string
		wantErr    bool
	}{
		{"Election termination found", 200, nil, `{"id": 1, "termino": { "seconds": 1536525322 }}`, false},
		{"Error on get", 500, errors.New("fail on get"), "nil", true},
		{"Error on body read", 500, nil, "", true},
		{"Election not ok", 404, nil, `{"id": 3, "termino": { "seconds": 1536525322 }}`, true},
		{"Election unmarshal error", 200, nil, `{city: 3, "termino": { "seconds": 1536525322 }}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gock.New("(.*)").
				Reply(tt.reply).
				BodyString(tt.jsonRes)

			termino, err := getElectionEnd(server, 1)
			if !tt.wantErr {
				var e pb.Election
				json.Unmarshal([]byte(tt.jsonRes), &e)
				assert.Nil(t, err)
				assert.Equal(t, termino, e.Termino)
			} else {
				assert.NotNil(t, err)
			}
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

/*******************************************/
/** Tests that must have mongod installed **/
/*******************************************/

func Example_spec_procVote_full() {
	cmd := exec.Command("mongod", "-version")
	if cmd.Run() != nil {
		log.Printf("Mongod is not installed.")
		return
	}

	var s spec
	s.Database = "test"
	s.Coll = "test1"

	tafter := ptypes.TimestampNow()

	tafter.Seconds = tafter.GetSeconds() + 1000

	tempDir, _ := ioutil.TempDir("", "testing")
	Server.SetPath(tempDir)
	mockSession := Server.Session()

	mockSession.DB(s.Database).C(s.Coll).Insert(&pb.Election{Id: 1, Inicio: tafter, Termino: tafter})

	mgoDal, err := db.NewMongoDAL(mockSession.LiveServers()[0], s.Database)
	if err != nil {
		log.Fatal("Failed to create mongo mock session")
	}
	s.mgoDal = mgoDal

	msg := &stan.Msg{}
	m, _ := proto.Marshal(&pb.Vote{ElectionId: 1, User: "test"})
	msg.Data = m
	err = msg.Unmarshal(m)
	if err != nil {
		log.Fatal("Failed to unmarshal message")
	}
	s.procVote(msg)
	// Output:

	mockSession.Close()
}

func Example_spec_procVote_election_ended() {
	cmd := exec.Command("mongod", "-version")
	err := cmd.Run()

	if err != nil {
		log.Printf("Mongod is not installed.")
		return
	}

	var s spec
	s.Database = "test"
	s.Coll = "test1"

	tbefore := ptypes.TimestampNow()

	tbefore.Seconds = tbefore.GetSeconds() - 10000

	tempDir, _ := ioutil.TempDir("", "testing")
	Server.SetPath(tempDir)
	mockSession := Server.Session()

	mockSession.DB(s.Database).C(s.Coll).Insert(&pb.Election{Id: 2, Inicio: tbefore, Termino: tbefore})

	mgoDal, err := db.NewMongoDAL(mockSession.LiveServers()[0], s.Database)
	if err != nil {
		log.Fatal("Failed to create mongo mock session")
	}
	s.mgoDal = mgoDal

	msg := &stan.Msg{}
	m, _ := proto.Marshal(&pb.Vote{ElectionId: 2, User: "test"})
	msg.Data = m
	err = msg.Unmarshal(m)
	if err != nil {
		log.Fatal("Failed to unmarshal message")
	}
	s.procVote(msg)
	// Output: Election has ended

	mockSession.Close()
}
