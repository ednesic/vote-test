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
	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	errEnvVarFail    = "Failed to get environment variables:"
	errInvalidID     = "Invalid Id"
	errConnFail      = "Connection failed"
	errInvalidData   = "Invalid election Data"
	errConn          = "Failed connect to mongodb"
	errRetrieveQuery = "Failed to retrieve query"
	errNotFound      = "Not found election"
	errUpsert        = "Failed to insert/update election"
	errInterrupt     = "Shutting down"
	errEnsureIndex   = "Error in err Ensurance"

	listenMsg   = "HTTP Sever listening"
	serviceName = "election"

	elecIDKey  = "id"
	stsCodeKey = "StatusCode"
)

type server struct {
	Port       string `envconfig:"PORT" default:"9223"`
	Database   string `envconfig:"DATABASE" default:"elections"`
	Collection string `envconfig:"COLLECTION" default:"election"`
	MgoURL     string `envconfig:"MONGO_URL" default:"localhost:27017"`

	mgoDal db.DataAccessLayer
	logger *zap.Logger
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

	s.mgoDal, err = db.NewMongoDAL(s.MgoURL, s.Database)
	if err != nil {
		s.logger.Fatal(errConnFail, zap.Error(err))
	}

	err = s.mgoDal.EnsureIndex(s.Collection, elecIDKey)
	if err != nil {
		s.logger.Fatal(errEnsureIndex, zap.Error(err))
	}

	defer s.logger.Sync()
	s.logger.Info(listenMsg, zap.String("Port", s.Port))
	s.logger.Fatal(errInterrupt, zap.Error(srv.ListenAndServe()))
}

func (s *server) initRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/"+serviceName, s.upsert).Methods(http.MethodPut)
	router.HandleFunc("/"+serviceName+"/{"+elecIDKey+"}", s.get).Methods(http.MethodGet)
	router.HandleFunc("/"+serviceName+"/{"+elecIDKey+"}", s.delete).Methods(http.MethodDelete)
	return router
}

func (s *server) upsert(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		election pb.Election
		stsCode  = http.StatusCreated
	)
	defer func() {
		defer s.logger.Info(http.MethodPut+serviceName, zap.Error(err), zap.Int32(elecIDKey, election.Id), zap.Int(stsCodeKey, stsCode))
	}()

	err = json.NewDecoder(r.Body).Decode(&election)
	if err != nil {
		stsCode = http.StatusBadRequest
		http.Error(w, errInvalidData, stsCode)
		return
	}

	if election.Id == 0 {
		stsCode = http.StatusBadRequest
		http.Error(w, errInvalidData, stsCode)
		return
	}

	election.Inicio = ptypes.TimestampNow()
	err = s.mgoDal.Upsert(s.Collection, bson.M{elecIDKey: election.Id}, &election)
	if err != nil {
		stsCode = http.StatusInternalServerError
		http.Error(w, errUpsert, stsCode)
		return
	}

	w.WriteHeader(stsCode)
	j, _ := json.Marshal(election)
	w.Write(j)
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		election pb.Election
		stsCode  = http.StatusOK
		vars     = mux.Vars(r)
		id       int64
	)
	defer func() {
		defer s.logger.Info(http.MethodGet+serviceName, zap.Error(err), zap.Int32(elecIDKey, election.Id), zap.Int(stsCodeKey, stsCode))
	}()

	id, err = strconv.ParseInt(vars[elecIDKey], 10, 32)
	if err != nil {
		stsCode = http.StatusBadRequest
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}

	err = s.mgoDal.FindOne(s.Collection, bson.M{elecIDKey: id}, &election)
	if err != nil {
		if err == mgo.ErrNotFound {
			stsCode = http.StatusNotFound
			http.Error(w, errNotFound, http.StatusNotFound)
			return
		}
		stsCode = http.StatusInternalServerError
		http.Error(w, errRetrieveQuery, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(stsCode)
	j, _ := json.Marshal(election)
	w.Write(j)
}

func (s *server) delete(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		election pb.Election
		stsCode  = http.StatusOK
		vars     = mux.Vars(r)
		id       int64
	)
	defer func() {
		defer s.logger.Info(http.MethodDelete+serviceName, zap.Error(err), zap.Int32(elecIDKey, election.Id), zap.Int(stsCodeKey, stsCode))
	}()

	id, err = strconv.ParseInt(vars[elecIDKey], 10, 32)
	if err != nil {
		stsCode = http.StatusBadRequest
		http.Error(w, errInvalidID, http.StatusBadRequest)
		return
	}

	err = s.mgoDal.Remove(s.Collection, bson.M{elecIDKey: id})
	if err != nil {
		stsCode = http.StatusInternalServerError
		http.Error(w, errRetrieveQuery, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(stsCode)
}

func main() {
	var s server
	s.run()
}
