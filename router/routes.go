package router

import (
	"net/http"
	"strconv"

	"github.com/empayne/redundantserializer"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

/*
	Page-rendering requests:
*/

func index(c *gin.Context) {
	params := c.Request.URL.Query()
	var username string

	// TODO: is this necessary?
	if len(params["username"]) > 0 {
		username = params["username"][0]
	} else {
		username = ""
	}

	c.HTML(
		http.StatusOK,
		"index.html",
		gin.H{"username": username},
	)
}

func app(c *gin.Context) {
	session := sessions.Default(c)
	db, err := getDatabaseConnection(c)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}
	id := session.Get("user").(string)

	u, err := db.ReadUserByID(id)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	c.HTML(
		http.StatusOK,
		"app.html",
		gin.H{"id": u.ID, "username": u.Username, "clicks": u.Clicks},
	)
}

func leaderboard(c *gin.Context) {
	session := sessions.Default(c)
	db, err := getDatabaseConnection(c)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}
	id := session.Get("user").(string)

	u, err := db.ReadUserByID(id)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	// TODO: make the '5' parameter (ie. scoreboard size) configurable
	leaders, err := db.ReadUsersByClicksDescending(5)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

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
	db, err := getDatabaseConnection(c)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}
	id := session.Get("user").(string)

	u, err := db.ReadUserByID(id)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

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
	db, err := getDatabaseConnection(c)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")

	u, err := db.ReadUserByUsername(username)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

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
		// OWASP Top 10 2017 #2: Broken Authentication
		// We do nothing to limit the number of failed login attempts, so we can
		// try to bruteforce a user's password via a wordlist of known passwords
		// (eg. rockyou.txt).
		//
		// We should be limiting the number of login attempts to prevent a
		// bruteforce attack (eg. increasing delays between failed attempts,
		// locking an account after N failed attempts).

		// TODO: I think that blindly concatenating URL parameters here is
		// safe in this case (eg. not an open redirect), but double-check this.
		c.Redirect(301, "/?username="+username)
	}
}

func click(c *gin.Context) {
	session := sessions.Default(c)
	db, err := getDatabaseConnection(c)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}
	id := session.Get("user").(string)

	// TODO: put these two db operations in a single transaction
	err = db.IncrementClicks(id, 1)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	err = db.UpdateLastClick(id)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	c.JSON(200, nil)
}

func reset(c *gin.Context) {
	db, err := getDatabaseConnection(c)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}
	id := c.PostForm("id")

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

	// TODO: put these two db operations in a single transaction
	err = db.ResetClicks(id)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	err = db.UpdateLastClick(id)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	c.JSON(200, nil)
}

// We only support updating the bio for now
func updateProfile(c *gin.Context) {
	session := sessions.Default(c)
	db, err := getDatabaseConnection(c)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	id := session.Get("user").(string)
	bio := c.PostForm("bio")

	err = db.UpdateBio(id, bio)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	c.JSON(200, nil)
}

/*
	Import / export functions:
*/

// Create 'save data' containing the user's bio and score using
// redundantserializer library, then send it back to the user via UI.
func exportData(c *gin.Context) {
	session := sessions.Default(c)
	db, err := getDatabaseConnection(c)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	id := session.Get("user").(string)

	// Contains bio and score to export
	u, err := db.ReadUserByID(id)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	// TODO: get rid of new 'score' terminology, it should just be 'clicks'
	serializableMap := redundantserializer.SerializableMap{
		"bio":   u.Bio,
		"score": strconv.Itoa(u.Clicks),
	}
	saveData, err := redundantserializer.Serialize(serializableMap)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	c.JSON(200, gin.H{"Data": saveData})
}

// User has sent the 'save data' back to us via UI. Get the original bio and
// score back from redundantserializer, then update the DB accordingly.
func importData(c *gin.Context) {
	session := sessions.Default(c)
	db, err := getDatabaseConnection(c)

	id := session.Get("user").(string)
	saveData := c.PostForm("savedata")

	deserializedMap, err := redundantserializer.Deserialize(&saveData)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	bio, score, err := readSaveData(deserializedMap)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	err = db.UpdateBio(id, *bio)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	err = db.UpdateClicks(id, *score)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	err = db.UpdateLastClick(id)
	if err != nil {
		setErrorOnContext(c, err)
		return
	}

	c.JSON(200, nil)
}
