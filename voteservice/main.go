package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/nuid"
)

const (
	errEnvVarFail  = `Failed to get environment variables:`
	errConnLost    = `Connection lost:`
	errFailPubVote = `Failed to publish vote`
	errInvalidData = `Invalid Vote Data`
	errInvalidID   = `Invalid Id`
	errInvalidUser = `Invalid User`
)

type server struct {
	Port          string `envconfig:"PORT" default:"9222"`
	NatsClusterID string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	VoteChannel   string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsServer    string `envconfig:"NATS_SERVER" default:"localhost:4222"`
	ClientID      string `envconfig:"CLIENT_ID" default:"vote-service"`

	srv      *http.Server
	stanConn stan.Conn
}

func (s *server) run() {
	var err error
	server := &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.initRoutes(),
	}

	s.stanConn, err = stan.Connect(
		s.NatsClusterID,
		nuid.Next(),
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatal(errConnLost, reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer s.stanConn.Close()
	log.Println("HTTP Sever listening on " + s.Port)
	log.Fatal(server.ListenAndServe())
}

func (s *server) initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/vote", s.createVote).Methods("POST")
	return router
}

func (s *server) createVote(w http.ResponseWriter, r *http.Request) {
	var vote pb.Vote
	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, errInvalidData, http.StatusBadRequest)
		return
	}
	if vote.GetId() <= 0 {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}
	if vote.GetUser() == "" {
		http.Error(w, errInvalidUser, http.StatusBadRequest)
		return
	}

	err = s.publishEvent(&vote)
	if err != nil {
		http.Error(w, errFailPubVote, http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusCreated)
	j, _ := json.Marshal(vote)
	w.Write(j)
}

func (s *server) publishEvent(vote *pb.Vote) error {
	voteJSON, err := proto.Marshal(vote)
	if err != nil {
		return err
	}
	return s.stanConn.Publish(s.VoteChannel, voteJSON)
}

func main() {
	var s server
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(errEnvVarFail, err.Error())
	}
	s.run()
}
