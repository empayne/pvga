package db

import (
	"database/sql"
	"log"
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

// InitDatabase reads environment variables, and then opens our connection to pg.
func InitDatabase() *Database {
	// TODO: de-hardcode
	conn, err := sql.Open("postgres", "postgres://postgres:postgres@localhost/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err) // kill server if we can't use DB on startup
	}
	return &Database{
		conn: conn,
	}
}

// Assumes that there's only _one_ row and _one_ user.
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

	if len(u) > 0 {
		return u[0], nil
	}

	return nil, nil
}

// ReadUserByID queries the top-level model by ID (used everywhere but login).
func (db *Database) ReadUserByID(id string) (*User, error) {
	// TODO: refactor so SELECT ... FROM ... isn't repeated in three places
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
	_, err := db.conn.Exec(`
		UPDATE users
		SET clicks = clicks + $1
		WHERE id = $2
	`, count, id)

	return err
}

// UpdateClicks sets the user's click count to a specific value. Used by our
// import save data functionality. Should be paird with UpdateLastClick.
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
	rows, err := db.conn.Query(`
		SELECT id, username, email, bio, password, clicks, last_click, is_admin
		FROM users
		ORDER BY clicks DESC
		LIMIT $1
	`, userCount)

	if err != nil {
		log.Fatal(err)
	}

	return readUsersFromRows(rows)
}

// UpdateBio updates the user's bio
func (db *Database) UpdateBio(id string, bio string) error {
	stmt := "UPDATE users SET bio = '" + bio + "' WHERE id = '" + id + "'"

	_, err := db.conn.Exec(stmt)

	// Return an error containing a stacktrace here to demonstrate #6: Security
	// Misconfiguration. See 'router.go' for more information.
	return errors.Wrap(err, "")
}
