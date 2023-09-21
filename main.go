package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Kassaking7/GoChatServer/pkg/websocket"
	"github.com/Kassaking7/GoChatServer/pkg/mongoDB"
    "github.com/gorilla/handlers"
)

func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

    cookie, err := r.Cookie("username")
    if err != nil {
        fmt.Println("Failed to get username from cookie:", err)
        return
    }
    username := cookie.Value
    fmt.Println("Received username:", username)
	client := &websocket.Client{
		ID:   username,
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func setRouter(pool *websocket.Pool) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})

	router.HandleFunc("/register", mongoDB.RegisterHandler)
	router.HandleFunc("/login", mongoDB.LoginHandler)
    router.HandleFunc("/chat-history", mongoDB.GetChatHistoryHandler)

    corsHandler := handlers.CORS(
        handlers.AllowCredentials(), // 允许跨域请求携带 Cookie
        handlers.AllowedOrigins([]string{"http://localhost:3000"}), // 允许来自指定域名的跨域请求
        handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), // 允许的 HTTP 方法
        handlers.AllowedHeaders([]string{"Content-Type", "Cookie"}), // 允许的请求头
    )(router)

	return corsHandler
}

func main() {
	fmt.Println("Distributed Chat App v0.01")

	err := mongoDB.InitDB()
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		return
	}
	err = mongoDB.CreateUniqueIndex()
	if err != nil {
		fmt.Println("Error creating unique index:", err)
		return
	}

	chatCollection := mongoDB.GetMongoClient().Database("GoChatDB").Collection("chat")
	pool := websocket.NewPool(chatCollection)
	go pool.Start()

	// 设置路由
	router := setRouter(pool)

	// 启动 HTTP 服务器并监听端口
	http.ListenAndServe(":8080", router)
}
