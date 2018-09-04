package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ednesic/vote-test/natsutil"
	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats-streaming"
)

const (
	ErrEnvVarFail  = "Failed to get environment variables:"
	ErrConnLost    = "Connection lost:"
	ErrFailPubVote = "Failed to publish vote"
	ErrInvalidData = "Invalid Vote Data"
	ErrInvalidId   = "Invalid Id"
	ErrInvalidUser = "Invalid User"
)

// Variaveis de ambiente
type server struct {
	Port          string `envconfig:"PORT" default:"9222"`
	NatsClusterID string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	VoteChannel   string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsServer    string `envconfig:"NATS_SERVER" default:"localhost:4222"`
	ClientID      string `envconfig:"CLIENT_ID" default:"vote-service"`

	srv     *http.Server
	strmCmp natsutil.StreamingComponent
}

func (s *server) run() {
	server := &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.initRoutes(),
	}

	s.strmCmp = natsutil.NewStreamingComponent(s.ClientID)
	err := s.strmCmp.ConnectToNATSStreaming(
		s.NatsClusterID,
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatal(ErrConnLost, reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer s.strmCmp.Shutdown()
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
		http.Error(w, ErrInvalidData, http.StatusBadRequest)
		return
	}

	if vote.GetId() <= 0 {
		http.Error(w, ErrInvalidId, http.StatusBadRequest)
		return
	}

	if vote.GetUser() == "" {
		http.Error(w, ErrInvalidUser, http.StatusBadRequest)
		return
	}

	err = s.publishEvent(s.strmCmp, &vote)

	if err != nil {
		http.Error(w, ErrFailPubVote, http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusCreated)
	j, _ := json.Marshal(vote)
	w.Write(j)
}

func (s *server) publishEvent(component natsutil.StreamingComponent, vote *pb.Vote) error {
	sc := component.NATS()
	voteJSON, err := proto.Marshal(vote)
	if err != nil {
		return err
	}
	return sc.Publish(s.VoteChannel, voteJSON)
}

func main() {
	var s server
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(ErrEnvVarFail, err.Error())
	}
	s.run()
}
