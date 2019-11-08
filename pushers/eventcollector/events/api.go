package events

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.GET("/sessions", epSessions)
	r.GET("/sessions/ssh", epSessionsSSH)
	r.GET("/session/:session-id", epSession)

	r.GET("/events", epEvents)
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

func epSessions(c *gin.Context) {
	log.Debugf("ENDPOINT /sessions (size: %v)", len(Sessions))

	// return json from slice sessions values
	c.JSON(http.StatusOK, getSessionsValues())
}

func epSessionsSSH(c *gin.Context) {

	var sshSessions []Session
	sessions := getSessionsValues()
	for _, s := range sessions {

		if s.Service == "ssh" {
			sshSessions = append(sshSessions, s)
		}
	}
	log.Debugf("ENDPOINT /sessions/ssh (size: %v)", len(sshSessions))

	c.JSON(http.StatusOK, sshSessions)
}

func epSession(c *gin.Context) {
	sessionID := c.Params.ByName("session-id")
	log.Debugf("ENDPOINT /session/%v", sessionID)
	if session, ok := Sessions[sessionID]; !ok {
		c.JSON(http.StatusOK, gin.H{})
		return
	} else {
		c.JSON(http.StatusOK, session)
	}
}

func epEvents(c *gin.Context) {
	log.Debugf("ENDPOINT /events (size: %v)", len(Events))
	c.JSON(http.StatusOK, Events)
}

func StartAPI() {
	r := setupRouter()
	// Listen and Serve in 0.0.0.0:8080
	r.Run(":8080")
}