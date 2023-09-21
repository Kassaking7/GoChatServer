package mongoDB

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Username string             `bson:"username"`
	Password string             `bson:"password"`
	Email    string             `bson:"email"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := ErrorResponse{
			Message: "Failed to register user",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	user.Password = string(hashedPassword)

	err = saveUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := ErrorResponse{
			Message: "Failed to register user",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := struct {
		Message string `json:"message"`
	}{
		Message: "Registration successful",
	}
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
    print("get JSON response")
	// 从数据库中获取用户
	dbUser, err := findUserByUsername(user.Username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResponse := ErrorResponse{
			Message: "Invalid username or password",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	print("get user from MongoDB")
	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errorResponse := ErrorResponse{
			Message: "Invalid username or password",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	fmt.Println("Setting cookie:", dbUser.Username)
	// 设置cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "username",
		Value: dbUser.Username,
	})

	// 返回响应
	w.WriteHeader(http.StatusOK)
	response := struct {
		Message string `json:"message"`
	}{
		Message: "Login successful",
	}
	json.NewEncoder(w).Encode(response)
}

func saveUser(user User) error {
	collection := GetMongoClient().Database("GoChatDB").Collection("users")

	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}

	return nil
}

func findUserByUsername(username string) (*User, error) {
	collection := GetMongoClient().Database("GoChatDB").Collection("users")

	filter := bson.M{"username": username}
	var user User
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func CreateUniqueIndex() error {
	collection := GetMongoClient().Database("GoChatDB").Collection("users")

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return err
	}

	return nil
}


func GetChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// 获取聊天记录
	chatHistory, err := getChatHistory()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := ErrorResponse{
			Message: "Failed to retrieve chat history",
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// 返回聊天记录
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatHistory)
}

func getChatHistory() ([]bson.M, error) {
	collection := GetMongoClient().Database("GoChatDB").Collection("chat")

	filter := bson.M{} // 可以根据需要添加过滤条件

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// 构造聊天记录数组
	var chatHistory []bson.M
	err = cursor.All(context.Background(), &chatHistory)
	if err != nil {
		return nil, err
	}

	return chatHistory, nil
}
