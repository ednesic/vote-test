package db

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	mgo "gopkg.in/mgo.v2"
)

const (
	ErrEnvVarFail int = iota + 1
)

var errCodeToMessage = map[int]string{
	ErrEnvVarFail: "Failed to get environment variables:",
}

var mgoSession *mgo.Session

type mongoConn struct {
	URL  string `envconfig:"MONGO_URL" default:"localhost"`
	Port string `envconfig:"MONGO_PORT" default:"27017"`
}

// Creates a new session if mgoSession is nil i.e there is no active mongo session.
//If there is an active mongo session it will return a Clone
func GetMongoSession() (*mgo.Session, error) {
	var mgoConn mongoConn
	err := envconfig.Process("", &mgoConn)
	if err != nil {
		log.Fatal(errCodeToMessage[ErrEnvVarFail], err.Error())
	}
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial(mgoConn.URL + ":" + mgoConn.Port)
		if err != nil {
			return nil, err
		}
	}
	return mgoSession.Clone(), nil
}
