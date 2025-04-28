const messageDiv = document.getElementById("messages");
const sendMessageBtn = document.getElementById("sendMessage");
const messageInput = document.getElementById("message");
const logoutBtn = document.getElementById("logout");

// Çerezi almak için getCookie fonksiyonu
function getCookie(name) {
    let decodedCookie = decodeURIComponent(document.cookie);
    let cookieArray = decodedCookie.split(';');
    for (let i = 0; i < cookieArray.length; i++) {
        let cookie = cookieArray[i].trim();
        if (cookie.indexOf(name + "=") === 0) {
            return cookie.substring(name.length + 1, cookie.length);
        }
    }
    return "";
}

// Çerez ayarlamak için setCookie fonksiyonu
function setCookie(name, value, days) {
    let date = new Date();
    date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000)); // gün sayısı ile süreyi ayarlıyoruz
    let expires = "expires=" + date.toUTCString();
    document.cookie = name + "=" + value + ";" + expires + ";path=/"; // Çerezi tanımlıyoruz
}

// Çerezi silmek için deleteCookie fonksiyonu
function deleteCookie(name) {
    document.cookie = name + "=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/"; // Geçersiz bir tarih ile çerezi siliyoruz
}

// Kullanıcı ID kontrolü ve WebSocket başlatma
async function initializeChat() {
    let userID = getCookie("userID");

    // Eğer cookie'de userID yoksa, yeni bir GUID oluştur ve cookie'ye kaydet
    if (!userID) {
        userID = generateGUID();  // Yeni GUID oluştur
        setCookie("userID", userID, 365);  // Cookie'ye kaydet
    }

    // Kullanıcı adını Redis'ten al
    let username = "";
    try {
        const response = await fetch(`/getUsername?userID=${userID}`);
        const data = await response.json();
        username = data.username || "Bilinmeyen Kullanıcı";
    } catch (error) {
        console.error("Kullanıcı adı alınırken hata:", error);
    }

    // WebSocket bağlantısını kur
    const conn = new WebSocket(`${document.location.protocol}//${document.location.host}/chat`);

    conn.onopen = () => {
        console.log("WebSocket bağlantısı başarılı");
        if (username) {
            conn.send(JSON.stringify({ type: "init", userID, username }));
        } else {
            alert("Lütfen giriş yapın.");
            window.location.href = "/dinamic/logIn.html";  // Giriş sayfasına yönlendir
        }
    };

    // Mesaj alma
    conn.onmessage = (event) => {
        const data = event.data;
        const now = new Date();
        const time = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

        const messageData = data.split(': '); // "username: mesaj" formatını ayırıyoruz
        const sender = messageData[0];
        const text = messageData.slice(1).join(': ');

        // Ana kapsayıcı
        const messageContainer = document.createElement("div");
        messageContainer.classList.add("message-wrapper");

        // Kullanıcı adı ve saat bilgisi (balon dışında)
        const infoDiv = document.createElement("div");
        infoDiv.classList.add("message-info");
        infoDiv.textContent = `${sender} • ${time}`;

        // Mesaj baloncuğu
        const messageBubble = document.createElement("div");
        messageBubble.classList.add("message");
        messageBubble.classList.add(sender === username ? "sent" : "received");
        messageBubble.textContent = text;

        // Elemanları birleştir
        messageContainer.appendChild(infoDiv);      // Kullanıcı adı ve saat
        messageContainer.appendChild(messageBubble); // Mesaj baloncuğu

        messageDiv.appendChild(messageContainer);
        messageDiv.scrollTop = messageDiv.scrollHeight;
    };

    // Mesaj gönderme fonksiyonu
    function sendMessage() {
        const message = messageInput.value.trim();
        if (message && username) {
            conn.send(JSON.stringify({ type: "message", userID, username, message }));
            messageInput.value = "";
        } else {
            alert("Lütfen bir mesaj giriniz.");
        }
    }

    // Butona tıklayarak mesaj gönderme
    sendMessageBtn.onclick = sendMessage;

    // Enter tuşu ile mesaj gönderme
    messageInput.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
            e.preventDefault();
            sendMessage();
        }
    });

    console.log("Kullanıcı adı: ", username)

    // Eski mesajları yükle
    window.onload = () => {
        console.log("Kullanıcı adı: ", username)
        if (!username) {
            // Uyarı göster ve login sayfasına yönlendir
            alert("Lütfen giriş yapın.");
            window.location.href = "/dinamic/logIn.html"; // Giriş sayfasına yönlendirme
        }

        fetch("/loadMessages")
            .then(response => response.json())
            .then(messages => {
                messages.forEach(msg => {
                    const [userInfo, ...messageParts] = msg.split(': ');
                    const messageText = messageParts.join(': ').trim();
                    const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

                    // Kullanıcı kontrolü: Mesaj kullanıcıya ait mi?
                    const isCurrentUser = userInfo === username;

                    // Mesaj konteyneri
                    const messageContainer = document.createElement("div");
                    messageContainer.classList.add("message-wrapper");

                    // Kullanıcı adı ve saat bilgisi
                    const infoDiv = document.createElement("div");
                    infoDiv.classList.add("message-info");
                    infoDiv.classList.add(isCurrentUser ? "sent" : "received"); // Sağ/Sol ayarı
                    infoDiv.textContent = `${userInfo} • ${time}`;

                    // Mesaj baloncuğu
                    const messageBubble = document.createElement("div");
                    messageBubble.classList.add("message");
                    messageBubble.classList.add(isCurrentUser ? "sent" : "received"); // Sağ/Sol ayarı
                    messageBubble.textContent = messageText;

                    // Elemanları birleştir
                    messageContainer.appendChild(infoDiv);
                    messageContainer.appendChild(messageBubble);

                    // Mesajı DOM'a ekle
                    document.getElementById("messages").appendChild(messageContainer);
                });

                // Scroll'u en alta çek
                messageDiv.scrollTop = messageDiv.scrollHeight;
            })
            .catch(error => console.error("Eski mesajlar alınırken hata:", error));
    };

    // Oturumu Sonlandırma
    logoutBtn.onclick = () => {
        if (confirm("Oturumu sonlandırmak istediğinize emin misiniz?")) {
            deleteCookie("userID");  // Cookie'yi sil
            window.location.href = "/dinamic/logIn.html";  // Giriş sayfasına yönlendir
        }
    };
}

initializeChat();