package db

import (
	mgo "gopkg.in/mgo.v2"
)

var mgoSession *mgo.Session

// Creates a new session if mgoSession is nil i.e there is no active mongo session.
//If there is an active mongo session it will return a Clone
func GetMongoSession(url string) (*mgo.Session, error) {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial(url)
		if err != nil {
			return nil, err
		}
	}
	return mgoSession.Clone(), nil
}
