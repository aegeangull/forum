package main

import (
	"database/sql"
	"fmt"
	"strings"
)

func addPost(db *sql.DB, UserID int, Title, Body string, Filepath string, categoryIDs []int) (int64, error) {
	// Begin a new transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert the post into the posts table
	stmt, err := tx.Prepare("INSERT INTO posts (user_id, title, body, filepath) VALUES (?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	// Execute the INSERT statement with the post values
	res, err := stmt.Exec(UserID, Title, Body, Filepath)
	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted post
	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert the category IDs into the post_categories table
	stmt, err = tx.Prepare("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	// Execute the INSERT statement for each category ID
	for _, categoryID := range categoryIDs {
		_, err = stmt.Exec(postID, categoryID)
		if err != nil {
			return 0, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	// Return the ID of the newly inserted post
	fmt.Println("Post added to database")
	return postID, nil
}

func addComment(db *sql.DB, UserID, PostID int, Body string) error {
	// Prepare the INSERT statement
	stmt, err := db.Prepare("INSERT INTO comments (user_id, post_id, body) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the INSERT statement with the comment values
	_, err = stmt.Exec(UserID, PostID, Body)
	if err != nil {
		return err
	}

	// Increase comment count
	_, err = db.Exec("UPDATE posts SET comment_count = comment_count + 1 WHERE id = ?", PostID)

	fmt.Println("Comment added to database")
	return nil
}
// this function enabls users to vote on posts, chnage their vote and remove their vote
func castVotePost(db *sql.DB, userID, postID int, value bool) error {
	query := "INSERT INTO votes_post (user_id, post_id, value) VALUES (?, ?, ?)"
	_, err := db.Exec(query, userID, postID, value)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			originalVote, err := getVotePost(db, postID, userID)
			if originalVote.Value == value { // if user double votes same vote then remove vote
				_, err = db.Exec("DELETE FROM votes_post WHERE user_id = ? AND post_id = ?", userID, postID)
				if value {
					_, err = db.Exec("UPDATE posts SET likes = likes - 1 WHERE id = ?", postID) // remeove like from votes_post
				} else {
					_, err = db.Exec("UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?", postID) // remove dislike from votes_post
				}
				return err
			} else {
				// if user changed their vote then update previous vote
				_, err = db.Exec("UPDATE votes_post SET value = ? WHERE user_id = ? AND post_id = ?", value, userID, postID)

				if value {
					_, err = db.Exec("UPDATE posts SET likes = likes + 1 WHERE id = ?", postID)
					_, err = db.Exec("UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?", postID)
				} else {
					_, err = db.Exec("UPDATE posts SET likes = likes - 1 WHERE id = ?", postID)
					_, err = db.Exec("UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?", postID)
				}

			}
			// Handle other errors here
			return err
		}

	}

	// Update the post's likes or dislikes based on the vote value
	if value {
		_, err = db.Exec("UPDATE posts SET likes = likes + 1 WHERE id = ?", postID)
	} else {
		_, err = db.Exec("UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?", postID)
	}
	if err != nil {
		return err
	}
	return nil
}
// check if user has voted on post and return vote
func castVoteComment(db *sql.DB, userID, commentID int, value bool) error {
	query := "INSERT INTO votes_comment (user_id, comment_id, value) VALUES (?, ?, ?)"
	_, err := db.Exec(query, userID, commentID, value)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			originalVote, err := getVoteComment(db, commentID, userID)
			if originalVote.Value == value { // if user double votes same vote then remove vote
				_, err = db.Exec("DELETE FROM votes_comment WHERE user_id = ? AND comment_id = ?", userID, commentID)
				if value {
					_, err = db.Exec("UPDATE comments SET likes = likes - 1 WHERE id = ?", commentID) // remeove like from votes_post
				} else {
					_, err = db.Exec("UPDATE comments SET dislikes = dislikes - 1 WHERE id = ?", commentID) // remove dislike from votes_post
				}
				return err
			} else {
				// if user changed their vote then update previous vote
				_, err = db.Exec("UPDATE votes_comment SET value = ? WHERE user_id = ? AND comment_id = ?", value, userID, commentID)

				if value {
					_, err = db.Exec("UPDATE comments SET likes = likes + 1 WHERE id = ?", commentID)
					_, err = db.Exec("UPDATE comments SET dislikes = dislikes - 1 WHERE id = ?", commentID)
				} else {
					_, err = db.Exec("UPDATE comments SET likes = likes - 1 WHERE id = ?", commentID)
					_, err = db.Exec("UPDATE comments SET dislikes = dislikes + 1 WHERE id = ?", commentID)
				}

			}
			// Handle other errors here
			return err
		}

	}

	// Update the post's likes or dislikes based on the vote value
	if value {
		_, err = db.Exec("UPDATE comments SET likes = likes + 1 WHERE id = ?", commentID)
	} else {
		_, err = db.Exec("UPDATE comments SET dislikes = dislikes + 1 WHERE id = ?", commentID)
	}
	if err != nil {
		return err
	}
	return nil
}
