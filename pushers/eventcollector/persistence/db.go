package persistence

import (
	"github.com/op/go-logging"
	"gopkg.in/mgo.v2"
	"os"
)

var log = logging.MustGetLogger("channels/eventcollector/persistence")

type DBConnection struct {
	//Session stores mongo session
	Session *mgo.Session

	// Mongo stores the mongodb connection string information
	Mongo *mgo.DialInfo
}

const (
	// MongoDBUrl is the default mongodb url that will be used to connect to the
	// database.
	MongoDBUri = "mongodb://localhost:27017/honeytrap"
)

func Connect() (conn *DBConnection) {
	uri := os.Getenv("MONGODB_HOST")
	if len(uri) == 0 {
		uri = MongoDBUri
	}

	mongo, err := mgo.ParseURL(uri)
	session, err := mgo.Dial(uri)
	if err != nil {
		log.Errorf("Can't connect to mongo: %v", err)
		panic(err.Error())
	}

	session.SetSafe(&mgo.Safe{})
	session.SetMode(mgo.Monotonic, true)

	log.Info("Connected to mongo on %v", uri)

	conn = &DBConnection{
		Session: session,
		Mongo: mongo,
	}

	return conn
}

func (conn *DBConnection) Use(tableName string) (collection *mgo.Collection) {
	return conn.Session.DB("honeytrap").C(tableName)
}

func (conn *DBConnection) Close() {
	conn.Session.Close()
	return
}