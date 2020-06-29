package models

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	CollectionSession = "sessions"
)

type Session struct {
	SessionID 		string 			`json:"session-id" form:"session-id" binding:"required" bson:"session-id"`
	Service 		string 			`json:"service" form:"service" binding:"required" bson:"service"`
	SourceIP 		string 			`json:"source-ip" form:"source-ip" binding:"required" bson:"source-ip"`
	SourcePort 		uint 			`json:"source-port" form:"source-port" binding:"required" bson:"source-port"`
	DestinationIP 	string 			`json:"destination-ip" form:"destination-ip" binding:"required" bson:"destination-ip"`
	DestinationPort uint 			`json:"destination-port" form:"destination-port" binding:"required" bson:"destination-port"`
	CreationDate 	string 			`json:"creation-date" form:"creation-date" binding:"required" bson:"creation-date"`
	UpdateDate 		string			`json:"update-date" form:"update-date" binding:"required" bson:"update-date"`
	EventCount		uint			`json:"event-count" form:"event-count" binding:"required" bson:"event-count"`
	ServiceMeta 	interface{}		`json:"service-meta" form:"service-meta" binding:"required" bson:"service-meta"`
}

type SessionSSH struct {
	Token			string					`json:"token" form:"token" binding:"required" bson:"token"`
	AuthAttempts	[]SessionSSHAuth		`json:"auth-attempts" form:"auth-attempts" binding:"required" bson:"auth-attempts"`
	AuthSuccess		bool					`json:"auth-success" form:"auth-success" binding:"required" bson:"auth-success"`
	AuthFailCount	uint					`json:"auth-fail-count" form:"auth-fail-count" binding:"required" bson:"auth-fail-count"`
	Payload			string					`json:"payload" form:"payload" binding:"required" bson:"payload"`
	Recording		[]SessionSSHRecording	`json:"recording" form:"recording" binding:"required" bson:"recording"`
}

type SessionSSHAuth struct {
	Timestamp 		string 		`json:"timestamp" form:"timestamp" binding:"required" bson:"timestamp"`
	AuthType 		string 		`json:"auth-type" form:"auth-type" binding:"required" bson:"auth-type"`
	Username 		string 		`json:"username" form:"username" binding:"required" bson:"username"`
	Password 		string 		`json:"password" form:"password" binding:"required" bson:"password"`
	PublicKey 		string 		`json:"public-key" form:"public-key" binding:"required" bson:"public-key"`
	PublicKeyType 	string 		`json:"public-key-type" form:"public-key-type" binding:"required" bson:"public-key-type"`
	Success 		bool 		`json:"success" form:"success" binding:"required" bson:"success"`
}

type SessionSSHRecording struct {
	Index 		    int		`json:"index" form:"index" bson:"index"`
	Command 		string 	    `json:"command" form:"command" bson:"command"`
	Output          string      `json:"output" form:"output" bson:"output"`
}

type SessionTelnet struct {
	Token			string			     	`json:"token" form:"token" binding:"required" bson:"token"`
	AuthAttempts	[]SessionTelnetAuth		`json:"auth-attempts" form:"auth-attempts" binding:"required" bson:"auth-attempts"`
	AuthSuccess		bool					`json:"auth-success" form:"auth-success" binding:"required" bson:"auth-success"`
	AuthFailCount	uint					`json:"auth-fail-count" form:"auth-fail-count" binding:"required" bson:"auth-fail-count"`
	Commands		[]SessionTelnetCommand	`json:"commands" form:"commands" binding:"required" bson:"commands"`
	Payload			string					`json:"payload" form:"payload" binding:"required" bson:"payload"`
	Recording		string					`json:"recording" form:"recording" binding:"required" bson:"recording"`
}

type SessionTelnetAuth struct {
	Timestamp 		string 		`json:"timestamp" form:"timestamp" binding:"required" bson:"timestamp"`
	AuthType 		string 		`json:"auth-type" form:"auth-type" binding:"required" bson:"auth-type"`
	Username 		string 		`json:"username" form:"username" binding:"required" bson:"username"`
	Password 		string 		`json:"password" form:"password" binding:"required" bson:"password"`
	Success 		bool 		`json:"success" form:"success" binding:"required" bson:"success"`
}

type SessionTelnetCommand struct {
	Command 		string		`json:"command" form:"command" binding:"required" bson:"command"`
	Timestamp 		string 		`json:"timestamp" form:"timestamp" binding:"required" bson:"timestamp"`
}

type SessionModel struct{}

func (s *SessionModel) Create(data Session) error {
	collection := db.Use("sessions")
	err := collection.Insert(data)
	return err
}

func (s *SessionModel) Update(sessionID string, data Session) error {
	collection := db.Use("sessions")
	err := collection.Update(bson.D{{"session-id", sessionID}}, data)
	return err
}

func (s *SessionModel) Get(id string) (session Session, err error) {
	collection := db.Use("sessions")
	err = collection.FindId(bson.ObjectIdHex(id)).One(&session)
	return session, err
}

func (s *SessionModel) Find() (list []Session, err error) {
	collection := db.Use("sessions")
	err = collection.Find(bson.M{}).All(&list)
	return list, err
}

func (s *SessionModel) DropAll() error {
	collection := db.Use("sessions")
	err := collection.DropCollection()
	return err
}