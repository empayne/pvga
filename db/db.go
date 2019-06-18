package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
)

// Database hold our connection to SQL / has our db operations defined on it
type Database struct {
	conn *sql.DB
}

// User is a top-level model mapping to our 'users' table.
// TODO: move this to a 'models' package.
type User struct {
	ID        string
	Username  string
	Email     string
	Bio       string
	Password  string
	Clicks    int
	LastClick time.Time
	IsAdmin   bool
}

// InitDatabase reads in a connection string from the environment, and then
// opens our connection to pg.
func InitDatabase() *Database {
	// eg. "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
	// TODO: enable SSL on DB
	conn, err := sql.Open("postgres", os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err) // kill server if we can't use DB on startup
	}
	return &Database{
		conn: conn,
	}
}

// Read one or more user from rows.
func readUsersFromRows(rows *sql.Rows) ([]*User, error) {
	var users []*User

	for rows.Next() {
		u := User{}
		err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.Bio,
			&u.Password,
			&u.Clicks,
			&u.LastClick,
			&u.IsAdmin,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, &u)
	}

	return users, nil
}

// ReadUserByUsername queries the top-level model by username (used for login).
func (db *Database) ReadUserByUsername(username string) (*User, error) {
	// TODO: refactor so SELECT ... FROM ... isn't repeated in three places
	// (ReadUserByUsername, ReadUserByID, ReadUsersByClicksDescending)
	rows, err := db.conn.Query(`
		SELECT id, username, email, bio, password, clicks, last_click, is_admin
		FROM users
		WHERE username = $1
		LIMIT 1
	`, username)

	if err != nil {
		return nil, err
	}

	u, err := readUsersFromRows(rows)

	if err != nil {
		return nil, err
	}

	if len(u) > 0 { // user exists
		return u[0], nil
	}

	return nil, nil // user doesn't exist
}

// ReadUserByID queries the top-level model by ID (used everywhere but login).
func (db *Database) ReadUserByID(id string) (*User, error) {
	// TODO: refactor so SELECT ... FROM ... isn't repeated in three places
	// (ReadUserByUsername, ReadUserByID, ReadUsersByClicksDescending)
	rows, err := db.conn.Query(`
		SELECT id, username, email, bio, password, clicks, last_click, is_admin
		FROM users
		WHERE id = $1
		LIMIT 1
	`, id)

	if err != nil {
		return nil, err
	}

	users, err := readUsersFromRows(rows) // should only contain one user
	return users[0], err
}

// IncrementClicks will update the user's click count in database.
// This will be called a lot. Should be paired with UpdateLastClick.
func (db *Database) IncrementClicks(id string, count int) error {
	// TODO: now that we have both IncrementClicks and UpdateClicks, should
	// this be refactored such that IncrementClicks doesn't need to exist?
	_, err := db.conn.Exec(`
		UPDATE users
		SET clicks = clicks + $1
		WHERE id = $2
	`, count, id)

	return err
}

// UpdateClicks sets the user's click count to a specific value. Used by our
// import save data functionality. Should be paired with UpdateLastClick.
func (db *Database) UpdateClicks(id string, count int) error {
	_, err := db.conn.Exec(`
		UPDATE users
		SET clicks = $1
		WHERE id = $2
	`, count, id)

	return err
}

// UpdateLastClick updates the 'last click' timestamp in database.
// This should be called along with every call to IncrementClicks.
func (db *Database) UpdateLastClick(id string) error {
	_, err := db.conn.Exec(`
		UPDATE users
		SET last_click = CURRENT_TIMESTAMP
		WHERE id = $1
	`, id)

	return err
}

// ResetClicks sets the user's click count to 0.
func (db *Database) ResetClicks(id string) error {
	_, err := db.conn.Exec(`
		UPDATE users
		SET clicks = 0
		WHERE id = $1
	`, id)

	return err
}

// ReadUsersByClicksDescending is used to construct the leaderboard. Read
// the users with the top N clicks.
func (db *Database) ReadUsersByClicksDescending(userCount int) ([]*User, error) {
	// TODO: refactor so SELECT ... FROM ... isn't repeated in three places
	// (ReadUserByUsername, ReadUserByID, ReadUsersByClicksDescending)
	rows, err := db.conn.Query(`
		SELECT id, username, email, bio, password, clicks, last_click, is_admin
		FROM users
		ORDER BY clicks DESC
		LIMIT $1
	`, userCount)

	if err != nil {
		return nil, err
	}

	return readUsersFromRows(rows)
}

// UpdateBio updates the user's bio
func (db *Database) UpdateBio(id string, bio string) error {
	// OWASP Top 10 2017 #1: Injection
	// We construct our SQL query via string concatenation with untrusted user
	// input, so we're vulnerable to a SQL injection. For example, if bio is:
	//
	// "' WHERE 0=1; DROP TABLE users; --"
	//
	// then the attacker can drop out table, as the query is expanded to:
	//
	// UPDATE users SET bio = '' WHERE 0=1; DROP TABLE users; --' WHERE id = ...
	//
	// The attacker can make arbitrary modifications to our database, as well as
	// read arbitrary DB contents via their own 'bio' field.
	//
	// To avoid injection, we should be using setting parameters in Golang's
	// "Query" function (see other DB functions in 'db.go'), rather than
	// constructing SQL queries via string concatenation.

	stmt := "UPDATE users SET bio = '" + bio + "' WHERE id = '" + id + "'"

	_, err := db.conn.Exec(stmt)

	// Return an error containing a stacktrace here to demonstrate #6: Security
	// Misconfiguration. See 'router.go' for more information.
	return errors.Wrap(err, "")
}
