package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
)

var ctx = context.Background()

// Redis bağlantısı
var rdb = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_URL"),
	Username: os.Getenv("REDIS_USERNAME"),
	Password: os.Getenv("REDIS_PASSWORD"),
	DB:       0,
})

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]string)
var clientsMu sync.Mutex

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	Type     string `json:"type"`
	UserID   string `json:"userID"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// Kullanıcı Kaydı (Sign In)
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	// JSON verisini al
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	// Kullanıcı adıyla daha önce kayıt olup olmadığını kontrol et
	exists, err := rdb.Exists(ctx, "user:"+user.Username).Result()
	if err != nil {
		http.Error(w, "Sunucu hatası", http.StatusInternalServerError)
		return
	}
	if exists > 0 {
		http.Error(w, "Bu kullanıcı adı zaten kullanılıyor", http.StatusConflict)
		return
	}

	// GUID oluşturma
	userID := generateGUID()

	// Kullanıcıyı Redis'e kaydet
	err = rdb.Set(ctx, "user:"+user.Username, user.Password, 0).Err()
	if err != nil {
		http.Error(w, "Kayıt sırasında bir hata oluştu", http.StatusInternalServerError)
		return
	}

	// Kullanıcı adı ile userID eşleşmesini Redis'e kaydet
	err = rdb.Set(ctx, "userID:"+userID, user.Username, 0).Err()
	if err != nil {
		http.Error(w, "GUID kaydedilemedi", http.StatusInternalServerError)
		return
	}

	err = rdb.Set(ctx, "username:"+user.Username, userID, 0).Err()
	if err != nil {
		http.Error(w, "Kullanıcı ID eşleşmesi kaydedilemedi", http.StatusInternalServerError)
		return
	}

	// GUID'yi cookie'ye kaydet
	http.SetCookie(w, &http.Cookie{
		Name:  "userID",
		Value: userID,
		Path:  "/",
	})

	// Başarılı kayıt yanıtı
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{
		"userID": userID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Kullanıcı Girişi
func LogInHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// JSON verisini al
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Geçersiz istek formatı", http.StatusBadRequest)
		return
	}

	// Kullanıcı adı ile şifreyi Redis'ten al
	storedPassword, err := rdb.Get(ctx, "user:"+requestData.Username).Result()
	if err != nil || storedPassword != requestData.Password {
		http.Error(w, "Kullanıcı adı veya şifre hatalı!", http.StatusUnauthorized)
		return
	}

	// Kullanıcının daha önce bir `userID`si var mı kontrol et
	userID, err := rdb.Get(ctx, "username:"+requestData.Username).Result()
	if err == redis.Nil {
		// Kullanıcı için yeni bir userID oluştur
		userID = generateGUID()

		// Redis'e yeni userID ile kullanıcı adı eşleşmesini kaydet
		rdb.Set(ctx, "userID:"+userID, requestData.Username, 0)
		rdb.Set(ctx, "username:"+requestData.Username, userID, 0)
	} else if err != nil {
		http.Error(w, "Sunucu hatası", http.StatusInternalServerError)
		return
	}

	// Kullanıcı için mevcut veya yeni oluşturulan GUID'yi cookie'ye kaydet
	http.SetCookie(w, &http.Cookie{
		Name:  "userID",
		Value: userID,
		Path:  "/",
	})

	// Başarılı giriş yanıtı
	response := map[string]string{"userID": userID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GUID oluşturma fonksiyonu
func generateGUID() string {
	return uuid.New().String()
}

// Redis'ten kullanıcı adı alma
func GetUsernameHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Hata: userID parametresi eksik", http.StatusBadRequest)
		return
	}

	// userID'ye bağlı olarak kullanıcı adını al
	username, err := rdb.Get(ctx, "userID:"+userID).Result()
	if err != nil {
		http.Error(w, "Kullanıcı adı alınırken hata oluştu.", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"username": username}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Chat için WebSocket
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket bağlantısı hatası:", err)
		return
	}

	/* upgrader.Upgrade(w, r, nil):

	upgrader, github.com/gorilla/websocket paketinden gelen bir WebSocket yükselticisi (upgrader) nesnesidir.
	Upgrade metodu, gelen HTTP isteğini bir WebSocket bağlantısına yükseltmeye çalışır.
	w: HTTP yanıtı yazıcısı (http.ResponseWriter).
	r: HTTP isteği (*http.Request).
	nil: WebSocket bağlantısı için ek HTTP başlıkları eklenmek istenmiyorsa nil olarak bırakılır.

	Dönüş Değerleri:
	conn: WebSocket bağlantısını temsil eden *websocket.Conn türünde bir nesnedir.
	err: Eğer bir hata oluşursa bu değişkene atanır.  */

	defer conn.Close()

	var userID string
	var username string

	clientsMu.Lock()
	clients[conn] = ""
	clientsMu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Kullanıcı (%s) bağlantıyı kapattı.\n", userID)
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			return
		}

		var data Message
		if err := json.Unmarshal(msg, &data); err != nil {
			log.Println("JSON çözümleme hatası:", err)
			continue
		}

		if data.Type == "init" {
			userID = data.UserID
			username = data.Username
			clientsMu.Lock()
			clients[conn] = userID
			clientsMu.Unlock()
			log.Printf("Kullanıcı (%s) bağlandı.\n", userID)
			continue
		}

		rdb.LPush(ctx, "chat_messages", fmt.Sprintf("%s: %s", username, data.Message))
		rdb.Publish(ctx, "chat_channel", fmt.Sprintf("%s: %s", username, data.Message))

	}
}

// Redis Mesaj Dinleyicisi
func ListenToRedisMessages() {
	pubsub := rdb.Subscribe(ctx, "chat_channel")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Println("Redis mesaj alma hatası:", err)
			continue
		}

		clientsMu.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				log.Println("WebSocket mesajı gönderme hatası:", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.Unlock()
	}
}

// Eski Mesajları Yükleme
func LoadMessagesHandler(w http.ResponseWriter, r *http.Request) {
	/* username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Geçersiz kullanıcı adı", http.StatusBadRequest)
		return
	} */

	// Redis'ten eski mesajları al
	messages, err := rdb.LRange(ctx, "chat_messages", 0, -1).Result()
	if err != nil {
		http.Error(w, "Mesajları alırken hata oluştu", http.StatusInternalServerError)
		return
	}

	// Mesajları ters çevir
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
