package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db  *sql.DB
	err error
)

func main() {
	db, err = sql.Open("sqlite3", "forum.db") // connecting to database
	if err != nil {
		panic(err)
	}

	defer db.Close()

	styles := http.FileServer(http.Dir("./assets/"))
	images := http.FileServer(http.Dir("./upload/"))
	http.Handle("/assets/", http.StripPrefix("/assets/", styles))
	http.Handle("/upload/", http.StripPrefix("/upload", images))

	http.HandleFunc("/post/", postHandler)
	http.HandleFunc("/newpost", newPostHandler)
	http.HandleFunc("/signup", signUpHandler)
	http.HandleFunc("/signin", signInHandler)
	http.HandleFunc("/signout", signOutHandler)
	http.HandleFunc("/", homeHandler)
	fmt.Println("Server is running on 8080")
	http.ListenAndServe(":8080", nil)
}
