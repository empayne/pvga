package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

/*
	Page-rendering requests:
*/

func index(c *gin.Context) {
	// OWASP Top 10 2017 #5: Broken Access Control
	params := c.Request.URL.Query()
	fmt.Println(params)
	var username string

	if len(params["username"]) > 0 {
		username = params["username"][0]
	} else {
		username = ""
	}

	fmt.Println(username)
	c.HTML(
		http.StatusOK,
		"index.html",
		gin.H{"username": username},
	)
}

func app(c *gin.Context) {
	session := sessions.Default(c)
	db := getDatabaseConnection(c)

	id := session.Get("user").(string)
	u := db.ReadUserByID(id)

	c.HTML(
		http.StatusOK,
		"app.html",
		gin.H{"id": u.ID, "username": u.Username, "clicks": u.Clicks},
	)
}

func leaderboard(c *gin.Context) {
	session := sessions.Default(c)
	db := getDatabaseConnection(c)

	id := session.Get("user").(string)
	u := db.ReadUserByID(id)
	leaders := db.ReadUsersByClicksDescending(5)

	// This is pretty dangerous - we pass in the entire user object into
	// our template. We only expose the properties we directly access in the
	// template, but we're one typo away from exposing other users' email /
	// password info to the client...
	// TODO: use an intermediate struct that only contains the data we're
	// interested in.
	c.HTML(
		http.StatusOK,
		"leaderboard.html",
		gin.H{"leaders": leaders, "isAdmin": u.IsAdmin},
	)
}

func profile(c *gin.Context) {
	session := sessions.Default(c)
	db := getDatabaseConnection(c)

	id := session.Get("user").(string)
	u := db.ReadUserByID(id)

	c.HTML(
		http.StatusOK,
		"profile.html",
		gin.H{"user": u},
	)
}

/*
	Non-rendering requests:
*/

// Inspired by: https://github.com/Depado/gin-auth-example/blob/master/main.go
func login(c *gin.Context) {
	session := sessions.Default(c)
	db := getDatabaseConnection(c)

	username := c.PostForm("username")
	password := c.PostForm("password")

	u := db.ReadUserByUsername(username)
	valid := u != nil &&
		len(u.Username) > 0 &&
		username == u.Username &&
		password == u.Password

	if valid {
		session.Set("user", u.ID)
		err := session.Save()
		if err != nil {
			c.Redirect(301, "/")
		} else {
			c.Redirect(301, "/app")
		}
	} else {
		// TODO: I think that blindly concatenating URL parameters here is
		// safe in this case (eg. not an open redirect), but double-check this.
		c.Redirect(301, "/?username="+username)
	}
}

func click(c *gin.Context) {
	session := sessions.Default(c)
	db := getDatabaseConnection(c)

	id := session.Get("user").(string)
	db.IncrementClicks(id, 1)
	db.UpdateLastClick(id)

	c.JSON(200, nil)
}

func reset(c *gin.Context) {
	db := getDatabaseConnection(c)

	// OWASP Top 10 2017 #5: Broken Access Control
	// We blindly trust that the ID in the POST form hasn't been modified and
	// matches the value that we rendered. As a result, any user can reset any
	// other user's score (provided they know their ID).

	// We can't simply use the session ID, though, since an admin can reset any
	// user's score. The correct implementation would be:
	// - get the user's ID via session
	// - read the user from database via ID
	// - if admin == true: use the id from the POST form (can reset any user)
	// - if admin == false: use the id from the session, ignore POST param

	id := c.PostForm("id")
	db.ResetClicks(id)
	db.UpdateLastClick(id)

	c.JSON(200, nil)
}

func updateProfile(c *gin.Context) {
	session := sessions.Default(c)
	db := getDatabaseConnection(c)

	id := session.Get("user").(string)
	bio := c.PostForm("bio")

	// We only support updating the bio for now
	db.UpdateBio(id, bio)

	c.JSON(200, nil)
}
