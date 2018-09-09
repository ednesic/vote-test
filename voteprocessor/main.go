package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/nuid"
	"gopkg.in/mgo.v2/bson"
)

const (
	errEnvVarFail       = "Failed to get environment variables:"
	errConnLost         = "Connection lost:"
	errFailProcVote     = "Failed to process vote"
	errParseTimestamp   = "Failed to parse timestamp"
	errElectionNotFound = "Could not get election:"
	errElectionEnded    = "Election has ended"
)

type spec struct {
	VoteChannel   string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsClusterID string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	NatsServer    string `envconfig:"NATS_SERVER" default:"localhost:4222"`
	ClientID      string `envconfig:"CLIENT_ID" default:"vote-processor"`
	DurableID     string `envconfig:"DURABLE_ID" default:"store-durable"`
	QueueGroup    string `envconfig:"QUEUE_GROUP" default:"vote-processor-q"`
	MgoURL        string `envconfig:"MONGO_URL" default:"localhost:27017"`
	ElectionColl  string `envconfig:"ELECTION_COLLECTION" default:"election"`
	VoteColl      string `envconfig:"VOTE_COLLECTION" default:"vote"`
	Database      string `envconfig:"DATABASE" default:"elections"`
	mgoDal        db.DataAccessLayer
}

func main() {
	var s spec
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(errEnvVarFail, err.Error())
	}
	stanConn, err := stan.Connect(
		s.NatsClusterID,
		nuid.Next(),
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatal(errConnLost, reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	stanConn.QueueSubscribe(s.VoteChannel, s.QueueGroup, s.procVote, stan.DurableName(s.DurableID))

	s.mgoDal, err = db.NewMongoDAL(s.MgoURL, s.Database)
	if err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()
	defer stanConn.Close()
}

func (s *spec) procVote(msg *stan.Msg) {
	v := pb.Vote{}
	err := proto.Unmarshal(msg.Data, &v)
	if err != nil {
		fmt.Println(errFailProcVote, err)
		return
	}

	end, err := getElectionEnd(s.mgoDal, s.ElectionColl, v.Id)
	if err != nil {
		fmt.Println(errElectionNotFound, err)
		return
	}

	if isElectionOver(end) {
		fmt.Println(errElectionEnded)
		return
	}

	if err = vote(s.mgoDal, s.VoteColl, &v); err != nil {
		fmt.Println(errFailProcVote, err)
	}
}

func getElectionEnd(dal db.DataAccessLayer, coll string, id int32) (*timestamp.Timestamp, error) {
	var election pb.Election
	if err := dal.FindOne(coll, bson.M{"id": id}, &election); err != nil {
		return nil, err
	}
	return election.Termino, nil
}

func isElectionOver(end *timestamp.Timestamp) bool {
	t, err := ptypes.Timestamp(end)
	if err != nil {
		fmt.Println(errParseTimestamp, err)
		return true
	}
	return time.Now().After(t)
}

func vote(dal db.DataAccessLayer, coll string, vote *pb.Vote) error {
	return dal.Insert(coll, &vote)
}
