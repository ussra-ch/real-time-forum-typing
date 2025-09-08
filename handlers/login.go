package handlers

import (
	"database/sql"
	"encoding/json"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"handlers/databases"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

var UsersStatus = make(map[int]string)

type loginInformation struct {
	Nickname string `json:"Nickname"`
	Password string `json:"password"`
}

type data struct {
	Nickname  string `json:"Nickname"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
	Age       string `json:"age"`
	Firstname string `json:"first_name"`
	Lastname  string `json:"last_name"`
	Password  string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandler(http.StatusMethodNotAllowed, w)
		return
	}

	var loginInformations loginInformation
	err := json.NewDecoder(r.Body).Decode(&loginInformations)
	if err != nil {
		errorHandler(http.StatusBadRequest, w)
	}

	var dbPassword string
	var userID int

	err = databases.DB.QueryRow("SELECT id, password FROM users WHERE( nickname = ? or email = ?)", loginInformations.Nickname, loginInformations.Nickname).Scan(&userID, &dbPassword)
	if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(loginInformations.Password)) != nil {
		errorr := ErrorStruct{
			Type: "error",
			Text: "Invalid Nickname or password",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorr)
		return
	} else if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}

	sessionID := generateSessionID(w, userID)
	if sessionID == "" {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})
}

func IsAuthenticated(w http.ResponseWriter, r *http.Request) {
	isloggedn, userID := IsLoggedIn(r)
	if !isloggedn {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": "User not authenticated",
		})
		return
	}
	notifs := unreadMessages(userID)
	user := map[string]interface{}{
		"ok":            true,
		"id":            userID,
		"notifications": notifs,
	}
	mu.Lock()
	UsersStatus[userID] = "online"
	mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorHandler(http.StatusUnauthorized, w)
		return
	}
	_, userId := IsLoggedIn(r)
	mu.Lock()
	UsersStatus[userId] = "offline"
	broadcastUserStatus(nil, userId, "offline", w)
	mu.Unlock()

	cookie, err := r.Cookie("session")
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	_, err = databases.DB.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	for _, connection := range ConnectedUsers[float64(userId)] {
		for i := range OpenedConversations[connection] {
			OpenedConversations[connection][i] = false
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var userInformation data
	err := json.NewDecoder(r.Body).Decode(&userInformation)
	if err != nil {
		errorHandler(http.StatusBadRequest, w)
	}
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)

	if !re.MatchString(userInformation.Email) {
		errorr := ErrorStruct{
			Type: "error",
			Text: "Invalid email",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorr)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInformation.Password), bcrypt.DefaultCost)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}

	tmpAge, _ := strconv.Atoi(userInformation.Age)
	if len(strings.TrimSpace(userInformation.Nickname)) == 0 || tmpAge < 13 || tmpAge > 120 ||
		userInformation.Gender == "" || len(strings.TrimSpace(userInformation.Firstname)) == 0 ||
		len(strings.TrimSpace(userInformation.Lastname)) == 0 || userInformation.Email == "" || userInformation.Password == "" ||
		(userInformation.Gender != "male" && userInformation.Gender != "female") {
		errorr := ErrorStruct{
			Type: "error",
			Text: "Please make sure to fill out all the fields with valid information",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorr)
		return
	}
	res, err := databases.DB.Exec(`
		INSERT INTO users (nickname, age, gender, first_name, last_name, email, password)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		html.EscapeString(userInformation.Nickname), userInformation.Age, html.EscapeString(userInformation.Gender), html.EscapeString(userInformation.Firstname), html.EscapeString(userInformation.Lastname), html.EscapeString(userInformation.Email), hashedPassword)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}

	userID, err := res.LastInsertId()
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	mu.Lock()
	UsersStatus[int(userID)] = "online"
	mu.Unlock()
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = databases.DB.Exec(`
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, ?)`,
		sessionID, userID, expiresAt)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	mu.Lock()
	_, userId := IsLoggedIn(r)
	broadcastUserStatus(nil, userId, "online", w)
	mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
	})
}

func logoutWhenSessionIsDeleted(conn *websocket.Conn) float64 {
	userId, state := findKeyByConn(conn)
	if !state {
		return 0
	}
	return userId
}

