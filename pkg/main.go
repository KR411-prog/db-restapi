package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	APIPATH = "/apis/v1/books"
)

type Book struct {
	Id,Name,Isbn string
}

type bookslibrary struct {
	dbHost, dbPass, dbName string
}

func main() {

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost:3306"
	}
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "pass123"
	}
	apiPath := os.Getenv("API_PATH")
	if apiPath == "" {
		apiPath = APIPATH
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "library"
	}

	l := bookslibrary{
		dbHost: dbHost,
		dbPass: dbPass,
		dbName: dbName,
	}
	r := mux.NewRouter()
    r.HandleFunc(apiPath, l.getBooks).Methods(http.MethodGet)
	r.HandleFunc(apiPath, l.postBook).Methods(http.MethodPost)
	http.ListenAndServe(":8086", r)
 }

 func (l bookslibrary)postBook(w http.ResponseWriter, r *http.Request) {
	log.Println("post books was called")
	book := Book{}
	json.NewDecoder(r.Body).Decode(&book)
	db := l.openConnection()
	insertQuery, err := db.Prepare("insert into books values(?,?,?)")
	if err != nil {
		log.Fatalf("preparing the db query %s \n", err.Error())
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("preparing the db query %s\n", err.Error())
	}
	_, err = tx.Stmt(insertQuery).Exec(book.Id,book.Name,book.Isbn)
	if err != nil {
		log.Fatalf("execing the insert %s\n", err.Error())
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("while commit the transaction %s\n", err.Error())
	}

	l.closeConnection(db)
 }

 func (l bookslibrary)getBooks(w http.ResponseWriter, r *http.Request) {
	//open connection

	log.Println("getbooks was called")
	db := l.openConnection()
	// read all books
	rows,err := db.Query("select * from books")
	if err != nil {
		log.Fatalf("quering the books table %s\n", err.Error())
	}

	books := []Book{}

	for rows.Next() {
		var id,name,isbn string
		err := rows.Scan(&id,&name,&isbn)
		if err != nil {
			log.Fatalf("while scanning the row %s\n", err.Error())
		}
		aBook := Book{
			Id : id,
			Name: name,
			Isbn: isbn,
		}
		books = append(books,aBook)
	}
	json.NewEncoder(w).Encode(books)
	l.closeConnection(db)
 }

 func (l bookslibrary)openConnection() *sql.DB{
	//username:password@protocol(address)/dbname?param=value
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s","root",l.dbPass, l.dbHost,l.dbName))
	if err != nil {
		log.Fatalf("opening connection to database %s\n", err.Error())
	}
	return db
 }

 func (l bookslibrary) closeConnection(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatalf("closing connection %s\n", err.Error())
	}
 }