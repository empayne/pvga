package main

import (
	"github.com/empayne/pvga/db"
	"github.com/empayne/pvga/router"
	_ "github.com/lib/pq"
)

// OWASP Top 10 2017 #10: Insufficient Logging & Monitoring
// We didn't explicitly add any logging statements to our code, so we can only
// rely on Gin's default logging behaviour. If a breach does occur, it would be
// difficult to detect it via our logs.
//
// We should add logging statements to the pvga codebase/

func main() {
	db := db.InitDatabase()
	router := router.CreateRouter(db)
	router.Run()
}
