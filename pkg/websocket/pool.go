package websocket
import (
    "fmt"
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type Pool struct {
    Register   chan *Client
    Unregister chan *Client
    Clients    map[*Client]bool
    Broadcast  chan Message
    ChatCollection *mongo.Collection
}


func NewPool(chatCollection *mongo.Collection) *Pool {
    return &Pool{
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
        Clients:    make(map[*Client]bool),
        Broadcast:  make(chan Message),
        ChatCollection: chatCollection,
    }
}

func (pool *Pool) Start() {
    for {
        select {
        case client := <-pool.Register:
            pool.Clients[client] = true
            fmt.Println("Size of Connection Pool: ", len(pool.Clients))
            for client, _ := range pool.Clients {
                fmt.Println(client)
                client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."})
            }
            break
        case client := <-pool.Unregister:
            delete(pool.Clients, client)
            fmt.Println("Size of Connection Pool: ", len(pool.Clients))
            for client, _ := range pool.Clients {
                client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."})
            }
            break
        case message := <-pool.Broadcast:
            fmt.Println("Sending message to all clients in Pool")
            for client, _ := range pool.Clients {
                if err := client.Conn.WriteJSON(message); err != nil {
                    fmt.Println(err)
                    return
                }
            }
            err := pool.saveChatMessage(message)
            if err != nil {
                fmt.Println("Failed to save chat message:", err)
            }
        }
    }
}
func (pool *Pool) saveChatMessage(message Message) error {
    // 创建要保存的聊天消息对象
    chatMessage := bson.M{
        "sender":  message.Sender,
        "content": message.Body,
        // 添加其他字段，如时间戳等
    }

    // 将聊天消息保存到数据库
    _, err := pool.ChatCollection.InsertOne(context.Background(), chatMessage)
    if err != nil {
        return err
    }

    return nil
}