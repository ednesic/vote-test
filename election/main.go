package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/pb"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	errEnvVarFail          = "Failed to get environment variables:"
	errInvalidID           = "Invalid Id"
	errInvalidElectionData = "Invalid election Data"
	errInvalidVoteData     = "Invalid Vote Data"
	errInvalidMgoSession   = "Failed to retrieve mongo session"
	errParseTimestamp      = "Failed to parse timestamp"
	errConn                = "Failed connect to mongodb"
	errRetrieveQuery       = "Failed to retrieve query"
	errNotFound            = "Not found election"
)

type server struct {
	Port       string `envconfig:"PORT" default:"9223"`
	Database   string `envconfig:"DATABASE" default:"elections"`
	Collection string `envconfig:"COLLECTION" default:"election"`
	MgoURL     string `envconfig:"MONGO_URL" default:"localhost:27017"`

	mgoDal db.DataAccessLayer
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

	log.Println("HTTP Sever listening on " + s.Port)
	log.Fatal(s.srv.ListenAndServe())
}

func (s *server) initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/election", s.createElection).Methods("PUT")
	router.HandleFunc("/election/{id}", s.getElection).Methods("GET")
	router.HandleFunc("/election/{id}", s.getElection).Methods("DELETE")
	return router
}

func (s *server) createElection(w http.ResponseWriter, r *http.Request) {
	var election pb.Election
	if json.NewDecoder(r.Body).Decode(&election) != nil {
		http.Error(w, errInvalidElectionData, http.StatusBadRequest)
		return
	}

	if election.Id == 0 {
		http.Error(w, errInvalidElectionData, http.StatusBadRequest)
	}

	election.Inicio = ptypes.TimestampNow()
	if s.mgoDal.Upsert(s.Collection, bson.M{"id": election.Id}, &election) != nil {
		http.Error(w, "Failed to insert/update election", 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	j, _ := json.Marshal(election)
	w.Write(j)
}

func (s *server) getElection(w http.ResponseWriter, r *http.Request) {
	var election pb.Election
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}

	if err := s.mgoDal.FindOne(s.Collection, bson.M{"id": id}, &election); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, errNotFound, http.StatusNotFound)
			return
		}
		http.Error(w, errRetrieveQuery, 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	j, _ := json.Marshal(election)
	w.Write(j)
}

func (s *server) deleteElection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}

	if err := s.mgoDal.Remove(s.Collection, bson.M{"id": id}); err != nil {
		http.Error(w, errRetrieveQuery, 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	var s server
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.run()
}
