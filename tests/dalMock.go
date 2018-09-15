package tests

import "github.com/stretchr/testify/mock"

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
