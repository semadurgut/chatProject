package location_info

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"io"
	"net/http"
)

// Redis bağlantısı için gerekli değişkenler
var ctx = context.Background()
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

// Konum bilgilerini tutacak yapı
type Location struct {
	City     string `json:"city"`
	Region   string `json:"regionName"`
	Country  string `json:"country"`
	Timezone string `json:"timezone"`
}

/*
// Şu an kullanılmıyor ama ileride gerekirse aktif edilebilir
// Kullanıcının IP adresini alma fonksiyonu
func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0]) // Eğer birden fazla IP varsa ilkini al
	}
	return r.RemoteAddr
}
*/

// IP'ye göre konum bilgisi çeken fonksiyon ( kullanıcının dış (public) IP adresini alıp, ardından bu IP adresini kullanarak ipinfo.io API'sinden konum bilgilerini çeker)
func getLocation() (Location, error) {
	var loc Location

	// İlk olarak dış IP'yi al
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return loc, err
	}
	defer resp.Body.Close()

	// IP'yi al
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return loc, err
	}

	var ipResponse map[string]string
	err = json.Unmarshal(body, &ipResponse)
	if err != nil {
		return loc, err
	}

	ip := ipResponse["ip"]
	fmt.Printf(ip) // Gelen ip değerini kontrol etme

	// IP'yi kullanarak IPinfo.io API'sine sorgu yap
	url := fmt.Sprintf("https://ipinfo.io/%s/json?token=2c67db470c3c82", ip)
	resp, err = http.Get(url)
	if err != nil {
		return loc, err
	}
	defer resp.Body.Close()

	// API yanıtını al
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return loc, err
	}

	// fmt.Println("IP API Yanıtı: ", string(body)) // API'den gelen yanıtı logla
	err = json.Unmarshal(body, &loc)
	if err != nil {
		return loc, err
	}

	return loc, nil
}

// Kullanıcının GUID'ine göre Redis'ten kullanıcı adını çeken fonksiyon
func getUserName(guid string) (string, error) {
	username, err := redisClient.Get(ctx, "userID:"+guid).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("kullanıcı bulunamadı")
	} else if err != nil {
		return "", err
	}
	return username, nil
}

// Kullanıcının konumunu ve kullanıcı adını JSON olarak döndüren handler
func LocationHandler(w http.ResponseWriter, r *http.Request) {
	// Kullanıcı ID'sini cookie'den al
	guidCookie, err := r.Cookie("userID")
	if err != nil {
		http.Error(w, "Kullanıcı ID bulunamadı", http.StatusUnauthorized)
		return
	}
	guid := guidCookie.Value

	// Redis'ten kullanıcı adını al
	username, err := getUserName(guid)
	if err != nil {
		http.Error(w, "Kullanıcı adı alınamadı", http.StatusInternalServerError)
		return
	}

	// Kullanıcının konumunu al
	location, err := getLocation()
	if err != nil {
		http.Error(w, "Konum bilgisi alınamadı", http.StatusInternalServerError)
		return
	}

	// Yanıt olarak JSON döndür
	response := map[string]interface{}{
		"username": username,
		"city":     location.City,
		"country":  location.Country,
		"timezone": location.Timezone,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("JSON kodlama hatası: %v", err), http.StatusInternalServerError)
	}
}

/*
Dış IP adresini api.ipify.org API’sinden alır.
Bu IP adresini ipinfo.io API’sine göndererek konum bilgilerini alır.
Gelen JSON yanıtını Location struct’ına çevirir.
Başarıyla çalışırsa Location nesnesini döndürür, hata olursa error döndürür.
*/
