package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/ednesic/vote-test/pb"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/nuid"
	"go.uber.org/zap"
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

	logger   *zap.Logger
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

	s.logger, err = zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	defer s.logger.Sync()
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
		defer s.logger.Error(errInvalidData, zap.Int32("electionId", vote.ElectionId), zap.String("User", vote.User), zap.Int("StatusCode", http.StatusBadRequest), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}
	if vote.GetElectionId() <= 0 {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		defer s.logger.Error(errInvalidID, zap.Int32("electionId", vote.ElectionId), zap.String("User", vote.User), zap.Int("StatusCode", http.StatusBadRequest), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}
	if vote.GetUser() == "" {
		http.Error(w, errInvalidUser, http.StatusBadRequest)
		defer s.logger.Error(errInvalidUser, zap.Int32("electionId", vote.ElectionId), zap.String("User", vote.User), zap.Int("StatusCode", http.StatusBadRequest), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	err = s.publishEvent(&vote)
	if err != nil {
		http.Error(w, errFailPubVote, http.StatusInternalServerError)
		defer s.logger.Error(errFailPubVote, zap.Int32("electionId", vote.ElectionId), zap.String("User", vote.User), zap.Int("StatusCode", http.StatusInternalServerError), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	w.WriteHeader(http.StatusCreated)
	j, _ := json.Marshal(vote)
	w.Write(j)
	defer s.logger.Info("Vote Created", zap.Int32("electionId", vote.ElectionId), zap.String("User", vote.User), zap.Int("StatusCode", http.StatusCreated), zap.String("hostname", os.Getenv("HOSTNAME")))
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
