package repository

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID    int
	Name  string
	Email string
}

type Bookmark struct {
	ID        int
	UserID    int
	Title     string
	URL       string
	CreatedAt time.Time
}

// ErrEmailTaken signals that the email is already taken.
var ErrEmailTaken = errors.New("email already in use")

// CreateUser inserts a new user, returning ErrEmailTaken if the email is already taken.
func CreateUser(db *sql.DB, name, email string) (*User, error) {
	// Check uniqueness
	var exists bool

	// QueryRow executes a query expected to return at most one row.
	row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email)

	// Scan copies the columns from the matched row into the values pointed to by its arguments.
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrEmailTaken
	}

	// Exec executes a query without returning any rows.
	// The '?' are placeholders for the parameters that follow the query string.
	res, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", name, email)
	if err != nil {
		return nil, err
	}

	// LastInsertId returns the integer ID of the last row inserted.
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &User{ID: int(id), Name: name, Email: email}, nil
}

// CreateBookmark inserts a new bookmark for a given user.
func CreateBookmark(db *sql.DB, userID int, title, url string) (*Bookmark, error) {
	res, err := db.Exec(
		"INSERT INTO bookmarks (user_id, title, url) VALUES (?, ?, ?)", userID, title, url,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	bm := &Bookmark{
		ID:     int(id),
		UserID: userID,
		Title:  title,
		URL:    url,
	}

	// fetch created_at
	err = db.QueryRow("SELECT created_at FROM bookmarks WHERE id = ?", id).Scan(&bm.CreatedAt)
	if err != nil {
		// if we can't get the timestamp, it's better to return the error
		// than a partially populated object.
		return nil, err
	}
	return bm, nil
}

// Helper func to add a bookmark.
func AddBookmark(db *sql.DB, reader *bufio.Reader, userID int) {
	fmt.Print("Title: ")
	title, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading title:", err)
		return
	}
	title = strings.TrimSpace(title)

	fmt.Print("URL: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	bm, err := CreateBookmark(db, userID, title, url)
	if err != nil {
		fmt.Println("could not save bookmark: ", err)
		return
	}
	fmt.Printf("Saved: %s\n", bm.Title)
}

// FetchBookmarks retrieves all bookmarks for a user
func FetchBookmarks(db *sql.DB, userID int) ([]*Bookmark, error) {
	rows, err := db.Query("SELECT id, title, url, created_at FROM bookmarks WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Bookmark
	for rows.Next() {
		var bm Bookmark
		bm.UserID = userID
		if err := rows.Scan(&bm.ID, &bm.Title, &bm.URL, &bm.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &bm)
	}
	return list, nil
}

// Helper func to list bookmarks.
func ListBookmarks(db *sql.DB, userID int) {
	bms, err := FetchBookmarks(db, userID)
	if err != nil {
		fmt.Println("could not retrieve bookmarks:", err)
		return
	}

	if len(bms) == 0 {
		fmt.Println("Empty. You haven't added any bookmarks yet.")
		return
	}

	fmt.Println("\n--- Your bookmarks ---")
	for _, bm := range bms {
		fmt.Printf("\nTitle: %s\nURL: %s\nCreated At: %s\n", bm.Title, bm.URL, bm.CreatedAt.Format(time.RFC3339))
	}
}

// GetUserByEmail retrieves a user by their email address.
// It returns sql.ErrNoRows if no user is found.
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	user := &User{Email: email}

	// since emails are unique, we only want one user, QueryRow is appropriate.
	// db.Query for multiple rows.
	err := db.QueryRow("SELECT id, name FROM users WHERE email = ?", email).Scan(&user.ID, &user.Name)
	if err != nil {
		return nil, err // sql.ErrNoRows is returned if no user is found
	}
	return user, nil
}
