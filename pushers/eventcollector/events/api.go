package events

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"

)

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	//CORS - specific parameters
/*	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://honeynet.ubiwhere.com"},
		AllowMethods:     []string{"PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))*/

	//CORS -> allow all origins
	r.Use(cors.Default())


	r.GET("/sessions", epSessionsFind)
	r.GET("/sessions/purge", epSessionsPrune)
	r.GET("/session/:session-id", epSessions)
	r.GET("/events", epEventsFind)
	r.GET("/event/:event-id", epEventGet)

	//r.GET("/ssh/sessions", endpointSSHSessions)
	//r.GET("/ssh/session/:session_id", endpointSSHSessions)

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}



func epSessionsFind(c *gin.Context) {
	//page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	//limit, _ := strconv.Atoi(c.DefaultQuery("limit", "3"))


	list, err := sessionModel.Find()
	log.Debugf("ENDPOINT /sessions (size: %v)", len(list))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Find error", "error": err.Error()})
		c.Abort()
	} else {
		c.JSON(http.StatusOK, gin.H{"sessions": list})
	}
}

/*
func epSessionsSSH(c *gin.Context) {
	// var sshSessions []Session
	sshSessions := make([]models.Session, 0, 0)
	sessions := getSessionsValues()
	for _, s := range sessions {
		if s.Service == "ssh" {
			sshSessions = append(sshSessions, s)go g
		}
	}
	log.Debugf("ENDPOINT /sessions/ssh (size: %v)", len(sshSessions))
	c.JSON(http.StatusOK, sshSessions)
}

func epSessionsTelnet(c *gin.Context) {
	telnetSessions := make([]models.Session, 0, 0)
	sessions := getSessionsValues()
	for _, s := range sessions {
		if s.Service == "telnet" {
			telnetSessions = append(telnetSessions, s)
		}
	}
	log.Debugf("ENDPOINT /sessions/telnet (size: %v)", len(telnetSessions))
	c.JSON(http.StatusOK, telnetSessions)
}
*/


func epSessions(c *gin.Context) {
	sessionID := c.Params.ByName("session-id")
	log.Debugf("ENDPOINT /sessions/%v", sessionID)
	if session, ok := Sessions[sessionID]; !ok {
		c.JSON(http.StatusOK, gin.H{})
		return
	} else {
		c.JSON(http.StatusOK, session)
	}
}

func epSessionsPrune(c *gin.Context) {
	log.Debugf("ENDPOINT /sessions/prune")
	err := sessionModel.DropAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message:": "Couldn't remove all sessions", "error": err})
		c.Abort()
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "All sessions were removed successfully"})
	}

}

func epEventsFind(c *gin.Context) {
	list, err := eventModel.Find()
	log.Debugf("ENDPOINT /events (size: %v)", len(list))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Find error", "error": err.Error()})
		c.Abort()
	} else {
		c.JSON(http.StatusOK, gin.H{"data": list})
	}
}

func epEventGet(c *gin.Context) {
	eventID := c.Params.ByName("event-id")
	log.Debugf("ENDPOINT /events/%v", eventID)
	event, err := eventModel.Get(eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Event not found", "error": err.Error()})
		c.Abort()
	} else {
		c.JSON(http.StatusOK, gin.H{"data": event})
	}
}

func epEvent(c *gin.Context) {
	eventID := c.Params.ByName("event-id")
	log.Debugf("ENDPOINT /event/%v", eventID)

	if event, ok := Events[eventID]; !ok {
		c.JSON(http.StatusOK, gin.H{})
		return
	} else {
		c.JSON(http.StatusOK, event)
	}
}

func StartAPI() {
	r := setupRouter()
	// Listen and Serve in 0.0.0.0:8080
	r.Run(":8080")
}