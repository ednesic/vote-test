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
	"go.uber.org/zap"
)

const (
	errEnvVarFail  = `Failed to get environment variables:`
	errConnLost    = `Connection lost:`
	errConnFailed  = `Connection failed`
	errFailPubVote = `Failed to publish vote`
	errInvalidData = `Invalid Vote Data`
	errInvalidID   = `Invalid Id`
	errInvalidUser = `Invalid User`
	errInterrupt   = `Shutting down`

	listenMsg     = "HTTP Sever listening"
	voteCreateMsg = "POST vote creation"
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

	s.logger, err = zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	err = envconfig.Process("", s)
	if err != nil {
		s.logger.Fatal(errEnvVarFail, zap.Error(err))
	}

	srv := &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.initRoutes(),
	}

	s.stanConn, err = stan.Connect(
		s.NatsClusterID,
		nuid.Next(),
		stan.NatsURL(s.NatsServer),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			s.logger.Fatal(errConnLost, zap.Error(reason))
		}),
	)
	if err != nil {
		s.logger.Fatal(errConnFailed, zap.Error(err))
	}

	defer s.logger.Sync()
	defer s.stanConn.Close()
	s.logger.Info(listenMsg, zap.String("Port", s.Port))
	s.logger.Fatal(errInterrupt, zap.Error(srv.ListenAndServe()))
}

func (s *server) initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/vote", s.createVote).Methods(http.MethodPost)
	return router
}

func (s *server) createVote(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		vote    pb.Vote
		stsCode = http.StatusCreated
	)
	defer func() {
		defer s.logger.Info(voteCreateMsg, zap.Error(err), zap.Int32("electionId", vote.ElectionId), zap.String("User", vote.User), zap.Int("StatusCode", stsCode))
	}()

	err = json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		stsCode = http.StatusBadRequest
		http.Error(w, errInvalidData, stsCode)
		return
	}
	if vote.GetElectionId() <= 0 {
		stsCode = http.StatusBadRequest
		http.Error(w, errInvalidID, stsCode)
		return
	}
	if vote.GetUser() == "" {
		stsCode = http.StatusBadRequest
		http.Error(w, errInvalidUser, stsCode)
		return
	}

	err = s.publishEvent(&vote)
	if err != nil {
		stsCode = http.StatusInternalServerError
		http.Error(w, errFailPubVote, stsCode)
		return
	}

	w.WriteHeader(stsCode)
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
	s.run()
}
