package main

import (
	"fmt"
	"log"
	"net/http"

	"handlers/databases"
	"handlers/handlers"
)




func main() {
	databases.InitDB("forum.db")
	defer databases.DB.Close()

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/static/", handlers.ProtectStaticDir)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.RateLimitLoginMiddleware(handlers.LoginHandler))
	http.HandleFunc("/api/logout", handlers.LogoutHandler)
	http.HandleFunc("/api/authenticated", handlers.IsAuthenticated)
	http.HandleFunc("/api/post", handlers.RatelimitMiddleware(handlers.PostHandler, "posts", 10))
	http.HandleFunc("/api/fetch_posts", handlers.FetchPostsHandler)
	http.HandleFunc("/comment", handlers.RatelimitMiddleware(handlers.CommentHandler, "comments", 50))
	http.HandleFunc("/api/fetch_comments", handlers.FetchCommentsHandler)
	http.HandleFunc("/user", handlers.FetchUsers)
	http.HandleFunc("/chat", handlers.WebSocketHandler)
	http.HandleFunc("/api/fetchMessages", handlers.FetchMessages)
	fmt.Println("Server started at http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
