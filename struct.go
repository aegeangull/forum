package main

import (
	"time"
)

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
}

type Category struct {
	ID   int
	Name string
}

type Post struct {
	ID           int
	UserID       int
	Author       string // add it later for reading
	Title        string
	Body         string
	Filepath	 string
	CommentCount int
	Comments     []Comment  // add it later for reading
	Categories   []Category // add it later	for reading
	Votes        []VotePost // add it later for reading
	Likes        int
	Dislikes     int
	CreatedAt    time.Time
}

type Comment struct {
	ID        	int
	UserID    	int
	PostID    	int
	Body      	string
	Likes     	int
	Dislikes  	int
	CreatedAt 	time.Time
	Author    	string
}

type VotePost struct {
	ID        int
	UserID    int
	PostID    int
	Value     bool
	CreatedAt time.Time
}

type VoteComment struct {
	ID        int
	UserID    int
	CommentID int
	Value     bool
	CreatedAt time.Time
}
