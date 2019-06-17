package router

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/empayne/redundantserializer"

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
func getDatabaseConnection(c *gin.Context) (*db.Database, error) {
	dbConn, ok := c.MustGet("databaseConn").(*db.Database)
	if !ok {
		return nil, errors.New("Could not use database in handler")
	}
	return dbConn, nil
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

	errorJSON := gin.H{"Error": err.Error()}
	_, ok := err.(error).(stackTracer)

	// OWASP Top 10 2017 #6: Security Misconfiguration
	// We shouldn't send a stack trace in an error message, but if DEBUG is set
	// in our environment, this information will be provided to an attacker.
	// See UpdateBio in 'db.go' for a sample error that returns a stack trace.
	//
	// We should have a more robust method to stop stack traces from getting
	// into production (eg. don't just check that a custom environment variable
	// is defined).
	if ok && len(os.Getenv("DEBUG")) > 0 {
		tracer := err.(error).(stackTracer)
		// TODO: concatenating strings in a for loop is suboptimal, fix that
		stackTraceString := ""
		// StackTrace handling from https://godoc.org/github.com/pkg/errors
		for _, f := range tracer.StackTrace() {
			stackTraceString = stackTraceString + fmt.Sprintf("%+s:%d\n", f, f)
		}

		errorJSON["StackTrace"] = stackTraceString
	}

	c.JSON(http.StatusInternalServerError, errorJSON)
}

func readSaveData(deserializedMap redundantserializer.SerializableMap) (*string, *int, error) {
	bio, okBio := deserializedMap["bio"]
	scoreStr, okScore := deserializedMap["score"]
	if !(okBio && okScore) {
		return nil, nil, errors.New("Could not read bio and score from save data")
	}

	score, err := strconv.Atoi(scoreStr)
	if err != nil {
		return nil, nil, err
	}

	return &bio, &score, nil
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
	authRequired.GET("/export", exportData)
	authRequired.POST("/import", importData)

	return &Router{
		engine: router,
		db:     db,
	}
}

// Run the router we initialized in CreateRouter.
func (r *Router) Run() {
	r.engine.Run()
}
