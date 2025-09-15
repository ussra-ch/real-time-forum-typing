package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"handlers/databases"
)

func SortPostsHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	category := r.URL.Query().Get("category")
	query := `SELECT 
    p.id,
    p.user_id,
    p.content,
    p.title,
    p.created_at,
    GROUP_CONCAT(c.name, ', ') AS categories
	FROM 
		posts AS p
	INNER JOIN 
		categories_post AS cp ON p.id = cp.postID
	INNER JOIN 
		categories AS c ON cp.categoryID = c.id
	WHERE 
		p.id IN (
			SELECT 
				p2.id
			FROM 
				posts AS p2
			INNER JOIN 
				categories_post AS cp2 ON p2.id = cp2.postID
			INNER JOIN 
				categories AS c2 ON cp2.categoryID = c2.id
			WHERE 
				c2.name = ?
		)
	GROUP BY 
		p.id, p.user_id, p.content, p.title, p.created_at
	ORDER BY 
    p.id;
`
	if category == "All" {
		query = `SELECT 
    p.id,
    p.user_id,
    p.content,
    p.title,
    p.created_at,
    GROUP_CONCAT(c.name, ', ') AS categories
	FROM 
		posts AS p
	LEFT JOIN 
		categories_post AS cp ON p.id = cp.postID
	LEFT JOIN 
		categories AS c ON cp.categoryID = c.id
	GROUP BY 
		p.id, p.user_id, p.content, p.title, p.created_at
	ORDER BY 
		p.created_at DESC;
`
	}

	rows, err := databases.DB.Query(query, category)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var content, title, interest string
		var createdAt string
		if err := rows.Scan(&id, &userID, &content, &title, &createdAt, &interest); err != nil {
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
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
