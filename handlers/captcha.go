package handlers

import (
	"net/http"
	"sync"
)

// CaptchaStore stores captcha images for serving
type CaptchaStore struct {
	Captchas map[string][]byte
	mutex    sync.RWMutex
}

// Global captcha store
var captchaStore = &CaptchaStore{
	Captchas: make(map[string][]byte),
}

// StoreCaptchaImage stores a captcha image for serving
func StoreCaptchaImage(id string, imageBytes []byte) {
	captchaStore.mutex.Lock()
	defer captchaStore.mutex.Unlock()
	captchaStore.Captchas[id] = imageBytes
}

// CaptchaImageHandler serves captcha images
func CaptchaImageHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Captcha ID is required", http.StatusBadRequest)
		return
	}

	captchaStore.mutex.RLock()
	imageBytes, exists := captchaStore.Captchas[id]
	captchaStore.mutex.RUnlock()

	if !exists {
		http.Error(w, "Captcha not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(imageBytes)
}
