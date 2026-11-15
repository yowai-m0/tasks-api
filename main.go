package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type Server struct {
	db *sql.DB
}

type Task struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Done       bool      `json:"done"`
	Created_at time.Time `json:"created_at"`
}

type TaskDone struct {
	Done bool `json:"done"`
}

func getID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, err
	}
	return id, nil
}

// func PrintAllTasks(db *sql.DB, r *http.Request) error {
//     ctx, cancel := context.WithTimeout(r.Context(), 3 * time.Second)
//     defer cancel()
// 	rows, err := db.QueryContext(ctx, "SELECT id, title, done, created_at FROM tasks")
// 	if err != nil {
// 		fmt.Print("Errors: ")
// 		return err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var (
// 			id         int
// 			title      string
// 			done       bool
// 			created_at time.Time
// 		)

// 		err := rows.Scan(&id, &title, &done, &created_at)
// 		if err != nil {
// 			fmt.Print("Errors read data")
// 			return err
// 		}
// 		fmt.Printf("%d: %s - Status: %v (Time created: %v)\n", id, title, done, created_at)
// 	}

// 	if err := rows.Err(); err != nil {
// 		fmt.Print("ошибка при обходе: ")
// 		return err
// 	}
// 	return nil
// }

func (s *Server) getHadnler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	dataColums := make([]Task, 0)

	rows, err := s.db.QueryContext(ctx, "SELECT id, title, done, created_at FROM tasks")
	if err != nil {
		http.Error(w, "Errors get query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id         int
			title      string
			done       bool
			created_at time.Time
		)

		err := rows.Scan(&id, &title, &done, &created_at)
		if err != nil {
			http.Error(w, "Errors read data", http.StatusInternalServerError)
			return
		}

		dataColums = append(dataColums, Task{
			ID:         id,
			Title:      title,
			Done:       done,
			Created_at: created_at,
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Errors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(dataColums); err != nil {
		http.Error(w, "Errors encode data", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getHandlerPath(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	id, err := getID(r)
	if err != nil {
		http.Error(w, "Errors, id must be num", http.StatusBadRequest)
		return
	}

	var (
		tasks      Task
		title      string
		done       bool
		created_at time.Time
	)
	errRow := s.db.QueryRowContext(ctx, "SELECT title, done, created_at FROM tasks WHERE id = $1",
		id).Scan(
		&title,
		&done,
		&created_at,
	)

	tasks = Task{
		ID:         id,
		Title:      title,
		Done:       done,
		Created_at: created_at,
	}

	if errors.Is(errRow, sql.ErrNoRows) {
		http.Error(w, "Errors invalid id", http.StatusNotFound)
		return
	}
	if errRow != nil {
		http.Error(w, "Errors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "Errors", http.StatusInternalServerError)
		return
	}
}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	var tasks Task
	if err := json.NewDecoder(r.Body).Decode(&tasks); err != nil {
		http.Error(w, "Errors decode data", http.StatusBadRequest)
		return
	}
	var (
		id         int
		title      string
		done       bool
		created_at time.Time
	)

	if tasks.Title == "" || len([]rune(tasks.Title)) > 100 {
		http.Error(w, "Errors invalid title", http.StatusBadRequest)
		return
	}

	err := s.db.QueryRowContext(ctx, "INSERT INTO tasks (title) VALUES ($1) RETURNING id, title, done, created_at",
		tasks.Title).Scan(&id, &title, &done, &created_at)
	if err != nil {
		http.Error(w, "errors invalid data", http.StatusInternalServerError)
		return
	}
	tasks = Task{
		ID:         id,
		Title:      title,
		Done:       done,
		Created_at: created_at,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "errors encode data", http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	id, err := getID(r)
	if err != nil {
		http.Error(w, "Errors invalid id", http.StatusBadRequest)
		return
	}

	res, err := s.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Errors delete", http.StatusInternalServerError)
		return
	}

	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "errors", http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		http.Error(w, "Errors not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) patchHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	id, err := getID(r)
	if err != nil {
		http.Error(w, "Errors, Invalid id", http.StatusBadRequest)
		return
	}

	var (
		title      string
		done       bool
		created_at time.Time
	)

	task := TaskDone{
		Done: done,
	}
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Errors decode data", http.StatusBadRequest)
		return
	}

	err = s.db.QueryRowContext(ctx, "UPDATE tasks SET done = $1 WHERE id = $2 RETURNING id, title, done, created_at",
		task.Done, id).Scan(&id, &title, &done, &created_at)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Errors not found", http.StatusNotFound)
			return
		} else {
			log.Println(err)
			http.Error(w, "Errors server", http.StatusInternalServerError)
			return
		}
	}

	tasks := Task{
		ID:         id,
		Title:      title,
		Done:       done,
		Created_at: created_at,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Println("Errors json encode", err)
		return
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла:", err)
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")
	user := os.Getenv("DB_USER")

	connectStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		host, port, dbName, user, password)

	db, err := sql.Open("pgx", connectStr)
	if err != nil {
		fmt.Println("Errors open db", err)
		return
	}

	srv := &Server{db: db}

	defer srv.db.Close()

	err = srv.db.Ping()
	if err != nil {
		fmt.Println("Errors connect to db", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", srv.getHadnler)
	mux.HandleFunc("GET /tasks/{id}", srv.getHandlerPath)
	mux.HandleFunc("POST /tasks", srv.postHandler)
	mux.HandleFunc("DELETE /tasks/{id}", srv.deleteHandler)
	mux.HandleFunc("PATCH /tasks/{id}", srv.patchHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
