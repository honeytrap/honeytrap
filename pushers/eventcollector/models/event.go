package models

import (
	"gopkg.in/mgo.v2/bson"
)

const (
	CollectionEvent = "events"
)

type Event struct {
	EventID 		bson.ObjectId   `json:"event_id" bson:"_id,omitempty"`
	//event_id = generate
	AgentType 		string  		`json:"agent_type" form:"agent_type" binding:"required" bson:"agent_type"`
	AgentID			string			`json:"agent_id" form:"agent_id" binding:"required" bson:"agent_id"`
	Timestamp 		string 			`json:"timestamp" form:"timestamp" binding:"required" bson:"timestamp"`
	SourceIP 		string 			`json:"sourceip" form:"sourceip" binding:"required" bson:"sourceip"`
	Count 			uint 			`json:"count" form:"count" binding:"required" bson:"count"`
	//count = 0
	Type 			string 			`json:"type" form:"type" binding:"required" bson:"type"`
	Priority 		string 			`json:"priority" form:"priority" binding:"required" bson:"priority"`
	//priority = medium
	Name 			string 			`json:"name" form:"name" binding:"required" bson:"name"`
	Context 		string 			`json:"context" form:"context" binding:"required" bson:"context"`
	Metadata 		interface{} 	`json:"metadata" form:"metadata" bson:"metadata"`
}

type EventMetadataSSH struct {
	EventType 		string 		`json:"event_type" form:"event_type" binding:"required" bson:"event_type"`
	SourcePort 		uint 		`json:"source_port" form:"source_port" binding:"required" bson:"source_port"`
	DestinationIP 	string 		`json:"dest_ip" form:"dest_ip" binding:"required" bson:"dest_ip"`
	DestinationPort uint 		`json:"dest_port" form:"dest_port" binding:"required" bson:"dest_port"`
	SessionID 		string 		`json:"session_id" form:"session_id" binding:"required" bson:"session_id"`
	Username 		string 		`json:"username" form:"username" binding:"required" bson:"username"`
	Token 			string 		`json:"token" form:"token" binding:"required" bson:"token"`
	TransactionType string 		`json:"transaction_type" form:"transaction_type" binding:"required" bson:"transaction_type"`
	PublicKey 		string 		`json:"public_key" form:"public_key" binding:"required" bson:"public_key"`
	PublicKeyType 	string 		`json:"public_key_type" form:"public_key_type" binding:"required" bson:"public_key_type"`
	Password 		string 		`json:"password" form:"password" binding:"required" bson:"password"`
	Recording 		string 		`json:"recording" form:"recording" binding:"required" bson:"recording"`
	Open            bool        `json:"open" form:"open" binding:"required" bson:"open"`
}

type EventMetadataTelnet struct {
	EventType 		string 		`json:"event_type" form:"event_type" binding:"required" bson:"event_type"`
	SourcePort 		uint 		`json:"source_port" form:"source_port" binding:"required" bson:"source_port"`
	DestinationIP 	string 		`json:"dest_ip" form:"dest_ip" binding:"required" bson:"dest_ip"`
	DestinationPort uint 		`json:"dest_port" form:"dest_port" binding:"required" bson:"dest_port"`
	SessionID 		string 		`json:"session_id" form:"session_id" binding:"required" bson:"session_id"`
	Username 		string 		`json:"username" form:"username" binding:"required" bson:"username"`
	Token 			string 		`json:"token" form:"token" binding:"required" bson:"required"`
	TransactionType string 		`json:"transaction_type" form:"transaction_type" binding:"required" bson:"transaction_type"`
	Password 		string 		`json:"password" form:"password" binding:"required" bson:"password"`
	Recording 		string 		`json:"recording" form:"recording" binding:"required" bson:"recording"`
}

type EventModel struct{}


func (e *EventModel) Create(data Event) error {
	collection := db.Use("events")
	err := collection.Insert(data)
	return err
}

func (e* EventModel) Get(id string) (event Event, err error) {
	collection := db.Use("events")
	err = collection.FindId(bson.ObjectIdHex(id)).One(&event)
	return event, err
}

func (e *EventModel) Find() (list []Event, err error) {
	collection := db.Use("events")
	err = collection.Find(bson.M{}).All(&list)
	return list, err
}

func (e *EventModel) DropAll() error {
	collection := db.Use("events")
	err := collection.DropCollection()
	return err
}