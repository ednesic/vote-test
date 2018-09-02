package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ednesic/vote-test/db"
	"github.com/ednesic/vote-test/pb"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	errEnvVarFail        = "Failed to get environment variables:"
	errInvalidData       = "Invalid Vote Data"
	errInvalidMgoSession = "Failed to retrieve mongo session"
	errParseTimestamp    = "Failed to parse timestamp"
	errConn              = "Failed connect to mongodb"
	errRetrieveQuery     = "Failed to retrieve query"
	errNotFound          = "Not found election"
)

type server struct {
	Port       string `envconfig:"PORT" default:"9223"`
	Database   string `envconfig:"DATABASE" default:"elections"`
	Collection string `envconfig:"COLLECTION" default:"election"`

	srv *http.Server
}

func (s *server) run() {
	s.srv = &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.initRoutes(),
	}
	log.Println("HTTP Sever listening on " + s.Port)
	log.Fatal(s.srv.ListenAndServe())
}

func (s *server) initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/election", s.createElection).Methods("POST")
	router.HandleFunc("/election/{id}", s.getElection).Methods("GET")
	return router
}

func (s *server) createElection(w http.ResponseWriter, r *http.Request) {
	var end int64

	if json.NewDecoder(r.Body).Decode(&end) != nil {
		http.Error(w, errInvalidData, 500)
		return
	}

	session, err := db.GetMongoSession()
	if err != nil {
		http.Error(w, errInvalidMgoSession, 500)
		return
	}
	defer session.Close()
	conn := session.DB(s.Database).C(s.Collection)

	num, err := conn.Find(bson.M{}).Count()
	if err != nil {
		http.Error(w, errRetrieveQuery, 500)
		return
	}

	tmstp, err := ptypes.TimestampProto(time.Unix(end, 0))
	if err != nil {
		http.Error(w, errParseTimestamp, 500)
		return
	}

	election := &pb.Election{
		Id:      int32(num),
		Inicio:  ptypes.TimestampNow(),
		Termino: tmstp,
	}

	if conn.Insert(&election) != nil {
		http.Error(w, errConn, 500)
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
		http.Error(w, errInvalidData, 500)
		return
	}

	session, err := db.GetMongoSession()
	if err != nil {
		http.Error(w, errInvalidMgoSession, 500)
		return
	}
	defer session.Close()
	conn := session.DB(s.Database).C(s.Collection)

	if err = conn.Find(bson.M{"id": id}).One(&election); err != nil {
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

func main() {
	var s server
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.run()
}
