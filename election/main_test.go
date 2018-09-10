package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mgo "gopkg.in/mgo.v2"
)

type DataAccessLayerMock struct {
	mock.Mock
}

func (m *DataAccessLayerMock) Insert(collName string, doc interface{}) error {
	args := m.Called(collName, doc)
	return args.Error(0)
}
func (m *DataAccessLayerMock) FindOne(collName string, query interface{}, doc interface{}) error {
	args := m.Called(collName, query, doc)
	return args.Error(0)
}

func (m *DataAccessLayerMock) Update(collName string, selector interface{}, update interface{}) error {
	args := m.Called(collName, selector, update)
	return args.Error(0)
}

func (m *DataAccessLayerMock) Upsert(collName string, selector interface{}, update interface{}) error {
	args := m.Called(collName, selector, update)
	return args.Error(0)
}

func (m *DataAccessLayerMock) Remove(collName string, selector interface{}) error {
	args := m.Called(collName, selector)
	return args.Error(0)
}

func Test_server_upsertElection(t *testing.T) {
	mgoDal := &DataAccessLayerMock{}

	tests := []struct {
		name       string
		body       string
		statusCode int
		queryRet   error
	}{
		{"Create Election", `{"id": 3, "termino": { "seconds": 1536525322 }}`, http.StatusCreated, nil},
		{"Upsert function do not work", `{"id": 3, "termino": { "seconds": 1536525322 }}`, http.StatusUnprocessableEntity, errors.New("Upsert fail")},
		{"Id = 0", `{"id": 0, "termino": { "seconds": 1536525322 }}`, http.StatusBadRequest, nil},
		{"Invalid payload", ``, http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				mgoDal: mgoDal,
			}
			mgoDal.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(tt.queryRet).Once()

			req, err := http.NewRequest("PUT", "localhost:9223/election", strings.NewReader(tt.body))
			assert.Nil(t, err, "could not create request")
			rec := httptest.NewRecorder()
			s.upsertElection(rec, req)
			res := rec.Result()
			defer res.Body.Close()
			fmt.Println(rec.Body.String())

			assert.Equal(t, tt.statusCode, res.StatusCode, "Did not get the same response code")
		})
	}
}

func Test_server_getElection(t *testing.T) {
	mgoDal := &DataAccessLayerMock{}

	tests := []struct {
		name       string
		ID         string
		statusCode int
		queryRet   error
	}{
		{"Get Election Ok", "1", http.StatusOK, nil},
		{"Find fail(not found)", "1", http.StatusNotFound, mgo.ErrNotFound},
		{"Find fail", "1", http.StatusUnprocessableEntity, errors.New("test error")},
		{"Id != int", "test", http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				mgoDal: mgoDal,
			}

			mgoDal.On("FindOne", mock.Anything, mock.Anything, mock.Anything).Return(tt.queryRet).Once()

			req, err := http.NewRequest("GET", "localhost:9223/election/", nil)
			assert.Nil(t, err, "could not create request")
			rec := httptest.NewRecorder()

			vars := map[string]string{
				"id": tt.ID,
			}
			req = mux.SetURLVars(req, vars)

			s.getElection(rec, req)
			res := rec.Result()
			defer res.Body.Close()
			fmt.Println(rec.Body.String())

			assert.Equal(t, tt.statusCode, res.StatusCode, "Did not get the same response code")
		})
	}
}

func Test_server_deleteElection(t *testing.T) {
	mgoDal := &DataAccessLayerMock{}

	tests := []struct {
		name       string
		ID         string
		statusCode int
		queryRet   error
	}{
		{"Delete Election Ok", "1", http.StatusOK, nil},
		{"Delete fail", "1", http.StatusUnprocessableEntity, errors.New("test error")},
		{"Id != int", "test", http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &server{
				mgoDal: mgoDal,
			}

			mgoDal.On("Remove", mock.Anything, mock.Anything).Return(tt.queryRet).Once()

			req, err := http.NewRequest("Delete", "localhost:9223/election/", nil)
			assert.Nil(t, err, "could not create request")
			rec := httptest.NewRecorder()

			vars := map[string]string{
				"id": tt.ID,
			}
			req = mux.SetURLVars(req, vars)

			s.deleteElection(rec, req)
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

/*******************************************/
/** Tests that must have mongod installed **/
/*******************************************/
