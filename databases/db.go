package databases

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
	var err error

	dsn := getDataSourceName()

	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	err = DB.Ping()
	if err != nil {
		panic("Failed to ping database: " + err.Error())
	}

	createTables()
	createIndexOnTable()
}

func getDataSourceName() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s",
		user,
		password,
		host,
		port,
		dbName,
	)
}
