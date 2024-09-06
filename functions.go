package main

import (
	"fmt"
	"net/http"
	"net/mail"
)

func validEmail(email string) bool { // check if the email format is valid
	_, err := mail.ParseAddress(email)
	return err == nil
}

func votedCheck(r *http.Request) (bool, error) { // true liked, false disliked, error = neither
	likeAdded := r.FormValue("Like")
	dislikeAdded := r.FormValue("Dislike")
	if likeAdded != "" {
		return true, nil
	} else if dislikeAdded != "" {
		return false, nil
	}

	return false, fmt.Errorf("Didn't vote")
}

func hasAddedComment(r *http.Request) (string, bool) {
	comment := r.FormValue("comment")
	if comment != "" {
		return comment, true
	}

	return "", false
}
