package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"handlers/databases"

	"github.com/gorilla/websocket"
)

type Client struct {
	Id       int
	Username string
	Conn     *websocket.Conn
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Message struct {
	SenderId       float64 `json:"senderId"`
	ReceiverId     float64 `json:"receiverId"`
	MessageContent string  `json:"messageContent"`
	Seen           bool    `json:"seen"`
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	ReceiverName   string  `json:"receivername"`
}

type Notification struct {
	Type string `json:"type"` // "notification"
	// SenderId    int    `json:"senderId"`
	UnreadCount int `json:"unreadCount"`
}

var (
	ConnectedUsers      = make(map[float64][]*websocket.Conn)
	OpenedConversations = make(map[*websocket.Conn]map[float64]bool)
)

// Send
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	mu.Lock()
	ok, userId := IsLoggedIn(r)
	if !ok || userId == 0 {
		// Not logged in, close connection and return
		conn.Close()
		mu.Unlock()
		return
	}
	// Mark user as online and add to ConnectedUsers
	UsersStatus[userId] = "online"
	ConnectedUsers[float64(userId)] = append(ConnectedUsers[float64(userId)], conn)
	broadcastUserStatus(conn, userId, "online", w)
	sendUnreadNotifications(userId, ConnectedUsers[float64(userId)])
	mu.Unlock()

	defer func() {
		mu.Lock()
		// Remove connection from ConnectedUsers
		deleteOneconnection(userId, conn)
		UsersStatus[userId] = "offline"
		broadcastUserStatus(conn, userId, "offline", w)
		conn.Close()
		mu.Unlock()
	}()

	for {
		_, message, err := conn.NextReader()
		if message == nil {
			mu.Lock()
			UsersStatus[userId] = "offline"
			broadcastUserStatus(conn, userId, "offline", w)
			mu.Unlock()
			break
		}
		if err != nil {
			errorHandler(http.StatusInternalServerError, w)
			return
		}

		var messageStruct Message
		var isConversationOpened bool
		var toolMap map[string]interface{}
		decoder := json.NewDecoder(message)
		_ = decoder.Decode(&toolMap)

		if typeValue, ok := toolMap["type"].(string); ok {
			if typeValue == "OpenConversation" || typeValue == "CloseConversation" {
				mu.Lock()
				conversationOpened(conn, toolMap["receiverId"].(float64), toolMap["type"].(string))
				if typeValue == "OpenConversation" {
					err = updateSeenValue(int(toolMap["senderId"].(float64)), int(toolMap["receiverId"].(float64)))
					if err != nil {
						errorHandler(http.StatusInternalServerError, w)
						return
					}
				}
				sendUnreadNotifications(userId, ConnectedUsers[float64(userId)])
				mu.Unlock()
			}
			if typeValue == "message" {
				messageStruct.SenderId = toolMap["senderId"].(float64)
				var username, receiverName string
				err := databases.DB.QueryRow("SELECT nickname FROM users WHERE id = ?", messageStruct.SenderId).Scan(&username)
				if err != nil {
					errorHandler(http.StatusInternalServerError, w)
					return
				}
				messageStruct.ReceiverId = toolMap["receiverId"].(float64)
				messageStruct.Type = toolMap["type"].(string)
				messageStruct.MessageContent = toolMap["messageContent"].(string)
				messageStruct.Name = username
				messageStruct.ReceiverName = receiverName
				err = messageHandler(messageStruct)
				if err != nil {
					errorHandler(http.StatusInternalServerError, w)
					return
				}
				isConversationOpened = IsConversationOpened(conn, toolMap["receiverId"].(float64), float64(userId))
			}
			if typeValue == "typing" {
				typingJson, err := json.Marshal(map[string]interface{}{
					"type":   "typing",
					"sender": toolMap["senderId"].(float64),
				})
				if err != nil {
					errorHandler(http.StatusInternalServerError, w)
					return
				}
				if ConnectedUsers[toolMap["receiverId"].(float64)] != nil {
					for _, con := range ConnectedUsers[toolMap["receiverId"].(float64)] {
						err = con.WriteMessage(websocket.TextMessage, []byte(typingJson))
						if err != nil {
							errorHandler(http.StatusInternalServerError, w)
							return
						}
					}
				}
			}
			if typeValue == "offline" {
				userID := logoutWhenSessionIsDeleted(conn)
				mu.Lock()
				UsersStatus[int(userID)] = "offline"
				for i := range OpenedConversations[conn] {
					OpenedConversations[conn][i] = false
				}
				broadcastUserStatus(conn, userId, "offline", w)
				deleteOneconnection(userId, conn)
				mu.Unlock()
			}
		}

		if len(messageStruct.MessageContent) > 0 {
			if isConversationOpened {
				Message, err := json.Marshal(messageStruct)
				err1 := updateSeenValue(int(messageStruct.ReceiverId), int(messageStruct.SenderId))
				if err1 != nil || err != nil {
					errorHandler(http.StatusInternalServerError, w)
					return
				}
				for _, con := range ConnectedUsers[messageStruct.ReceiverId] {
					err = con.WriteMessage(websocket.TextMessage, []byte(Message))
					if err != nil {
						errorHandler(http.StatusInternalServerError, w)
						return
					}
				}
			}
			sendUnreadNotifications(int(messageStruct.ReceiverId), ConnectedUsers[messageStruct.ReceiverId])
			Message, err := json.Marshal(messageStruct)
			if err != nil {
				errorHandler(http.StatusInternalServerError, w)
				return
			}
			for _, con := range ConnectedUsers[float64(userId)] {
				if con == conn {
					continue
				}
				err = con.WriteMessage(websocket.TextMessage, []byte(Message))
				if err != nil {
					errorHandler(http.StatusInternalServerError, w)
					return
				}
			}
			sendUnreadNotifications(int(messageStruct.SenderId), ConnectedUsers[messageStruct.SenderId])
		}
	}
}

func FetchMessages(w http.ResponseWriter, r *http.Request) {
	// fetch data
	if r.Method != http.MethodGet {
		errorHandler(http.StatusMethodNotAllowed, w)
		return
	}
	_, userId := IsLoggedIn(r)
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	senderID := r.URL.Query().Get("sender")
	receiverName := ""

	offset, err1 := strconv.Atoi(offsetStr)
	limit, err2 := strconv.Atoi(limitStr)

	if err1 != nil || err2 != nil || limit <= 0 {
		errorHandler(http.StatusBadRequest, w)
		return
	}

	query := fmt.Sprintf(`
    SELECT * FROM messages 
    WHERE (sender_id = ? AND receiver_id = ?) OR (receiver_id = ? AND sender_id = ?)
    ORDER BY id DESC
    LIMIT %d OFFSET %d;`, limit, offset)

	rows, err := databases.DB.Query(query, userId, senderID, userId, senderID)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}

	err = databases.DB.QueryRow("SELECT nickname FROM users WHERE id = ?", userId).Scan(&receiverName)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
		return
	}
	var messages []map[string]interface{}
	for rows.Next() {
		var id, userId, sender_id int
		var content string
		var time time.Time
		var seen bool

		if err := rows.Scan(&id, &sender_id, &userId, &content, &time, &seen); err != nil {
			errorHandler(http.StatusInternalServerError, w)
			return
		}
		message := map[string]interface{}{
			"id":           id,
			"sender_id":    sender_id,
			"userId":       userId,
			"content":      content,
			"time":         time,
			"receiverName": receiverName,
		}

		messages = append(messages, message)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func broadcastUserStatus(conn *websocket.Conn, userId int, tyype string, w http.ResponseWriter) {
	User := make(map[string]interface{})
	User["type"] = tyype
	User["userId"] = userId
	toSend, err := json.Marshal(User)
	if err != nil {
		errorHandler(http.StatusInternalServerError, w)
	}
	for _, connections := range ConnectedUsers {
		for _, con := range connections {
			con.WriteMessage(websocket.TextMessage, []byte(toSend))
		}
	}
}

func userOffline(userId int, conn *websocket.Conn) {
	delete(ConnectedUsers, float64(userId))
	UsersStatus[userId] = "offline"
	newUser := make(map[string]interface{})
	newUser["type"] = "offline"
	newUser["userId"] = userId
	toSend, _ := json.Marshal(newUser)
	for _, connections := range ConnectedUsers {
		for _, con := range connections {
			con.WriteMessage(websocket.TextMessage, []byte(toSend))
		}
	}
	conn.Close()
}

func conversationOpened(conn *websocket.Conn, receiverId float64, typeValue string) {
	if OpenedConversations[conn] == nil {
		OpenedConversations[conn] = make(map[float64]bool)
	}
	if typeValue == "CloseConversation" {
		if receiverId == 0 {
			for i := range OpenedConversations[conn] {
				OpenedConversations[conn][i] = false
			}
		} else {
			OpenedConversations[conn][receiverId] = false
		}
	} else {
		OpenedConversations[conn][receiverId] = true
	}
}

func messageHandler(messageStruct Message) error {
	_, err := databases.DB.Exec(`INSERT INTO messages (sender_id,receiver_id,content,seen )
					VALUES (?, ?, ?, ?);`, messageStruct.SenderId, messageStruct.ReceiverId, html.EscapeString(messageStruct.MessageContent), false)
	if err != nil {
		return err
	}
	return nil
}

func updateSeenValue(receiverId, senderId int) error {
	query := `UPDATE messages
	SET seen = 1
	WHERE messages.sender_id = ? AND messages.receiver_id = ?;`
	_, err := databases.DB.Exec(query, senderId, receiverId)
	if err != nil {
		return err
	}
	return nil
}

func unreadMessages(receiverId int) int {
	var unreadCount int
	err := databases.DB.QueryRow(`
					SELECT COUNT(*) FROM messages
					WHERE receiver_id = ? AND seen = false;
				`, receiverId).Scan(&unreadCount)
	if err != nil {
	}

	return unreadCount
}

func sendUnreadNotifications(userId int, conn []*websocket.Conn) {
	notifs := Notification{
		Type:        "unreadMessage",
		UnreadCount: unreadMessages(userId),
	}
	Notifs, err := json.Marshal(notifs)
	if err != nil {
	}
	for _, con := range conn {
		con.WriteMessage(websocket.TextMessage, Notifs)
	}
}

func deleteOneconnection(userId int, conn *websocket.Conn) {
	for id, c := range ConnectedUsers {
		if id == float64(userId) {
			for i, connection := range c {
				if connection == conn {
					ConnectedUsers[id] = append(c[:i], c[i+1:]...)
					if len(ConnectedUsers[id]) == 0 {
						delete(ConnectedUsers, id)
					}
					break
				}
			}
			break
		}
	}
}

func IsConversationOpened(con *websocket.Conn, receiverId, senderId float64) bool {
	for _, conn := range ConnectedUsers[receiverId] {
		if OpenedConversations[conn][senderId] {
			return true
		}
	}
	return false
}
