package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Todo struct {
	Id           int64     `json:"id"`
	Title        string    `json:"title"`
	IsDone       bool      `json:"isDone,omitempty"`
	CreationDate time.Time `json:"creationDate"`
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func AddTodoToDatabase(todo Todo) sql.Result {
	stmt, err := db.Prepare("INSERT INTO todo(title, done, createTime) values (?, ?, ?)")
	CheckErr(err)
	res, err := stmt.Exec(todo.Title, todo.IsDone, todo.CreationDate.Unix())
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
	res := AddTodoToDatabase(todo)

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"error": 500, "message": "could not insert data"}`, http.StatusServiceUnavailable)
		return
	}
	todo.Id = id

	a, _ := json.Marshal(todo)
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, string(a))
}

func HandleTodoPut(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var todo Todo
	var qtodo Todo
	qtodo.Id = -1

	rm, _ := ioutil.ReadAll(r.Body)
	reader := strings.NewReader(string(rm))

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	decoder.Decode(&todo)

	db.QueryRow(`SELECT id, title, done, createTime FROM todo WHERE id = ?`, todo.Id).Scan(&qtodo.Id, &qtodo.Title, &qtodo.IsDone, &qtodo.CreationDate)

	if qtodo.Id == -1 {
		http.Error(w, `{"error": 400, "message": "todo id not found"}`, http.StatusBadRequest)
		return
	}

	reader = strings.NewReader(string(rm))
	decoder = json.NewDecoder(reader)
	decoder.Decode(&qtodo)

	db.Exec(`UPDATE todo SET done = ?, title = ? WHERE id = ?`, qtodo.IsDone, qtodo.Title, qtodo.Id)
	a, _ := json.Marshal(qtodo)
	fmt.Fprint(w, string(a))
}

func HandleTodoDelete(w http.ResponseWriter, r *http.Request) {

}
