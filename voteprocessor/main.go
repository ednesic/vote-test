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
	VoteChannel string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsServer  string `envconfig:"NATS_SERVER" default:"create-vote"`
}

const (
	clientID   = "order-query-store2"
	durableID  = "store-durable1"
	queueGroup = "order-query-store-group"
)

var s Specification

func main() {
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	comp := natsutil.NewStreamingComponent(clientID)
	err = comp.ConnectToNATSStreaming(
		"test-cluster",
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	sc := comp.NATS()
	sc.QueueSubscribe(s.VoteChannel, queueGroup, func(msg *stan.Msg) {

		vote := pb.Vote{}
		err := proto.Unmarshal(msg.Data, &vote)
		//verificar erro
		if err == nil {
			fmt.Println(vote)
		}
	}, stan.DurableName(durableID),
	)
	runtime.Goexit()
}
