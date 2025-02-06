package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "postgres"
	DB_NAME     = "ping_db"
)

type PingResult struct {
	IPAddress      string `json:"ip_address"`
	PingTime       string `json:"ping_time"`
	LastSuccessful string `json:"last_successful"`
}

var db *sql.DB

func initDB() {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ping_results (
		id SERIAL PRIMARY KEY,
		ip_address VARCHAR(50) NOT NULL,
		ping_time TIMESTAMP NOT NULL,
		last_successful TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}
}

func getPingResults(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT ip_address, ping_time, last_successful FROM ping_results")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []PingResult
	for rows.Next() {
		var result PingResult
		if err := rows.Scan(&result.IPAddress, &result.PingTime, &result.LastSuccessful); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func addPingResult(w http.ResponseWriter, r *http.Request) {
	var result PingResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Некорректные данные запроса", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO ping_results (ip_address, ping_time, last_successful) VALUES ($1, $2, $3)", result.IPAddress, result.PingTime, result.LastSuccessful)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/ping_results", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPingResults(w, r)
		case http.MethodPost:
			addPingResult(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
