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
type server struct {
	port          string `envconfig:"PORT" default:"9222"`
	natsClusterID string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	voteChannel   string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	natsServer    string `envconfig:"NATS_SERVER" default:"localhost:4222"`

	srv     *http.Server
	strmCmp *natsutil.StreamingComponent
}

func (s *server) run() {
	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: s.initRoutes(),
	}
	s.strmCmp = natsutil.NewStreamingComponent(clientID)
	s.connectNATS(s.strmCmp)
	log.Println("HTTP Sever listening on " + s.port)
	log.Fatal(server.ListenAndServe())
}

func (s *server) connectNATS(cmp *natsutil.StreamingComponent) {
	err := cmp.ConnectToNATSStreaming(
		s.natsClusterID,
		stan.NatsURL(s.natsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *server) initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/vote", s.createVote).Methods("POST")
	return router
}

func (s *server) createVote(w http.ResponseWriter, r *http.Request) {
	var vote pb.Vote

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid Order Data", 500)
		return
	}

	err = s.publishEvent(s.strmCmp, &vote)

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

func (s *server) publishEvent(component *natsutil.StreamingComponent, vote *pb.Vote) error {
	sc := component.NATS()
	voteJSON, err := proto.Marshal(vote)
	if err != nil {
		return err
	}
	return sc.Publish(s.voteChannel, voteJSON)
}

func main() {
	var server server
	err := envconfig.Process("", &server)
	if err != nil {
		log.Fatal(err.Error())
	}
	server.run()
}
