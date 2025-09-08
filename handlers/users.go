package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"handlers/databases"
)

var mu sync.Mutex

type User struct {
	Nickname string         `json:"nickname"`
	UserId   int            `json:"userId"`
	Photo    sql.NullString `json:"photo"`
	Status   string         `json:"status"`
}

func FetchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorHandler(http.StatusMethodNotAllowed, w)
		return
	}
	mu.Lock()
	loggedIn, userID := IsLoggedIn(r)
	mu.Unlock()

	if !loggedIn {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"loggedIn":    false,
			"nickname":    nil,
			"onlineUsers": []string{},
		})
		return
	}

	var myNickname string
	err := databases.DB.QueryRow("SELECT nickname FROM users WHERE id = ?", userID).Scan(&myNickname)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}

	rows, err := databases.DB.Query(`
		SELECT u.nickname, u.id
		FROM users u
		
	`, userID)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	defer rows.Close()

	type User struct {
		Nickname string    `json:"nickname"`
		UserId   int       `json:"userId"`
		Status   string    `json:"status"`
		Time     time.Time `json:"time"`
	}
	var onlineUsers []User
	for rows.Next() {
		var nickname string
		var userId int

		if err := rows.Scan(&nickname, &userId); err != nil {
			errorHandler(http.StatusInternalServerError, w)
			return
		}

		var T time.Time
		q := `SELECT sent_at FROM messages 
          WHERE (sender_id = ? AND receiver_id = ?) OR (receiver_id = ? AND sender_id = ?)
          ORDER BY sent_at DESC
          LIMIT 1`
		row, err := databases.DB.Query(q, userId, userID, userId, userID)
		if err == nil {
			for row.Next() {
				var sentAt time.Time
				if err := row.Scan(&sentAt); err == nil {
					T = sentAt
				}
			}
			row.Close()
		}

		mu.Lock()
		if _, exists := ConnectedUsers[float64(userId)]; exists {
			UsersStatus[userId] = "online"
		} else {
			UsersStatus[userId] = "offline"
		}
		onlineUsers = append(onlineUsers, User{
			Nickname: nickname,
			UserId:   userId,
			Status:   UsersStatus[userId],
			Time:     T,
		})
		mu.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	mu.Lock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"loggedIn":    true,
		"nickname":    myNickname,
		"onlineUsers": onlineUsers,
		"UserId":      userID,
		"status":      UsersStatus[userID],
	})
	mu.Unlock()
}
