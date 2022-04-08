package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Todo struct {
	Title        string `json:"title"`
	IsDone       bool
	CreationDate time.Time
}

func AddTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "Request method is not acceptable")
		log.Printf("AddTodo Handler: can't handle %v method instead of POST", r.Method)
		return
	}

}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func AddTodoToDatabase(todo Todo) sql.Result {
	stmt, err := db.Prepare("INSERT INTO todo(title, done, createTime) values (?, ?, ?)")
	CheckErr(err)
	var isDone int
	if todo.IsDone {
		isDone = 1
	} else {
		isDone = 0
	}
	res, err := stmt.Exec(todo.Title, isDone, todo.CreationDate.Unix())
	CheckErr(err)
	return res
}

func CreateTodoTable() {
	if db != nil {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS todo (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title VARCHAR(255) NOT NULL,
			done BOOLEAN DEFAULT 0 NOT NULL,
			createTime DATETIME NOT NULL 
		)`)
		CheckErr(err)
	}
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./todos.db")
	CheckErr(err)
	CreateTodoTable()

	mux := http.NewServeMux()
	mux.HandleFunc("/todo", HandleTodos)
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}

func HandleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandleTodoGet(w, r)
	case http.MethodPost:
		HandleTodoPost(w, r)
	case http.MethodPut:
		HandleTodoPut(w, r)
	case http.MethodDelete:
		HandleTodoDelete(w, r)
	default:
		http.Error(w, `{ "error": -1, "message": "Method not supported"}`, http.StatusMethodNotAllowed)
	}
}

func HandleTodoGet(w http.ResponseWriter, r *http.Request) {

}

func HandleTodoPost(w http.ResponseWriter, r *http.Request) {
	var todo Todo

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&todo)

	if err != nil {
		http.Error(w, `{"error": -1, "message": "error decoding json"}`, http.StatusBadRequest)
		return
	}

	if todo.Title == "" {
		http.Error(w, `{"error": 400, "message": "title should not be null"}`, http.StatusBadRequest)
		return
	}

	todo.IsDone = false
	todo.CreationDate = time.Now()
	AddTodoToDatabase(todo)
}

func HandleTodoPut(w http.ResponseWriter, r *http.Request) {

}

func HandleTodoDelete(w http.ResponseWriter, r *http.Request) {

}
