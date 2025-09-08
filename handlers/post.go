package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"handlers/databases"
)

type PostData struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
}
type Category struct {
	Id   int
	Name string
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandler(http.StatusMethodNotAllowed, w)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		errorHandler(http.StatusBadRequest, w)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	topics := r.Form["topics"]
	if len(strings.TrimSpace(title)) == 0 || len(strings.TrimSpace(description)) == 0 || len(topics) == 0 {
		errorHandler(http.StatusBadRequest, w)
		return
	}

	categoriesRows, err := databases.DB.Query("SELECT * FROM categories")
	defer categoriesRows.Close()

	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}

	var allCategories []Category
	for categoriesRows.Next() {
		var name string
		var categoryId int
		if err := categoriesRows.Scan(&categoryId, &name); err != nil {
			log.Fatal(err)
		}
		allCategories = append(allCategories, Category{Id: categoryId, Name: name})
	}

	var updatedTopics []Category
	found := true
	for _, topic := range topics {
		ok, id := contains(allCategories, topic)
		if !ok {
			found = false
			break
		} else {
			updatedTopics = append(updatedTopics, Category{Id: id, Name: topic})
		}
	}

	if !found {
		errorHandler(http.StatusBadRequest, w)
		return
	}

	_, userID := IsLoggedIn(r)

	query := `
		INSERT INTO posts (title, content, user_id)
		VALUES (?, ?, ?)
		`
	res, err := databases.DB.Exec(query, html.EscapeString(title), html.EscapeString(description), userID)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	mu.Lock()
	postID, err := res.LastInsertId()
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	for _, x := range updatedTopics {
		query2 := `INSERT INTO categories_post (categoryID, postID) VALUES (?, ?)`
		_, err := databases.DB.Exec(query2, x.Id, postID)
		if err != nil {
			fmt.Println("11")
			errorHandler(http.StatusInternalServerError, w)
			return
		}
	}
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Data received successfully",
		"title":    title,
		"content":  description,
		"interest": strings.Join(topics, ","),
		"post_id":  postID,
	})
}

func FetchPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorHandler(http.StatusMethodNotAllowed, w)
		return
	}
	mu.Lock()
	_, UserID := IsLoggedIn(r)
	mu.Unlock()

	query := `SELECT id, user_id, content, title, created_at FROM posts`
	rows, err := databases.DB.Query(query)
	if err != nil {
		
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var content, title, interest string
		var createdAt string

		if err := rows.Scan(&id, &userID, &content, &title, &createdAt); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}

		var nickname string
		err = databases.DB.QueryRow(`SELECT nickname FROM users WHERE id = ?`, userID).Scan(&nickname)
		if err != nil {
			log.Println("Nickname not found for user_id:", userID)
			nickname = "Unknown"
		}

		post := map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"content":    content,
			"title":      title,
			"interest":   interest,
			"created_at": createdAt,
			"nickname":   nickname,
			"myId":       UserID,
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
