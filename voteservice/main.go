package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ednesic/vote-test/natsutil"
	"github.com/ednesic/vote-test/pb"
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

	comp := natsutil.NewStreamingComponent(clientID)
	err := comp.ConnectToNATSStreaming(
		clusterID,
		stan.NatsURL(stan.DefaultNatsURL),
	)
	if err != nil {
		log.Fatal(err)
	}
	//arrumar
	streamingComponent = comp

	log.Println("HTTP Sever listening...")
	server.ListenAndServe()
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

	go publishEvent(streamingComponent, &vote)

	fmt.Println("created vote")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	j, _ := json.Marshal(vote)
	w.Write(j)
}

func publishEvent(component *natsutil.StreamingComponent, vote *pb.VoteRequest) {
	sc := component.NATS()
	eventMsg := []byte(vote.GetUser())
	sc.Publish(channel, eventMsg)
	log.Println("Published message on channel: " + channel)
}
