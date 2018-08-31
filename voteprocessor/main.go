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
		log.Fatal(err.Error())
	}
	comp := natsutil.NewStreamingComponent(clientID)
	err = comp.ConnectToNATSStreaming(
		s.NatsClusterID,
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	comp.NATS().QueueSubscribe(s.VoteChannel, queueGroup, func(msg *stan.Msg) {
		vote := pb.Vote{}
		err := proto.Unmarshal(msg.Data, &vote)
		if err != nil {
			fmt.Println("Falha ao gerar Voto", err)
		}
		fmt.Println(vote)
	}, stan.DurableName(durableID),
	)
	runtime.Goexit()
}
