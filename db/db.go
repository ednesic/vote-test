package db

import (
	mgo "gopkg.in/mgo.v2"
)

type DataAccessLayer interface {
	Insert(collectionName string, docs interface{}) error
	FindOne(collName string, query interface{}, doc interface{}) error
	Update(collName string, selector interface{}, update interface{}) error
	Upsert(collName string, selector interface{}, update interface{}) error
	Remove(collName string, selector interface{}) error
}

type MongoDAL struct {
	session *mgo.Session
	dbName  string
}

// NewMongoDAL creates a MongoDAL
func NewMongoDAL(dbURI string, dbName string) (DataAccessLayer, error) {
	session, err := mgo.Dial(dbURI)
	mongo := &MongoDAL{
		session: session,
		dbName:  dbName,
	}
	return mongo, err
}

// Insert stores documents in mongo
func (m *MongoDAL) Insert(collName string, doc interface{}) error {
	session := m.session.Clone()
	defer session.Close()
	return session.DB(m.dbName).C(collName).Insert(doc)
}

// FindOne finds one document in mongo
func (m *MongoDAL) FindOne(collName string, query interface{}, doc interface{}) error {
	session := m.session.Clone()
	defer session.Close()
	return session.DB(m.dbName).C(collName).Find(query).One(doc)
}

func (m *MongoDAL) Update(collName string, selector interface{}, update interface{}) error {
	session := m.session.Clone()
	defer session.Close()
	return session.DB(m.dbName).C(collName).Update(selector, update)
}

func (m *MongoDAL) Upsert(collName string, selector interface{}, update interface{}) error {
	session := m.session.Clone()
	defer session.Close()
	_, err := session.DB(m.dbName).C(collName).Upsert(selector, update)
	return err
}

func (m *MongoDAL) Remove(collName string, selector interface{}) error {
	session := m.session.Clone()
	defer session.Close()
	err := session.DB(m.dbName).C(collName).Remove(selector)
	return err
}
