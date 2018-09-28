package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	"go.uber.org/zap"
)

const (
	errEnvVarFail       = "Failed to get environment variables:"
	errConnLost         = "Connection lost:"
	errConnFail         = "Connection failed"
	errFailProcVote     = "Failed to process vote"
	errParseTimestamp   = "Failed to parse timestamp"
	errElectionNotFound = "Could not get election:"
	errElectionEnded    = "Election has ended"

	voteProcessed   = "Vote processed"
	initVoteProcMsg = "Processor running"
)

type spec struct {
	VoteChannel     string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsClusterID   string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	NatsServer      string `envconfig:"NATS_SERVER" default:"localhost:4222"`
	ClientID        string `envconfig:"CLIENT_ID" default:"vote-processor"`
	DurableID       string `envconfig:"DURABLE_ID" default:"vote-processor"`
	QueueGroup      string `envconfig:"QUEUE_GROUP" default:"vote-processor"`
	MgoURL          string `envconfig:"MONGO_URL" default:"localhost:27017"`
	Coll            string `envconfig:"COLLECTION" default:"election"`
	Database        string `envconfig:"DATABASE" default:"elections"`
	ElectionService string `envconfig:"ELECTION_SERVICE" default:"http://localhost:9223"`

	mgoDal db.DataAccessLayer
	logger *zap.Logger
}

func main() {
	var (
		err error
		s   spec
	)
	s.logger, err = zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	err = envconfig.Process("", &s)
	if err != nil {
		s.logger.Fatal(errEnvVarFail, zap.Error(err))
	}

	stanConn, err := stan.Connect(
		s.NatsClusterID,
		nuid.Next(),
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			s.logger.Fatal(errConnLost, zap.Error(reason))
		}),
	)
	if err != nil {
		s.logger.Fatal(errConnFail, zap.Error(err))
	}

	sub, err := stanConn.QueueSubscribe(s.VoteChannel, s.QueueGroup, s.procVote)
	if err != nil {
		s.logger.Fatal(errConnFail, zap.Error(err))
	}

	s.mgoDal, err = db.NewMongoDAL(s.MgoURL, s.Database)
	if err != nil {
		s.logger.Fatal(errConnFail, zap.Error(err))
	}

	s.logger.Info(initVoteProcMsg)
	defer s.logger.Sync()
	runtime.Goexit()
	defer sub.Unsubscribe()
	defer stanConn.Close()
}

func (s *spec) procVote(msg *stan.Msg) {
	var (
		err error
		v   pb.Vote
	)
	defer func() {
		s.logger.Info(voteProcessed, zap.Error(err), zap.Int32("electionId", v.ElectionId), zap.String("User", v.User))
	}()

	err = proto.Unmarshal(msg.Data, &v)
	if err != nil {
		return
	}

	end, err := getElectionEnd(s.ElectionService, v.ElectionId)
	if err != nil {
		return
	}

	if isElectionOver(end) {
		err = errors.New(errElectionEnded)
		return
	}

	err = vote(s.mgoDal, s.Coll, &v)
}

func getElectionEnd(serviceName string, id int32) (*timestamp.Timestamp, error) {
	var e pb.Election
	resp, err := http.Get(serviceName + "/election/" + fmt.Sprint(id))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	err = json.Unmarshal(body, &e)
	if err != nil {
		return nil, err
	}

	return e.Termino, nil
}

func isElectionOver(end *timestamp.Timestamp) bool {
	t, err := ptypes.Timestamp(end)
	if err != nil {
		return true
	}
	return time.Now().After(t)
}

func vote(dal db.DataAccessLayer, coll string, vote *pb.Vote) error {
	return dal.Insert(coll, &vote)
}
