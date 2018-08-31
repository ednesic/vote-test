package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

type server struct {
	Port string `envconfig:"PORT" default:"9223"`

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
	router.HandleFunc("/election", s.getElection).Methods("GET")
	router.HandleFunc("/election", s.deleteElection).Methods("DELETE")
	return router
}

func (s *server) createElection(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func (s *server) getElection(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func (s *server) deleteElection(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func main() {
	var s server
	err := envconfig.Process("", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.run()
}
