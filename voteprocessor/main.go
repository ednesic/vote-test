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
	ClusterID   string `envconfig:"CLUSTER_ID" default:"test-cluster" required:"true"`
	VoteChannel string `envconfig:"VOTE_CHANNEL" default:"create-vote" required:"true"`
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
		s.ClusterID,
		stan.NatsURL(stan.DefaultNatsURL),
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
