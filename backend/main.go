package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const (
    DB_USER     = "postgres"
    DB_PASSWORD = "admin"
    DB_HOST     = "postgres"
    DB_PORT     = "5432"
    DB_NAME     = "ping_db"
)

type PingResult struct {
	IPAddress      string  `json:"ip_address"`
	PingTime       float64 `json:"ping_time"`
	LastSuccessful string  `json:"last_successful"`
}

var db *sql.DB

func initDB() {
	const maxAttempts = 5
	const retryInterval = 2 * time.Second

	adminUser := getEnv("DB_USER", "postgres")
	adminPassword := getEnv("DB_PASSWORD", "admin")
	dbName := getEnv("DB_NAME", "ping_db")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432"), adminUser, adminPassword)

	var dbTemp *sql.DB
	var err error

	// Подключение с несколькими попытками
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		dbTemp, err = sql.Open("postgres", connStr)
		if err == nil {
			err = dbTemp.Ping()
		}
		if err == nil {
			break
		}
		log.Printf("Попытка %d подключения к PostgreSQL не удалась: %v", attempts, err)
		time.Sleep(retryInterval)
	}

	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer dbTemp.Close()

	// Проверка существования базы данных
	var exists bool
	err = dbTemp.QueryRow(fmt.Sprintf("SELECT EXISTS (SELECT FROM pg_database WHERE datname = '%s')", dbName)).Scan(&exists)
	if err != nil {
		log.Fatalf("Ошибка проверки базы данных: %v", err)
	}

	if !exists {
		_, err = dbTemp.Exec(fmt.Sprintf("CREATE DATABASE %s OWNER %s", dbName, adminUser))
		if err != nil {
			log.Fatalf("Ошибка создания базы данных: %v", err)
		}
		log.Printf("База данных %s создана.", dbName)
	}

	connStrWithDB := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432"), adminUser, adminPassword, dbName)

	for attempts := 1; attempts <= maxAttempts; attempts++ {
		db, err = sql.Open("postgres", connStrWithDB)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			break
		}
		log.Printf("Попытка %d подключения к базе данных не удалась: %v", attempts, err)
		time.Sleep(retryInterval)
	}

	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	// Создание таблицы при отсутствии
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS ping_results (
		id SERIAL PRIMARY KEY,
		ip_address VARCHAR(50) NOT NULL,
		ping_time INT NOT NULL,
		last_successful TIMESTAMP
	);`)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}

	log.Println("Инициализация базы данных завершена.")
}


func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    // Возвращаем статус 200 OK, если сервер работает
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
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

	pingTimeDuration := time.Duration(result.PingTime) * time.Microsecond

	_, err := db.Exec(
		"INSERT INTO ping_results (ip_address, ping_time, last_successful) VALUES ($1, $2, $3)",
		result.IPAddress, pingTimeDuration, result.LastSuccessful,
	)
	if err != nil {
		http.Error(w, "Ошибка записи в базу данных", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}



func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/ping_results", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPingResults(w, r)
		case http.MethodPost:
			addPingResult(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/health", corsMiddleware(healthCheck))

	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Обработка preflight-запроса OPTIONS
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
