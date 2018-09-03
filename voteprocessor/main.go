package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/natsutil"
	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats-streaming"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	errEnvVarFail        = "Failed to get environment variables:"
	errConnLost          = "Connection lost:"
	errFailProcVote      = "Failed to process vote"
	errInvalidMgoSession = "Failed to retrieve mongo session"
	errRetrieveQuery     = "Failed to retrieve query"
	errParseTimestamp    = "Failed to parse timestamp"
	errConn              = "Failed connect to mongodb"
	errVoteOver          = "Passou o tempo da votacao"
)

type specification struct {
	VoteChannel   string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsClusterID string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	NatsServer    string `envconfig:"NATS_SERVER" default:"localhost:4222"`
	ClientID      string `envconfig:"CLIENT_ID" default:"vote-processor"`
	DurableID     string `envconfig:"DURABLE_ID" default:"store-durable"`
	QueueGroup    string `envconfig:"QUEUE_GROUP" default:"vote-processor-q"`
	Database      string `envconfig:"DATABASE" default:"elections"`
	ElectionColl  string `envconfig:"ELECTION_COLLECTION" default:"election"`
	VoteColl      string `envconfig:"VOTE_COLLECTION" default:"vote"`
	MgoURL        string `envconfig:"MONGO_URL" default:"localhost:27017"`
}

var s specification

func main() {
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(errEnvVarFail, err.Error())
	}
	comp := natsutil.NewStreamingComponent(s.ClientID)
	err = comp.ConnectToNATSStreaming(
		s.NatsClusterID,
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatal(errConnLost, reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	comp.NATS().QueueSubscribe(s.VoteChannel, s.QueueGroup, procVote, stan.DurableName(s.DurableID))
	runtime.Goexit()
	defer comp.Shutdown()
}

func procVote(msg *stan.Msg) {
	var election pb.Election
	vote := pb.Vote{}
	err := proto.Unmarshal(msg.Data, &vote)
	if err != nil {
		fmt.Println(errFailProcVote, err)
	}

	session, err := db.GetMongoSession(s.MgoURL)
	if err != nil {
		fmt.Println(errInvalidMgoSession, err)
		return
	}
	defer session.Close()
	//find election
	c := session.DB(s.Database).C(s.ElectionColl)

	if err = c.Find(bson.M{"id": vote.Id}).One(&election); err != nil {
		if err == mgo.ErrNotFound {
			fmt.Println(http.StatusNotFound, err)
			return
		}
		fmt.Println(errRetrieveQuery, err)
		return
	}
	t, err := ptypes.Timestamp(election.Termino)
	if err != nil {
		fmt.Println(errParseTimestamp, err)
		return
	}
	if time.Now().After(t) {
		fmt.Println(errVoteOver)
		return
	}
	//submit vote
	cvote := session.DB(s.Database).C(s.VoteColl)

	if cvote.Insert(&vote) != nil {
		fmt.Println(errConn)
		return
	}
}
