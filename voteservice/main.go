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

const (
	ErrEnvVarFail int = iota + 1
	ErrConnLost
	ErrFailPubVote
	ErrInvalidData
)

var errCodeToMessage = map[int]string{
	ErrEnvVarFail:  "Failed to get environment variables:",
	ErrConnLost:    "Connection lost:",
	ErrFailPubVote: "Failed to publish vote",
	ErrInvalidData: "Invalid Vote Data",
}

// Variaveis de ambiente
type server struct {
	Port          string `envconfig:"PORT" default:"9222"`
	NatsClusterID string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	VoteChannel   string `envconfig:"VOTE_CHANNEL" default:"create-vote"`
	NatsServer    string `envconfig:"NATS_SERVER" default:"localhost:4222"`
	ClientID      string `envconfig:"CLIENT_ID" default:"vote-service"`

	srv     *http.Server
	strmCmp *natsutil.StreamingComponent
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
			log.Fatal(errCodeToMessage[ErrConnLost], reason)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

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

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, errCodeToMessage[ErrInvalidData], 500)
		return
	}

	err = s.publishEvent(s.strmCmp, &vote)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println(errCodeToMessage[ErrFailPubVote], err)
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
	return sc.Publish(s.VoteChannel, voteJSON)
}

func main() {
	var s server
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(errCodeToMessage[ErrEnvVarFail], err.Error())
	}
	s.run()
}
