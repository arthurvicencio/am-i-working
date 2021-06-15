package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Log struct {
	Name      string
	TimeSpent string
}

var db *sql.DB

func init() {
	db, _ = sql.Open("sqlite3", "./log.db")
}

func logWindow() {

	db.Exec(`
		CREATE TABLE IF NOT EXISTS log (
		   id INTEGER PRIMARY KEY AUTOINCREMENT,
		   name VARCHAR(500),
		   created DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)

	currentWindow := ""
	for {
		if window := GetWindow(); currentWindow != window {

			db.Exec(
				fmt.Sprintf(`INSERT INTO log(name) VALUES ("%s")`, window),
			)

			currentWindow = window
		}
	}
}

func getLogs() []Log {
	logs := make([]Log, 0)

	stmt, err := db.Prepare("SELECT name FROM log;")
	if err != nil {
		return logs
	}

	defer stmt.Close()

	rows, _ := stmt.Query()

	for rows.Next() {

		var log Log

		err = rows.Scan(
			&log.Name,
		)
		if err != nil {
			return nil
		}

		logs = append(logs, log)
	}

	return logs
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("index.html")

		t.Execute(w, getLogs())
	})

	go logWindow()

	log.Printf("Serving %s on HTTP port: %s\n", ".", "1234")
	log.Fatal(http.ListenAndServe(":1234", nil))
}
