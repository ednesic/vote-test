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
	channel    = "event.Channel"
	durableID  = "store-durable1"
	queueGroup = "order-query-store-group"
)

func main() {
	// Register new NATS component within the system.
	comp := natsutil.NewStreamingComponent(clientID)

	// Connect to NATS Streaming server
	err := comp.ConnectToNATSStreaming(
		clusterID,
		stan.NatsURL(stan.DefaultNatsURL),
	)
	if err != nil {
		log.Fatal(err)
	}
	// Get the NATS Streaming Connection
	sc := comp.NATS()
	// aw, _ := time.ParseDuration("60s")
	sc.QueueSubscribe(channel, queueGroup, func(msg *stan.Msg) {
		// msg.Ack() // Manual ACK

		vote := pb.VoteRequest{}
		err := json.Unmarshal(msg.Data, &vote)
		if err == nil {
			fmt.Println(vote)
		}
		fmt.Println("passei")
	}, stan.DurableName(durableID),
	// 	stan.MaxInflight(25),
	// 	stan.SetManualAckMode(),
	// 	stan.AckWait(aw),
	)
	runtime.Goexit()
}
