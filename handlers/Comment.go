package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"

	"handlers/databases"
)

type CommentData struct {
	PostID  string `json:"post_id"`
	Content string `json:"comment"`
}

type Comment struct {
	ID        int
	Content   string
	CreatedAt string
	UserID    string
	PostID    string
	Name      string
}

func CommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost{
		errorHandler(http.StatusMethodNotAllowed, w)
		return
	}
	var cd CommentData
	if err := json.NewDecoder(r.Body).Decode(&cd); err != nil {
		errorHandler(http.StatusBadRequest, w)
		return
	}
	if len(strings.TrimSpace(cd.Content)) == 0{
		errorHandler(http.StatusBadRequest, w)
		return
	}
	var exists bool
	err := databases.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)", cd.PostID).Scan(&exists)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}

	if !exists {
		errorHandler(http.StatusBadRequest, w)
		return
	}

	// Insert comment
	_, userID := IsLoggedIn(r)
	_, err = databases.DB.Exec(`
		INSERT INTO comments (post_id, user_id, content)
		VALUES (?, ?, ?)
	`, cd.PostID, userID, html.EscapeString(cd.Content))
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Comment created successfully",
	})
}

func FetchCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet{
		errorHandler(http.StatusBadRequest, w)
		return
	}
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	offset, err1 := strconv.Atoi(offsetStr)
	limit, err2 := strconv.Atoi(limitStr)

	if err1 != nil || err2 != nil || limit <= 0 {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	rows, err := databases.DB.Query(fmt.Sprintf(`
    SELECT 
        comments.id,
        comments.content,
        comments.created_at,
        comments.user_id,
        comments.post_id,
        users.nickname
    FROM comments
    JOIN users ON comments.user_id = users.id
    ORDER BY comments.created_at DESC
    LIMIT %d OFFSET %d;`, limit, offset))
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var c Comment
		err := rows.Scan(&c.ID, &c.Content, &c.CreatedAt, &c.UserID, &c.PostID, &c.Name)
		if err != nil {
			log.Println("Error scanning comment:", err)
			continue
		}
		comments = append(comments, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
