package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	errFailProcVote     = "Failed to process vote"
	errParseTimestamp   = "Failed to parse timestamp"
	errElectionNotFound = "Could not get election:"
	errElectionEnded    = "Election has ended"
)

type spec struct {
	VoteChannel     string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsClusterID   string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	NatsServer      string `envconfig:"NATS_SERVER" default:"localhost:4222"`
	ClientID        string `envconfig:"CLIENT_ID" default:"vote-processor"`
	DurableID       string `envconfig:"DURABLE_ID" default:"store-durable"`
	QueueGroup      string `envconfig:"QUEUE_GROUP" default:"vote-processor-q"`
	MgoURL          string `envconfig:"MONGO_URL" default:"localhost:27017"`
	Coll            string `envconfig:"COLLECTION" default:"election"`
	Database        string `envconfig:"DATABASE" default:"elections"`
	ElectionService string `envconfig:"ELECTION_SERVICE" default:"localhost:9223"`

	mgoDal db.DataAccessLayer
	logger *zap.Logger
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

	s.logger, err = zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	stanConn.QueueSubscribe(s.VoteChannel, s.QueueGroup, s.procVote, stan.DurableName(s.DurableID))

	s.mgoDal, err = db.NewMongoDAL(s.MgoURL, s.Database)
	if err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()
	defer s.logger.Sync()
	defer stanConn.Close()
}

func (s *spec) procVote(msg *stan.Msg) {
	v := pb.Vote{}
	err := proto.Unmarshal(msg.Data, &v)
	if err != nil {
		fmt.Println(errFailProcVote, err)
		defer s.logger.Error(errFailProcVote, zap.String("electionService", s.ElectionService), zap.Error(err), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	end, err := getElectionEnd(s.ElectionService, v.ElectionId)
	if err != nil {
		fmt.Println(errElectionNotFound, err)
		defer s.logger.Error(errElectionNotFound, zap.String("electionService", s.ElectionService), zap.Int32("electionId", v.ElectionId), zap.String("User", v.User), zap.Error(err), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	if isElectionOver(end) {
		defer s.logger.Warn(errElectionEnded, zap.Int32("electionId", v.ElectionId), zap.String("User", v.User), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	if err = vote(s.mgoDal, s.Coll, &v); err != nil {
		fmt.Println(errFailProcVote, err)
		defer s.logger.Error(errElectionEnded, zap.Int32("electionId", v.ElectionId), zap.Error(err), zap.String("User", v.User), zap.String("hostname", os.Getenv("HOSTNAME")))
	}
	defer s.logger.Info("Vote processed", zap.Int32("electionId", v.ElectionId), zap.String("User", v.User), zap.String("hostname", os.Getenv("HOSTNAME")))
}

func getElectionEnd(serviceName string, id int32) (*timestamp.Timestamp, error) {
	var e pb.Election
	resp, err := http.Get(serviceName + "/election/" + string(id))
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
