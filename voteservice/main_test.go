package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ednesic/vote-test/pb"
	"github.com/ednesic/vote-test/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func Test_server_publishEvent(t *testing.T) {
	log, _ := zap.NewProduction()
	stanMock := new(tests.StanConnMock)
	stanMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	type args struct {
		vote *pb.Vote
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Empty vote", args{&pb.Vote{}}, false},
		{"Empty user on vote", args{&pb.Vote{Candidate: "", ElectionId: 12}}, false},
		{"Empty id on vote", args{&pb.Vote{Candidate: "1", ElectionId: 0}}, false},
		{"Normal vote", args{&pb.Vote{Candidate: "1", ElectionId: 90}}, false},
		{"Negative id on vote", args{&pb.Vote{Candidate: "1", ElectionId: -1}}, false},
		{"High value on vote", args{&pb.Vote{Candidate: "1", ElectionId: 1000000000}}, false},
		{"Big string on vote", args{&pb.Vote{Candidate: "sdgnasodgsdgsino", ElectionId: 1}}, false},
		{"Nil vote must fail", args{nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{stanConn: stanMock, logger: log}
			if err := s.publishEvent(tt.args.vote); (err != nil) != tt.wantErr {
				t.Errorf("server.publishEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_server_createVote(t *testing.T) {
	log, _ := zap.NewProduction()

	stanMock := new(tests.StanConnMock)
	tests := []struct {
		name         string
		body         string
		statusCode   int
		responseBody string
		pubRes       error
	}{
		{"Could not process message", `{"electionId":12,"candidate":"abc"}`, http.StatusInternalServerError, errFailPubVote, errors.New("err")},
		{"Creation successful", `{"electionId":12,"candidate":"abc"}`, http.StatusCreated, `{"ElectionId":12,"candidate":"abc"}`, nil},
		{"Wrong user type", `{"electionId":"12"}`, http.StatusBadRequest, errInvalidData, nil},
		{"Wrong id type", `{"candidate":12}`, http.StatusBadRequest, errInvalidData, nil},
		{"Missing user", `{"electionId":12}`, http.StatusBadRequest, errInvalidUser, nil},
		{"Missing id", `{"candidate":"abc"}`, http.StatusBadRequest, errInvalidID, nil},
		{"Missing all", `{}`, http.StatusBadRequest, errInvalidID, nil},
		{"Wrong parameters", `{"electionId":12,"candidate":"abc","Home: 5"}`, http.StatusBadRequest, errInvalidData, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stanMock.On("Publish", mock.Anything, mock.Anything).Return(tt.pubRes).Once()
			s := &server{
				stanConn: stanMock,
				logger:   log,
			}
			req, err := http.NewRequest("POST", "localhost:9222/vote", strings.NewReader(tt.body))
			assert.Nil(t, err, "could not create request")

			rec := httptest.NewRecorder()
			s.createVote(rec, req)
			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.statusCode, res.StatusCode, "Did not get the same response code")
			assert.Equal(t, tt.responseBody, strings.TrimSuffix(rec.Body.String(), "\n"))
		})
	}
}

func Test_server_initRoutes(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Return non nil object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{}
			assert.NotNil(t, s.initRoutes())
		})
	}
}
