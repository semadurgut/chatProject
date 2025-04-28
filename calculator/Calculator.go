package calculator

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

type IPCalculator struct {
	Network   string `json:"network"`
	Broadcast string `json:"broadcast"`
	HostMin   string `json:"host_min"`
	HostMax   string `json:"host_max"`
}

// IP hesaplama fonksiyonu
func Calculate(ipStr, maskStr string) (IPCalculator, error) {
	_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%s", ipStr, maskStr))
	if err != nil {
		return IPCalculator{}, fmt.Errorf("Geçersiz IP veya Mask değeri")
	}

	network := ipNet.IP
	mask := ipNet.Mask

	// Broadcast adresini hesapla
	broadcast := net.IP(make([]byte, len(network)))
	copy(broadcast, network)
	for i := range mask {
		broadcast[i] |= ^mask[i]
	}

	// Küçük subnetler için kontrol
	if mask[len(mask)-1] >= 254 {
		return IPCalculator{
			Network:   network.String(),
			Broadcast: broadcast.String(),
			HostMin:   "Yok",
			HostMax:   "Yok",
		}, nil
	}

	// İlk ve son kullanılabilir IP
	hostMin := make(net.IP, len(network))
	hostMax := make(net.IP, len(broadcast))
	copy(hostMin, network)
	copy(hostMax, broadcast)
	hostMin[len(hostMin)-1]++
	hostMax[len(hostMax)-1]--

	return IPCalculator{
		Network:   network.String(),
		Broadcast: broadcast.String(),
		HostMin:   hostMin.String(),
		HostMax:   hostMax.String(),
	}, nil
}

// Hesaplama için GET handler
func GetHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	mask := r.URL.Query().Get("mask")

	if ip == "" || mask == "" {
		http.Error(w, "IP ve Mask parametreleri zorunludur", http.StatusBadRequest)
		return
	}

	info, err := Calculate(ip, mask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// Hesaplama için POST handler
func PostHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP   string `json:"ip"`
		Mask string `json:"mask"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Geçersiz JSON formatı", http.StatusBadRequest)
		return
	}

	if req.IP == "" || req.Mask == "" {
		http.Error(w, "IP ve Mask zorunludur", http.StatusBadRequest)
		return
	}

	info, err := Calculate(req.IP, req.Mask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
