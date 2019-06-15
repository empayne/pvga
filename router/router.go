package router

import (
	"log"
	"net/http"

	"github.com/empayne/pvga/db"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// Router olds the Gin routing engine, database connection, and config info.
type Router struct {
	engine *gin.Engine
	db     *db.Database
}

// Inspired by: https://stackoverflow.com/questions/34046194/
func useDatabase(conn *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("databaseConn", conn)
		c.Next()
	}
}

// Inspired by: https://stackoverflow.com/questions/34046194/
func getDatabaseConnection(c *gin.Context) *db.Database {
	dbConn, ok := c.MustGet("databaseConn").(*db.Database)
	if !ok {
		log.Fatal("Could not use database in handler.")
	}
	return dbConn
}

// Inspired by: https://github.com/Depado/gin-auth-example/blob/master/main.go
func checkAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			c.Redirect(301, "/")
		} else {
			c.Next() // Continue down the chain to handler etc
		}
	}
}

func setErrorOnContext(c *gin.Context, err error) {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	// TODO: how to make this work for pq errors?
	tracer := err.(error).(stackTracer)

	// OWASP Top 10 2017 #6: Security Misconfiguration
	// We shouldn't send a stack trace in an error message, but if DEBUG is set
	// in our environment, this information will be provided to an attacker.
	// See 'db.go' for more information.
	c.JSON(http.StatusInternalServerError, gin.H{"Error": err, "StackTrace": tracer.StackTrace()})
}

// CreateRouter sets up paths / stores database and config info.
func CreateRouter(db *db.Database) *Router {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	store := sessions.NewCookieStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	router.Use(useDatabase(db))

	router.GET("/", index)
	router.POST("/login", login)

	authRequired := router.Group("/app", checkAuth())
	authRequired.GET("/", app)
	authRequired.GET("/leaderboard", leaderboard)
	authRequired.GET("/profile", profile)
	authRequired.POST("/click", click)
	authRequired.POST("/reset", reset)
	authRequired.POST("/update_profile", updateProfile)

	return &Router{
		engine: router,
		db:     db,
	}
}

// Run the router we initialized in CreateRouter.
func (r *Router) Run() {
	r.engine.Run()
}
