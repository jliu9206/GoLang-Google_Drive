package mysql

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// init
func Initialize(username, password, address, dbname string) {
	//format dsn string
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, address, dbname)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error Opening databse")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error Pinging databse")
	}

	fmt.Println("Successfully connected to database!")
}

// DBConn: Expose sql.DB pointer
func DBConn() *sql.DB {
	return db
}

// DBClose: Close db connection
func DBClose() {
	if db != nil {
		db.Close()
	}
}
