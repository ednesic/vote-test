package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/ednesic/vote-test/natsutil"
	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats-streaming"
)

const (
	ErrEnvVarFail int = iota + 1
	ErrConnLost
	ErrFailProcVote
)

var errCodeToMessage = map[int]string{
	ErrEnvVarFail:   "Failed to get environment variables:",
	ErrConnLost:     "Connection lost:",
	ErrFailProcVote: "Failed to process vote",
}

type Specification struct {
	VoteChannel   string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsClusterID string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	NatsServer    string `envconfig:"NATS_SERVER" default:"localhost:4222"`
}

const (
	clientID   = "vote-processor"
	durableID  = "store-durable"
	queueGroup = "vote-processor-q"
)

var s Specification

func main() {
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(errCodeToMessage[ErrEnvVarFail], err.Error())
	}
	comp := natsutil.NewStreamingComponent(clientID)
	err = comp.ConnectToNATSStreaming(
		s.NatsClusterID,
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatal(errCodeToMessage[ErrConnLost], reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	comp.NATS().QueueSubscribe(s.VoteChannel, queueGroup, procVote, stan.DurableName(durableID))
	runtime.Goexit()
}

func procVote(msg *stan.Msg) {
	vote := pb.Vote{}
	err := proto.Unmarshal(msg.Data, &vote)
	if err != nil {
		fmt.Println(ErrFailProcVote, err)
	}
	fmt.Println(vote)
}
