package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/pb"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	errInvalidID           = "Invalid Id"
	errInvalidElectionData = "Invalid election Data"
	errConn                = "Failed connect to mongodb"
	errRetrieveQuery       = "Failed to retrieve query"
	errNotFound            = "Not found election"
	errUpsert              = "Failed to insert/update election"
)

type server struct {
	Port       string `envconfig:"PORT" default:"9223"`
	Database   string `envconfig:"DATABASE" default:"elections"`
	Collection string `envconfig:"COLLECTION" default:"election"`
	MgoURL     string `envconfig:"MONGO_URL" default:"localhost:27017"`

	mgoDal db.DataAccessLayer
	logger *zap.Logger
	srv    *http.Server
}

func (s *server) run() {
	var err error
	s.srv = &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.initRoutes(),
	}

	s.mgoDal, err = db.NewMongoDAL(s.MgoURL, s.Database)
	if err != nil {
		log.Fatal(errConn, err)
	}

	s.logger, err = zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	defer s.logger.Sync()
	s.logger.Info("HTTP Sever listening", zap.String("Port", s.Port), zap.String("hostname", os.Getenv("HOSTNAME")))
	log.Fatal(s.srv.ListenAndServe())
}

func (s *server) initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/election", s.upsertElection).Methods("PUT")
	router.HandleFunc("/election/{id}", s.getElection).Methods("GET")
	router.HandleFunc("/election/{id}", s.deleteElection).Methods("DELETE")
	return router
}

func (s *server) upsertElection(w http.ResponseWriter, r *http.Request) {
	var election pb.Election
	if json.NewDecoder(r.Body).Decode(&election) != nil {
		http.Error(w, errInvalidElectionData, http.StatusBadRequest)
		defer s.logger.Error(errInvalidElectionData, zap.Int("StatusCode", http.StatusBadRequest), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	if election.Id == 0 {
		http.Error(w, errInvalidElectionData, http.StatusBadRequest)
		defer s.logger.Error(errInvalidElectionData, zap.Int32("Id", election.Id), zap.Int("StatusCode", http.StatusBadRequest), zap.String("hostname", os.Getenv("HOSTNAME")))
	}

	election.Inicio = ptypes.TimestampNow()
	if s.mgoDal.Upsert(s.Collection, bson.M{"id": election.Id}, &election) != nil {
		http.Error(w, errUpsert, http.StatusInternalServerError)
		defer s.logger.Error(errUpsert, zap.Int32("Id", election.Id), zap.Int("StatusCode", http.StatusInternalServerError), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	w.WriteHeader(http.StatusCreated)
	j, _ := json.Marshal(election)
	w.Write(j)
	defer s.logger.Info("Election Created", zap.Any("election", election), zap.Int("StatusCode", http.StatusCreated), zap.String("hostname", os.Getenv("HOSTNAME")))
}

func (s *server) getElection(w http.ResponseWriter, r *http.Request) {
	var election pb.Election
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		defer s.logger.Error(errInvalidID, zap.Int64("Id", id), zap.Error(err), zap.Int("StatusCode", http.StatusBadRequest), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	if err := s.mgoDal.FindOne(s.Collection, bson.M{"id": id}, &election); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, errNotFound, http.StatusNotFound)
			defer s.logger.Error(errNotFound, zap.Int64("Id", id), zap.Error(err), zap.Int("StatusCode", http.StatusNotFound), zap.String("hostname", os.Getenv("HOSTNAME")))
			return
		}
		http.Error(w, errRetrieveQuery, http.StatusInternalServerError)
		defer s.logger.Error(errRetrieveQuery, zap.Int64("Id", id), zap.Error(err), zap.Int("StatusCode", http.StatusInternalServerError), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	w.WriteHeader(http.StatusOK)
	j, _ := json.Marshal(election)
	w.Write(j)
	defer s.logger.Info("Got election", zap.Int64("Id", id), zap.Int("StatusCode", http.StatusOK), zap.String("hostname", os.Getenv("HOSTNAME")))
}

func (s *server) deleteElection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		defer s.logger.Error(errInvalidID, zap.Int64("Id", id), zap.Error(err), zap.Int("StatusCode", http.StatusBadRequest), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	if err := s.mgoDal.Remove(s.Collection, bson.M{"id": id}); err != nil {
		http.Error(w, errRetrieveQuery, http.StatusInternalServerError)
		defer s.logger.Error(errRetrieveQuery, zap.Int64("Id", id), zap.Error(err), zap.Int("StatusCode", http.StatusInternalServerError), zap.String("hostname", os.Getenv("HOSTNAME")))
		return
	}

	w.WriteHeader(http.StatusOK)
	defer s.logger.Info("Deleted election", zap.Int64("Id", id), zap.Int("StatusCode", http.StatusOK), zap.String("hostname", os.Getenv("HOSTNAME")))
}

func main() {
	var s server
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.run()
}
