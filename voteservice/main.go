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
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats-streaming"
)

const clientID = "vote-service"

// Variaveis de ambiente
type Specification struct {
	Port        string `envconfig:"PORT" default:":9222" required:"true"`
	ClusterID   string `envconfig:"CLUSTER_ID" default:"test-cluster" required:"true"`
	VoteChannel string `envconfig:"VOTE_CHANNEL" default:"create-vote" required:"true"`
}

var (
	strmCmp *natsutil.StreamingComponent
	s       Specification
)

func main() {
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}

	server := &http.Server{
		Addr:    s.Port,
		Handler: initRoutes(),
	}

	strmCmp = natsutil.NewStreamingComponent(clientID)
	connectNATS(strmCmp)

	log.Println("HTTP Sever listening on " + s.Port)
	server.ListenAndServe()
}

func connectNATS(cmp *natsutil.StreamingComponent) {
	err := cmp.ConnectToNATSStreaming(
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
}

func initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/vote", createVote).Methods("POST")
	return router
}

func createVote(w http.ResponseWriter, r *http.Request) {
	var vote pb.Vote

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid Order Data", 500)
		return
	}

	err = publishEvent(strmCmp, &vote)

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

func publishEvent(component *natsutil.StreamingComponent, vote *pb.Vote) error {
	sc := component.NATS()
	voteJSON, err := proto.Marshal(vote)
	if err != nil {
		return err
	}
	return sc.Publish(s.VoteChannel, voteJSON)
}
