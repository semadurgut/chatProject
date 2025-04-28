package main

import (
	"chatProject/calculator"
	"chatProject/chat"
	"chatProject/location_info"
	"log"
	"net/http"
)

func main() {
	// Kullanıcı ilk geldiğinde login sayfasına yönlendirilecek
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// /dinamic/logIn.html dosyasına yönlendirme yapılacak
		http.Redirect(w, r, "/dinamic/logIn.html", http.StatusFound)
	})

	// Dinamik içeriklerin sunulması
	http.HandleFunc("/calculate", calculator.GetHandler)
	http.HandleFunc("/calculatePost", calculator.PostHandler)

	// LocationInfo işlemleri için handlerlar
	http.HandleFunc("/location", location_info.LocationHandler)

	// WebSocket mesajlarını dinlemeye başlatıyoruz
	go chat.ListenToRedisMessages()

	// Chat işlemleri için handler'lar
	http.HandleFunc("/signIn", chat.SignInHandler)
	http.HandleFunc("/logIn", chat.LogInHandler)
	http.HandleFunc("/getUsername", chat.GetUsernameHandler)
	http.HandleFunc("/chat", chat.ChatHandler)
	http.HandleFunc("/loadMessages", chat.LoadMessagesHandler)

	// Dinamik dosyaları da servis etmek için
	http.Handle("/dinamic/", http.StripPrefix("/dinamic", http.FileServer(http.Dir("./dinamic"))))

	// Sunucuyu başlatıyoruz
	log.Println("Sunucu başlatılıyor... Port: 8085")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
