package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/nats-io/go-nats-streaming"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/dbtest"
)

var Server dbtest.DBServer

func Test_spec_procVote(t *testing.T) {
	var s spec

	tnow := ptypes.TimestampNow()
	tbefore := tnow
	tafter := tnow
	tinvalid := tnow

	tbefore.Seconds = tbefore.GetSeconds() - 1000
	tafter.Seconds = tafter.GetSeconds() + 1000
	tinvalid.Seconds = -1

	/**** must have mongod installed!!! mongod test ****/
	tempDir, _ := ioutil.TempDir("", "testing")
	Server.SetPath(tempDir)
	mockSession := Server.Session()
	mockSession.DB(s.Database).C(s.ElectionColl).Insert(&pb.Election{Id: 1, Inicio: tafter, Termino: tafter})
	mockSession.DB(s.Database).C(s.ElectionColl).Insert(&pb.Election{Id: 2, Inicio: tbefore, Termino: tbefore})
	mockSession.DB(s.Database).C(s.ElectionColl).Insert(&pb.Election{Id: 3, Inicio: tinvalid, Termino: tinvalid})
	// mockSession.Close()

	s.mgoSession = mockSession
	/**** mongod test ****/

	type args struct {
		msg *pb.Vote
	}
	tests := []struct {
		name string
		args args
	}{
		{"Marshalling failling", args{}},
		{"Marshaling passing", args{msg: &pb.Vote{Id: 2, User: "Test2"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &stan.Msg{}
			m, _ := proto.Marshal(tt.args.msg)
			err := msg.Unmarshal(m)
			if err != nil {
				log.Fatal("fail")
			}
			fmt.Println("fdfd ", msg.Data)
			s.procVote(msg)
		})
	}
}

func Test_getElectionEnd(t *testing.T) {
	type args struct {
		s    *mgo.Session
		db   string
		coll string
		id   int32
	}
	tests := []struct {
		name    string
		args    args
		want    *timestamp.Timestamp
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getElectionEnd(tt.args.s, tt.args.db, tt.args.coll, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("getElectionEnd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getElectionEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isElectionOver(t *testing.T) {
	type args struct {
		end *timestamp.Timestamp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
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
	type args struct {
		s    *mgo.Session
		db   string
		coll string
		vote *pb.Vote
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := vote(tt.args.s, tt.args.db, tt.args.coll, tt.args.vote); (err != nil) != tt.wantErr {
				t.Errorf("vote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
