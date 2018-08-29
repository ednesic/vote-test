package main

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"

	"github.com/ednesic/vote-test/natsutil"
	"github.com/ednesic/vote-test/pb"
	"github.com/nats-io/go-nats-streaming"
)

const (
	clusterID  = "test-cluster"
	clientID   = "order-query-store2"
	channel    = "create-vote"
	durableID  = "store-durable1"
	queueGroup = "order-query-store-group"
)

func main() {
	comp := natsutil.NewStreamingComponent(clientID)
	err := comp.ConnectToNATSStreaming(
		clusterID,
		stan.NatsURL(stan.DefaultNatsURL),
	)
	if err != nil {
		log.Fatal(err)
	}
	sc := comp.NATS()

	sc.QueueSubscribe(channel, queueGroup, func(msg *stan.Msg) {

		vote := pb.VoteRequest{}
		err := json.Unmarshal(msg.Data, &vote)
		if err == nil {
			fmt.Println(vote)
		}
	}, stan.DurableName(durableID),
	)
	runtime.Goexit()
}
