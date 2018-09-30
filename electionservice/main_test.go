package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ednesic/vote-test/tests"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"
)

func Test_server_upsert(t *testing.T) {
	mgoDal := &tests.DataAccessLayerMock{}
	log, _ := zap.NewProduction()

	tests := []struct {
		name       string
		body       string
		statusCode int
		queryRet   error
	}{
		{"Create Election", `{"id": 3, "candidates": ["test1", "test2"],"end": { "seconds": 1536525322 }}`, http.StatusCreated, nil},
		{"Upsert function do not work", `{"id": 3, "candidates": ["test1", "test2"],"end": { "seconds": 1536525322 }}`, http.StatusInternalServerError, errors.New("Upsert fail")},
		{"Id = 0", `{"id": 0, "candidates": ["test1", "test2"], "end": { "seconds": 1536525322 }}`, http.StatusBadRequest, nil},
		{"Without candidates", `{"id": 4, "candidates": [], "end": { "seconds": 1536525322 }}`, http.StatusBadRequest, nil},
		{"Empty candidates", `{"id": 4, "end": { "seconds": 1536525322 }}`, http.StatusBadRequest, nil},
		{"Invalid payload", ``, http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				mgoDal: mgoDal,
				logger: log,
			}
			mgoDal.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(tt.queryRet).Once()

			req, err := http.NewRequest("PUT", "localhost:9223/election", strings.NewReader(tt.body))
			assert.Nil(t, err, "could not create request")
			rec := httptest.NewRecorder()
			s.upsert(rec, req)
			res := rec.Result()
			defer res.Body.Close()
			fmt.Println(rec.Body.String())

			assert.Equal(t, tt.statusCode, res.StatusCode, "Did not get the same response code")
		})
	}
}

func Test_server_valid(t *testing.T) {
	mgoDal := &tests.DataAccessLayerMock{}
	log, _ := zap.NewProduction()

	isOverRetFalse := func(end *timestamp.Timestamp) bool {
		return false
	}
	isOverRetTrue := func(end *timestamp.Timestamp) bool {
		return true
	}
	containsCandidateRetTrue := func(candidate string, candidates []string) bool {
		return true
	}
	containsCandidateRetFalse := func(candidate string, candidates []string) bool {
		return false
	}

	tests := []struct {
		name                  string
		ID                    string
		statusCode            int
		candidate             string
		isOverMock            func(end *timestamp.Timestamp) bool
		containsCandidateMock func(candidate string, candidates []string) bool
		queryRet              error
	}{
		{"Get Election Ok", "1", http.StatusOK, "", isOverRetFalse, containsCandidateRetTrue, nil},
		{"Get Election over", "1", http.StatusGone, "", isOverRetTrue, containsCandidateRetTrue, nil},
		{"Get Election query with wrong candidate", "1", http.StatusBadRequest, "mock1", isOverRetFalse, containsCandidateRetFalse, nil},
		{"Find fail(not found)", "1", http.StatusNotFound, "", isOverRetFalse, containsCandidateRetTrue, mgo.ErrNotFound},
		{"Find fail", "1", http.StatusInternalServerError, "", isOverRetFalse, containsCandidateRetTrue, errors.New("test error")},
		{"Id != int", "test", http.StatusBadRequest, "", isOverRetFalse, containsCandidateRetTrue, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				mgoDal:            mgoDal,
				logger:            log,
				isOver:            tt.isOverMock,
				containsCandidate: tt.containsCandidateMock,
			}

			mgoDal.On("FindOne", mock.Anything, mock.Anything, mock.Anything).Return(tt.queryRet).Once()

			req, err := http.NewRequest("GET", "localhost:9223/election/validate?candidate="+tt.candidate, nil)
			assert.Nil(t, err, "could not create request")
			rec := httptest.NewRecorder()

			vars := map[string]string{
				"id": tt.ID,
			}
			req = mux.SetURLVars(req, vars)

			s.valid(rec, req)
			res := rec.Result()
			defer res.Body.Close()
			fmt.Println(rec.Body.String())

			assert.Equal(t, tt.statusCode, res.StatusCode, "Did not get the same response code")
		})
	}
}

func Test_server_delete(t *testing.T) {
	mgoDal := &tests.DataAccessLayerMock{}
	log, _ := zap.NewProduction()

	tests := []struct {
		name       string
		ID         string
		statusCode int
		queryRet   error
	}{
		{"Delete Election Ok", "1", http.StatusOK, nil},
		{"Delete fail", "1", http.StatusInternalServerError, errors.New("test error")},
		{"Id != int", "test", http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				mgoDal: mgoDal,
				logger: log,
			}

			mgoDal.On("Remove", mock.Anything, mock.Anything).Return(tt.queryRet).Once()

			req, err := http.NewRequest("Delete", "localhost:9223/election/", nil)
			assert.Nil(t, err, "could not create request")
			rec := httptest.NewRecorder()

			vars := map[string]string{
				"id": tt.ID,
			}
			req = mux.SetURLVars(req, vars)

			s.delete(rec, req)
			res := rec.Result()
			defer res.Body.Close()
			fmt.Println(rec.Body.String())

			assert.Equal(t, tt.statusCode, res.StatusCode, "Did not get the same response code")
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

func Test_isOver(t *testing.T) {
	tbefore := ptypes.TimestampNow()
	tafter := ptypes.TimestampNow()
	tinvalid := ptypes.TimestampNow()

	tbefore.Seconds = tbefore.GetSeconds() - 1000
	tafter.Seconds = tafter.GetSeconds() + 1000
	tinvalid.Seconds = -1000
	tinvalid.Nanos = -1000

	type args struct {
		end *timestamp.Timestamp
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Time before now", args{end: tbefore}, true},
		{"Time after now", args{end: tafter}, false},
		{"Invalid time", args{end: tinvalid}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOver(tt.args.end); got != tt.want {
				t.Errorf("isElectionOver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_containsCandidate(t *testing.T) {
	type args struct {
		candidate  string
		candidates []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Contains", args{candidate: "test1", candidates: []string{"test1", "test2"}}, true},
		{"Not Contains", args{candidate: "test3", candidates: []string{"test1", "test2"}}, false},
		{"Empty candidates", args{candidate: "test3", candidates: nil}, false},
		{"Empty candidate", args{candidate: "", candidates: []string{"test1", "test2"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsCandidate(tt.args.candidate, tt.args.candidates); got != tt.want {
				t.Errorf("containsCandidate() = %v, want %v", got, tt.want)
			}
		})
	}
}
