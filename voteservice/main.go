package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ednesic/vote-test/natsutil"
	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/nats-io/go-nats-streaming"
)

const (
	port      = ":9222"
	clientID  = "event-store-api"
	clusterID = "test-cluster"
	channel   = "create-vote"
)

var streamingComponent *natsutil.StreamingComponent

func main() {
	server := &http.Server{
		Addr:    port,
		Handler: initRoutes(),
	}

	streamingComponent = natsutil.NewStreamingComponent(clientID)
	connectNATS(streamingComponent)

	log.Println("HTTP Sever listening on " + port)
	server.ListenAndServe()
}

func connectNATS(cmp *natsutil.StreamingComponent) {
	err := cmp.ConnectToNATSStreaming(
		clusterID,
		stan.NatsURL(stan.DefaultNatsURL),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/vote", createVote).Methods("POST")
	return router
}

func createVote(w http.ResponseWriter, r *http.Request) {
	var vote pb.VoteRequest

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid Order Data", 500)
		return
	}

	err = publishEvent(streamingComponent, &vote)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println("Could not publish message", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusCreated)
	j, _ := json.Marshal(vote)
	w.Write(j)
}

func publishEvent(component *natsutil.StreamingComponent, vote *pb.VoteRequest) error {
	sc := component.NATS()
	voteJSON, err := proto.Marshal(vote)
	if err != nil {
		return err
	}
	return sc.Publish(channel, voteJSON)
}
