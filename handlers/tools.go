package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"text/template"

	"handlers/databases"

	"github.com/gorilla/websocket"
)

func errorHandler(errorType int, w http.ResponseWriter) {
	errorr := ErrorStruct{
		Type: "error",
		Text: http.StatusText(errorType),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorType)
	json.NewEncoder(w).Encode(errorr)
}

func generateSessionID(w http.ResponseWriter, userID int) string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
	}
	_, err = databases.DB.Exec(`
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, DATETIME('now', '+1 hour'))
	`, hex.EncodeToString(bytes), userID)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return ""
	}
	return hex.EncodeToString(bytes)
}

func IsLoggedIn(r *http.Request) (bool, int) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false, 0
	}

	var userID int
	err = databases.DB.QueryRow(`
		SELECT user_id FROM sessions 
		WHERE id = ? AND expires_at > DATETIME('now')
	`, cookie.Value).Scan(&userID)
	if err != nil {
		return false, 0
	}

	return true, userID
}

func ProtectStaticDir(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.Dir("static"))
	path := r.URL.Path
	if path == "/static/" || path == "/static/uploads/" {
		errorHandler(http.StatusForbidden, w)
		return
	}

	http.StripPrefix("/static/", fs).ServeHTTP(w, r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
	}
	template, err := template.ParseFiles("index.html")
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
	}

	template.Execute(w, nil)
}

func findKeyByConn(conn *websocket.Conn) (float64, bool) {
	for key, conns := range ConnectedUsers {
		for _, c := range conns {
			if c == conn {
				return key, true
			}
		}
	}
	return 0, false
}

func contains(slice []Category, item string) (bool, int) {
	for _, v := range slice {
		if v.Name == item {
			return true, v.Id
		}
	}
	return false, -1
}

