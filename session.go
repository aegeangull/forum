package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)
//	Check if the request has a valid session cookie and return the USER ID
func checkSession(r *http.Request) (int, error) { // Check if the request has a valid session cookie and return the USER ID
	cookie, err := r.Cookie("session")
	if err != nil {
		if err == http.ErrNoCookie {
			return 0, nil
		}
		return 0, err
	}
	// Verify that the UUID is valid
	uuid, err := uuid.FromString(cookie.Value) // Verify that the uuID is valid
	fmt.Println(cookie.Value)
	if err != nil {
		return 0, err
	}
	// Get the user ID from the database
	var user_id int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE uuid = ? AND expires_at >= ?", uuid.String(), time.Now()).Scan(&user_id)
	if err != nil {
		return 0, err
	}

	return user_id, nil
}

func createSession(userID int, w *http.ResponseWriter) (*http.Cookie, error) {
	// Generate a new UUID for a session
	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(time.Hour)

	// Insert into the database generated UUID and the user ID
	_, err = db.Exec(`
		INSERT INTO sessions (uuid, user_id, expires_at)
		VALUES(?, ?, ?) 
		ON CONFLICT(user_id) 
		DO UPDATE SET uuid=excluded.uuid, expires_at=excluded.expires_at
		`, uuid.String(), userID, expiresAt)

	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{ // struct creates new session for 1 hour with unique UUID
		Name:    "session",
		Value:   uuid.String(),
		Expires: expiresAt,
	}

	// Return a cookie
	return cookie, nil
}
//	Delete session from database
func deleteSession(r *http.Request) error {
	cookie, err := r.Cookie("session")
	if err != nil {
		if err == http.ErrNoCookie { // if cookie is not present, return nil
			return nil
		}
		return err
	}
	// Verify that the UUID is valid
	uuid, err := uuid.FromString(cookie.Value)
	if err != nil {
		return err
	}
	// Delete the session from the database
	_, err = db.Exec("DELETE FROM sessions WHERE uuid = ?", uuid.String())
	if err != nil {
		return err
	}

	return nil
}
