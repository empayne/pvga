package main

import (
	"github.com/empayne/pvga/db"
	"github.com/empayne/pvga/router"
	_ "github.com/lib/pq"
)

func main() {
	db := db.InitDatabase()
	router := router.CreateRouter(db)
	router.Run()
}
