// package config

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"

// 	_ "github.com/lib/pq"
// )

// var DB *sql.DB

// func ConnectDB() {

// 	psqlInfo := fmt.Sprintf(
// 		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		os.Getenv("DB_HOST"),
// 		os.Getenv("DB_PORT"),
// 		os.Getenv("DB_USER"),
// 		os.Getenv("DB_PASSWORD"),
// 		os.Getenv("DB_NAME"),
// 	)

// 	db, err := sql.Open("postgres", psqlInfo)

// 	if err != nil {
// 		log.Fatal("failed to connect database:", err)
// 	}

// 	err = db.Ping()

// 	if err != nil {
// 		log.Fatal("database unreachable:", err)
// 	}

// 	DB = db

//		fmt.Println("database connected")
//	}
package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("database unreachable:", err)
	}

	DB = db

	log.Println("database connected successfully")
}
