package expenses

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB
var err error

func InitDB(url string) {
	// InitDB()
	// var err error
	url = os.Getenv("DATABASE_URL")
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Fatal("can't connect to database", err)
	}
	defer db.Close()
	CreateTable()
}

func CreateTable() {
	// CreateTable()
	createTable := `CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT, 
		amount FLOAT, 
		note TEXT, 
		tags TEXT[]
		);`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal("can't create table", err)
	}
}
