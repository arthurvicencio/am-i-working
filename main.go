package main

import (
	"database/sql"
	"fmt"
	"strings"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Log struct {
	Name     string
	Duration int
}

type LogSummary struct {
	Name string
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
		   started DATETIME DEFAULT CURRENT_TIMESTAMP,
		   ended DATETIME
		);
	`)

	currentWindow := ""
	for {
		if window := GetWindow(); currentWindow != window {

			db.Exec(`
				UPDATE log
				SET ended = CURRENT_TIMESTAMP
				WHERE id = (
					SELECT MAX(id)
					FROM log
				);
			`)
			db.Exec(
				fmt.Sprintf(`INSERT INTO log(name) VALUES ("%s")`, window),
			)

			currentWindow = window
		}
	}
}

func getLogs() []Log {
	logs := make([]Log, 0)

	stmt, err := db.Prepare(`
		SELECT
			name,
			SUM(
				strftime("%s", IFNULL(ended, CURRENT_TIMESTAMP)) -
				strftime("%s",started)
			)
		FROM log
		GROUP BY name
		ORDER BY name;
	`)
	if err != nil {
		return logs
	}

	defer stmt.Close()

	rows, _ := stmt.Query()

	for rows.Next() {

		var log Log

		err = rows.Scan(
			&log.Name,
			&log.Duration,
		)
		if err != nil {
			return nil
		}

		logs = append(logs, log)
	}

	return logs
}

func secondsToTimeFormat(seconds int) string {
	text := []string{"00","00","00"}

	s := seconds / (60*24)
	if s > 0 {
		text[0] = fmt.Sprintf("%02d", s)
	}
	seconds = seconds % (60*24)

	s = seconds / 60
	if s > 0 {
		text[1] = fmt.Sprintf("%02d", s)
	}
	seconds = seconds % 60

	if seconds > 0 {
		text[2] = fmt.Sprintf("%02d", seconds)
	}

	return strings.Join(text, ":")
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("index.html")


		logSummary := make([]LogSummary, 0)
		logs := getLogs()

		for _, lg := range logs {
			logSummary = append(logSummary, LogSummary{
				Name: lg.Name,
				TimeSpent: secondsToTimeFormat(lg.Duration),
			})
		}

		t.Execute(w, logSummary)
	})

	go logWindow()

	log.Printf("Serving %s on HTTP port: %s\n", ".", "1234")
	log.Fatal(http.ListenAndServe(":1234", nil))
}
