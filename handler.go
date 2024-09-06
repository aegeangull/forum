package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	//"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "golang.org/x/crypto/bcrypt"
)

// handles requests to the main page

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	signupError := ""
	id, _ := checkSession(r)

	if id != 0 {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	username, email, password := r.FormValue("username"), r.FormValue("email"), r.FormValue("password")
	fmt.Println(username, email, password, "form REGISTER")

	// email and username to lower case
	username, email = strings.ToLower(username), strings.ToLower(email)

	_, errEmail := getUserFromEmail(db, email)          // should give no rows error
	_, errUsername := getUserFromUsername(db, username) // should give no rows error

	if errEmail == nil {
		signupError = "An user with this e-mail already exists. Try another e-mail."
	} else if errUsername == nil {
		signupError = "An user with this username already exists. Try another username."
	} else if r.Method == http.MethodPost {
		// unique email and username, do additional checks
		if len(password) < 6 {
			signupError = "Your password is too short. Minimum length is 6 characters."
		} else if !validEmail(email) {
			signupError = "Your e-mail format is invalid."
		} else if len(username) <= 2 {
			signupError = "The minimum username length is 2 characters"
		} else if len(username) > 13 {
			signupError = "The maximum username length is 13 characters"
		} else {
			// attempt to register
			encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10) // turn password into a hash with the cost of 10
			if err != nil {
				fmt.Println("ERROR while generating password hash: ", err)
				return
			}

			result, err := db.Exec(`INSERT INTO users(username, email, password) VALUES (?, ?, ?)`, username, email, encryptedPassword)
			if err != nil {
				fmt.Println("failed to save into database ERROR: ", err)
				return
			}

			fmt.Println(result, " Successfully registered")
			// up to this point everything works i think?
			http.Redirect(w, r, "/signin", http.StatusFound)
		}
	}

	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		panic(err)
	}

	tmpl.Execute(w, signupError)
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := checkSession(r)
	if id != 0 {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	email, password := r.FormValue("email"), r.FormValue("password")
	loginError := ""

	user, err := getUserFromLogin(db, email, []byte(password))

	if len(email) > 0 && len(password) > 0 && err != nil {
		loginError = "Invalid e-mail or password"
	}

	if err == nil {
		cookie, err := createSession(user.ID, &w)
		if err == nil {
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}

	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		panic(err)
	}

	tmpl.Execute(w, loginError)
}

func signOutHandler(w http.ResponseWriter, r *http.Request) {
	// c := &http.Cookie{
	// 	Name:     "storage",
	// 	Value:    "",
	// 	Path:     "/",
	// 	MaxAge:   -1,
	// 	HttpOnly: true,
	// }

	// http.SetCookie(w, c)
	deleteSession(r)
	http.Redirect(w, r, "/", http.StatusFound)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the query parameters
	q := r.URL.Query()
	category := q.Get("category")
	myPosts := q.Get("my-posts")
	likedPosts := q.Get("liked-posts")

	userId, err := checkSession(r)

	user, err := getUser(db, userId)

	posts, err := getAllPostsTest(db)

	categories, err := getAllCategories(db)

	if category != "" {
		filteredPosts := make([]Post, 0)
		for _, post := range posts {
			for _, cat := range post.Categories {
				if cat.Name == category {
					filteredPosts = append(filteredPosts, post)
					break
				}
			}
		}
		posts = filteredPosts
	} else if myPosts != "" {
		filteredPosts := make([]Post, 0)
		for _, post := range posts {
			if post.UserID == user.ID {
				filteredPosts = append(filteredPosts, post)
			}
		}
		posts = filteredPosts
	} else if likedPosts != "" {
		filteredPosts := make([]Post, 0)
		for _, post := range posts {
			for _, vote := range post.Votes {
				if vote.UserID == user.ID {
					filteredPosts = append(filteredPosts, post)
					break
				}
			}
		}
		posts = filteredPosts
	}

	m := map[string]interface{}{
		"UserID":           user.ID,
		"Posts":            posts,
		"UserName":         user.Username,
		"Categories":       categories,
		"SelectedCategory": category,
	}

	// main page
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		panic(err)
	}

	tmpl.Execute(w, m)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var loggedInUser User

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	cookieUserID, err := checkSession(r)
	if err == nil {
		loggedInUser, err = getUser(db, cookieUserID)
	}

	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	voteValue, err := votedCheck(r) // false = disliked, true = liked
	if err == nil {                 // if error is nil then user voted
		if loggedInUser.ID == 0 {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		commentIDString := r.FormValue("CommentID")
		commentID, _ := strconv.Atoi(commentIDString)

		if commentID != 0 {
			castVoteComment(db, loggedInUser.ID, commentID, voteValue)
		} else {
			castVotePost(db, loggedInUser.ID, id, voteValue)
		}
	}

	post, err := getPost(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get post", http.StatusInternalServerError)
		return
	}

	user, err := getUser(db, post.UserID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	post.Author = user.Username
	post.Categories, err = getPostsCategories(db, id)
	if err != nil {
		http.Error(w, "Failed to get post categories", http.StatusInternalServerError)
		return
	}
	post.Comments, err = getCommentsByPost(db, id)
	if err != nil {
		http.Error(w, "Failed to get post comments", http.StatusInternalServerError)
		return
	}

	commentValue, isCommentAdded := hasAddedComment(r)

	if isCommentAdded {
		if loggedInUser.ID == 0 {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		addComment(db, loggedInUser.ID, post.ID, commentValue)
	}
	filePath := post.Filepath
	fmt.Println(filePath)
	m := map[string]interface{}{
		"UserID":   loggedInUser.ID,
		"UserName": loggedInUser.Username,
		"Post":     post,
		"FilePath": filePath, // Add the file path to the m map
	}

	err = tmpl.Execute(w, m)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

type CategoriesTemp struct { // for testing purposes
	Category   string
	Image      string // in-case we want to add an emoji to each category
	CategoryID int
}

// const uploadPath = "./upload"
func newPostHandler(w http.ResponseWriter, r *http.Request) {
	// Check user session exists
	id, err := checkSession(r)
	if err != nil || id == 0 {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	user, err := getUser(db, id)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	categoriesToAdd := []CategoriesTemp{
		{Category: "General", Image: "🌎", CategoryID: 1},
		{Category: "Sports", Image: "⚽️", CategoryID: 2},
		{Category: "Games", Image: "🎮", CategoryID: 3},
		{Category: "Technology", Image: "🖥️", CategoryID: 4},
		{Category: "Books", Image: "📚", CategoryID: 5},
	}

	if r.Method == http.MethodPost {
		// Parse form data
		err = r.ParseMultipartForm(20 << 20) // Max 20 MB file size
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Limit file size to 20 MB
		maxSize := int64(20 << 20) // 20 MB
		file, fileHeader, err := r.FormFile("file")
		if err != nil || file == nil {
			http.Error(w, "Invalid or missing file upload", http.StatusBadRequest)
			return
		}
		defer file.Close()

		limitedFile := io.LimitReader(file, maxSize)
		if limitedFile == nil {
			fmt.Println("Error: limitedFile is nil")
		}

		if fileHeader.Size > maxSize {
			http.Error(w, "File size exceeds 20 MB", http.StatusBadRequest)
			return
		}
		// Get post data from form
		postTitle := strings.TrimSpace(r.PostFormValue("postTitle"))
		postBody := strings.TrimSpace(r.PostFormValue("postBody"))
		categoryNames := r.PostForm["CategoryName"]

		// Check required fields are not empty
		if postTitle == "" || postBody == "" || len(categoryNames) == 0 {
			http.Error(w, "All fields are required to add a post", http.StatusBadRequest)
			return
		}

		ext := strings.ToLower(path.Ext(fileHeader.Filename))
		validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".svg": true, ".webp": true}
		if !validExts[ext] {
			http.Error(w, "Unsupported file type", http.StatusUnsupportedMediaType)
			return
		}

		// Save uploaded file
		originalFileName := filepath.Base(fileHeader.Filename)
		localFileName := "upload/" + getNameByDate() + filepath.Ext(originalFileName)

		// Log the file path
		log.Println("File saved to:", localFileName)

		out, err := os.Create(localFileName)
		if err != nil {
			http.Error(w, "Failed to save uploaded file: "+err.Error(), http.StatusInternalServerError)

			//http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}

		// Map category names to IDs and add post to database
		categoryIDs := []int{}

		for _, categoryName := range categoryNames {
			for _, tempCat := range categoriesToAdd {
				if strings.EqualFold(tempCat.Category, categoryName) {
					categoryIDs = append(categoryIDs, tempCat.CategoryID)
					break
				}
			}
		}

		postID, err := addPost(db, id, postTitle, postBody, localFileName, categoryIDs)
		if err != nil {
			http.Error(w, "Failed to add post", http.StatusInternalServerError)
			return
		}

		// Redirect to the new post page
		http.Redirect(w, r, "/post/?id="+strconv.FormatInt(postID, 10), http.StatusFound)
	} else {
		// Display new post form page

		// insert the file path into the template
		// Filepath := "/upload/" + getNameByDate()
		m := map[string]interface{}{
			"UserID":   id,
			"UserName": user.Username,
			"Error":    "",
			"Cat":      categoriesToAdd,
			//"Image":    Filepath,
		}

		tmpl, err := template.ParseFiles("templates/newPost.html")
		if err != nil {
			http.Error(w, "Failed to load page template", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, m)
		if err != nil {
			http.Error(w, "Failed to load page template", http.StatusInternalServerError)
			return
		}
	}
}

// get the current time in nanoseconds and convert it to string
func getNameByDate() string {
	nanosec := time.Now().Nanosecond()
	filename := strconv.Itoa(nanosec)
	return filename
}
