package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Conn struct {
	Conn *sql.DB
}

var DbConn *Conn

func Connect() *Conn {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return &Conn{
		Conn: db,
	}
}

func Init() {
	DbConn = Connect()
}
