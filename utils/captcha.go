package utils

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())
}

// Captcha represents a captcha challenge
type Captcha struct {
	ID        string
	Solution  string
	ImageBytes []byte
	Created   time.Time
}

// CaptchaStore is a simple in-memory store for captchas
type CaptchaStore struct {
	captchas map[string]Captcha
	mutex    sync.RWMutex
}

// Global captcha store
var captchaStore = &CaptchaStore{
	captchas: make(map[string]Captcha),
}

// GenerateCaptcha creates a new captcha challenge
func GenerateCaptcha() (Captcha, error) {
	// Generate a random ID
	idBytes := make([]byte, 16)
	_, err := cryptorand.Read(idBytes)
	if err != nil {
		return Captcha{}, err
	}
	id := base64.URLEncoding.EncodeToString(idBytes)

	// Generate a random solution (simple math problem)
	a, err := cryptorand.Int(cryptorand.Reader, big.NewInt(10))
	if err != nil {
		return Captcha{}, err
	}
	b, err := cryptorand.Int(cryptorand.Reader, big.NewInt(10))
	if err != nil {
		return Captcha{}, err
	}
	
	// Create the math problem and solution
	problem := fmt.Sprintf("%d + %d = ?", a, b)
	solution := fmt.Sprintf("%d", a.Int64()+b.Int64())

	// Generate the captcha image
	img := createCaptchaImage(problem)
	
	// Encode the image to PNG bytes
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return Captcha{}, err
	}

	// Create the captcha
	captcha := Captcha{
		ID:        id,
		Solution:  solution,
		ImageBytes: buf.Bytes(),
		Created:   time.Now(),
	}

	// Store the captcha
	captchaStore.mutex.Lock()
	captchaStore.captchas[id] = captcha
	captchaStore.mutex.Unlock()

	// Clean up old captchas
	go cleanupOldCaptchas()

	return captcha, nil
}

// ValidateCaptcha checks if a captcha solution is correct
func ValidateCaptcha(id, solution string) bool {
	captchaStore.mutex.RLock()
	captcha, exists := captchaStore.captchas[id]
	captchaStore.mutex.RUnlock()

	if !exists {
		return false
	}

	// Check if the captcha has expired (5 minutes)
	if time.Since(captcha.Created) > 5*time.Minute {
		captchaStore.mutex.Lock()
		delete(captchaStore.captchas, id)
		captchaStore.mutex.Unlock()
		return false
	}

	// Compare the solution (case insensitive)
	isValid := strings.TrimSpace(strings.ToLower(solution)) == strings.TrimSpace(strings.ToLower(captcha.Solution))

	// Remove the captcha after validation (one-time use)
	if isValid {
		captchaStore.mutex.Lock()
		delete(captchaStore.captchas, id)
		captchaStore.mutex.Unlock()
	}

	return isValid
}

// Helper function to create a simple captcha image with a math problem
func createCaptchaImage(text string) image.Image {
	width, height := 200, 80
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Fill background with light color
	bgColor := color.RGBA{240, 240, 240, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
	
	// Add some noise
	addNoise(img)
	
	// Draw the text
	drawText(img, text)
	
	return img
}

// Helper function to add noise to the image
func addNoise(img *image.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if rand.Intn(20) == 0 {
				r, g, b := uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256))
				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}
	
	// Add some random lines
	for i := 0; i < 5; i++ {
		x1 := rand.Intn(bounds.Max.X)
		y1 := rand.Intn(bounds.Max.Y)
		x2 := rand.Intn(bounds.Max.X)
		y2 := rand.Intn(bounds.Max.Y)
		r, g, b := uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256))
		drawLine(img, x1, y1, x2, y2, color.RGBA{r, g, b, 255})
	}
}

// Helper function to draw text on the image
func drawText(img *image.RGBA, text string) {
	bounds := img.Bounds()
	textColor := color.RGBA{0, 0, 0, 255}
	
	// Calculate position to center the text
	x := bounds.Min.X + 50
	y := bounds.Min.Y + bounds.Dy()/2
	
	// Draw each character with slight rotation
	for _, char := range text {
		// Skip spaces
		if char == ' ' {
			x += 15
			continue
		}
		
		// Draw the character
		drawChar(img, string(char), x, y, textColor)
		
		// Move to the next position
		x += 20
	}
}

// Helper function to draw a character on the image
func drawChar(img *image.RGBA, char string, x, y int, col color.Color) {
	// Simple font rendering (very basic)
	switch char {
	case "0":
		drawCircle(img, x, y, 10, col)
	case "1":
		drawLine(img, x, y-10, x, y+10, col)
	case "2":
		drawLine(img, x-5, y-10, x+5, y-10, col)
		drawLine(img, x+5, y-10, x+5, y, col)
		drawLine(img, x+5, y, x-5, y, col)
		drawLine(img, x-5, y, x-5, y+10, col)
		drawLine(img, x-5, y+10, x+5, y+10, col)
	case "3":
		drawLine(img, x-5, y-10, x+5, y-10, col)
		drawLine(img, x+5, y-10, x+5, y+10, col)
		drawLine(img, x-5, y+10, x+5, y+10, col)
		drawLine(img, x-5, y, x+5, y, col)
	case "4":
		drawLine(img, x-5, y-10, x-5, y, col)
		drawLine(img, x-5, y, x+5, y, col)
		drawLine(img, x+5, y-10, x+5, y+10, col)
	case "5":
		drawLine(img, x-5, y-10, x+5, y-10, col)
		drawLine(img, x-5, y-10, x-5, y, col)
		drawLine(img, x-5, y, x+5, y, col)
		drawLine(img, x+5, y, x+5, y+10, col)
		drawLine(img, x-5, y+10, x+5, y+10, col)
	case "6":
		drawLine(img, x-5, y-10, x+5, y-10, col)
		drawLine(img, x-5, y-10, x-5, y+10, col)
		drawLine(img, x-5, y, x+5, y, col)
		drawLine(img, x+5, y, x+5, y+10, col)
		drawLine(img, x-5, y+10, x+5, y+10, col)
	case "7":
		drawLine(img, x-5, y-10, x+5, y-10, col)
		drawLine(img, x+5, y-10, x+5, y+10, col)
	case "8":
		drawCircle(img, x, y-5, 5, col)
		drawCircle(img, x, y+5, 5, col)
	case "9":
		drawLine(img, x-5, y-10, x+5, y-10, col)
		drawLine(img, x-5, y-10, x-5, y, col)
		drawLine(img, x-5, y, x+5, y, col)
		drawLine(img, x+5, y-10, x+5, y+10, col)
	case "+":
		drawLine(img, x-5, y, x+5, y, col)
		drawLine(img, x, y-5, x, y+5, col)
	case "-":
		drawLine(img, x-5, y, x+5, y, col)
	case "=":
		drawLine(img, x-5, y-3, x+5, y-3, col)
		drawLine(img, x-5, y+3, x+5, y+3, col)
	case "?":
		drawLine(img, x-5, y-10, x+5, y-10, col)
		drawLine(img, x+5, y-10, x+5, y, col)
		drawLine(img, x, y, x+5, y, col)
		drawLine(img, x, y, x, y+5, col)
		drawPoint(img, x, y+8, col)
	default:
		// For any other character, just draw a dot
		drawPoint(img, x, y, col)
	}
}

// Helper function to draw a line
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))
	sx, sy := 1, 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy
	
	for {
		img.Set(x1, y1, col)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// Helper function to draw a point
func drawPoint(img *image.RGBA, x, y int, col color.Color) {
	img.Set(x, y, col)
}

// Helper function to draw a circle
func drawCircle(img *image.RGBA, x, y, r int, col color.Color) {
	for i := 0; i <= 360; i++ {
		rad := float64(i) * math.Pi / 180
		px := int(float64(r) * math.Cos(rad)) + x
		py := int(float64(r) * math.Sin(rad)) + y
		img.Set(px, py, col)
	}
}

// Helper function to encode an image to base64
func encodeImageToBase64(img image.Image) (string, error) {
	var buf strings.Builder
	
	// Write the image to a buffer
	w := base64.NewEncoder(base64.StdEncoding, &buf)
	err := png.Encode(w, img)
	if err != nil {
		return "", err
	}
	
	// Close the encoder to flush any partial blocks
	if err := w.Close(); err != nil {
		return "", err
	}
	
	return "data:image/png;base64," + buf.String(), nil
}

// Helper function to clean up old captchas
func cleanupOldCaptchas() {
	captchaStore.mutex.Lock()
	defer captchaStore.mutex.Unlock()
	
	now := time.Now()
	for id, captcha := range captchaStore.captchas {
		if now.Sub(captcha.Created) > 5*time.Minute {
			delete(captchaStore.captchas, id)
		}
	}
}
