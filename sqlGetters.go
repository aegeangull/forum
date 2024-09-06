package main

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func getUser(db *sql.DB, id int) (User, error) {
	var user User
	row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

func getUserFromEmail(db *sql.DB, email string) (User, error) {
	var user User
	row := db.QueryRow(`SELECT * FROM users WHERE email = ?`, email)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
	// WORKS BUT NO USAGE FOR IT RIGHT NOW, MIGHT NEED LATER SO WONT DELETE
}

func getUserFromUsername(db *sql.DB, username string) (User, error) {
	var user User
	row := db.QueryRow(`SELECT * FROM users WHERE username = ?`, username)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

func getUserFromLogin(db *sql.DB, email string, password []byte) (User, error) {
	// validate if email and password are correct
	var user User

	encryptedPassword, err := bcrypt.GenerateFromPassword(password, 10) // generate hash with cost of 10
	fmt.Println(string(encryptedPassword), err)
	if err != nil {
		return user, err
	}

	row := db.QueryRow(`SELECT * FROM users WHERE email = ?`, email)
	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		return user, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), password); err != nil {
		return user, fmt.Errorf("incorrect password")
	}

	return user, nil
}

func getPost(db *sql.DB, postID int) (Post, error) {
    var post Post

    row := db.QueryRow(`SELECT id, title, body, created_at, filepath FROM posts WHERE id = $1`, postID)

    err := row.Scan(&post.ID, &post.Title, &post.Body, &post.CreatedAt, &post.Filepath)
    if err != nil {
        if err == sql.ErrNoRows {
            return post, fmt.Errorf("post not found")
        }
        return post, err
    }

    return post, nil
}


func getAllPostsTest(db *sql.DB) ([]Post, error) {
	// m6te selles, et saab frondi jaoks juba sqliga
	var posts []Post
	rows, err := db.Query(`
		SELECT p.*, u.username as author
		FROM posts p 
		LEFT JOIN users u ON u.id = p.user_id`)
	if err != nil {
		// handle error
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err = rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Body, &post.Filepath, &post.CommentCount, &post.Likes, &post.Dislikes, &post.CreatedAt, &post.Author)

		if err != nil {
			return posts, err
		}

		post.Comments, err = getCommentsByPost(db, post.ID)

		if err != nil {
			return posts, err
		}

		post.Categories, err = getPostsCategories(db, post.ID)

		if err != nil {
			return posts, err
		}

		post.Votes, err = getAllVotePost(db, post.ID)

		if err != nil {
			return posts, err
		}

		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		// handle error
		return posts, err
	}
	return posts, nil
}

func getPostsCategories(db *sql.DB, postID int) ([]Category, error) {
	var categories []Category
	rows, err := db.Query(`
		SELECT c.* 
		FROM posts p 
		JOIN post_categories pc ON pc.post_id = p.id 
		JOIN categories c ON c.id = pc.category_id 
		WHERE p.id = ?`, postID)
	if err != nil {
		// handle error
		return categories, err
	}
	defer rows.Close()

	for rows.Next() {
		var category Category
		err = rows.Scan(&category.ID, &category.Name)
		if err != nil {
			// handle error
			return categories, err
		}
		categories = append(categories, category)
	}
	if err = rows.Err(); err != nil {
		// handle error
		return categories, err
	}
	return categories, nil
}

func getCommentsByPost(db *sql.DB, postID int) ([]Comment, error) {
	var comments []Comment
	rows, err := db.Query(`
		SELECT c.*, u.username as author
		FROM comments c 
		LEFT JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ?`, postID)

	if err != nil {
		// handle error
		return comments, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		err = rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.Body, &comment.Likes, &comment.Dislikes, &comment.CreatedAt, &comment.Author)
		if err != nil {
			// handle error
			return comments, err
		}
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		// handle error
		return comments, err
	}
	return comments, nil
}

func getVotePost(db *sql.DB, postID int, userID int) (VotePost, error) { // check if user has already voted on the post
	var vote VotePost

	row := db.QueryRow(`SELECT * FROM votes_post WHERE post_id = ? AND user_id = ?`, postID, userID)

	err := row.Scan(&vote.ID, &vote.UserID, &vote.PostID, &vote.Value, &vote.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return vote, fmt.Errorf("user hasn't voted on this post")
		}
		return vote, err
	}

	return vote, nil
}

func getVoteComment(db *sql.DB, commentID int, userID int) (VoteComment, error) { // check if user has already voted on the post
	var vote VoteComment

	row := db.QueryRow(`SELECT * FROM votes_comment WHERE comment_id = ? AND user_id = ?`, commentID, userID)

	err := row.Scan(&vote.ID, &vote.UserID, &vote.CommentID, &vote.Value, &vote.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return vote, fmt.Errorf("user hasn't voted on this comment")
		}
		return vote, err
	}

	return vote, nil
}

func getAllCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func getAllVotePost(db *sql.DB, postID int) ([]VotePost, error) {
	rows, err := db.Query("SELECT id, user_id, post_id, value, created_at FROM votes_post WHERE post_id = $1", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []VotePost
	for rows.Next() {
		var vp VotePost
		err := rows.Scan(&vp.ID, &vp.UserID, &vp.PostID, &vp.Value, &vp.CreatedAt)
		if err != nil {
			return nil, err
		}
		votes = append(votes, vp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return votes, nil
}
