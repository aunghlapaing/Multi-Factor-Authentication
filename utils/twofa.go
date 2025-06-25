package utils

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// Generate2FASecret generates a new TOTP secret for 2FA
func Generate2FASecret() (string, error) {
	// Use the library's built-in key generation which is more reliable
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Login Form App",
		AccountName: "user", // This will be replaced in the QR code generation
		SecretSize:  20,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", err
	}
	
	// Return the secret in base32 encoding
	return key.Secret(), nil
}

// GenerateQRCodePNG generates a QR code as PNG bytes for 2FA setup
func GenerateQRCodePNG(secret, email string) ([]byte, error) {
	// Create the proper TOTP URL with the provided secret
	otpURL := "otpauth://totp/" + 
		"Login%20Form%20App:" + email + 
		"?secret=" + secret + 
		"&issuer=Login%20Form%20App" + 
		"&algorithm=SHA1" + 
		"&digits=6" + 
		"&period=30"
	
	// Generate QR code image from the URL directly to PNG bytes
	pngBytes, err := qrcode.Encode(otpURL, qrcode.Medium, 200)
	if err != nil {
		return nil, err
	}
	
	return pngBytes, nil
}

// Generate2FAQRCode generates a QR code for 2FA setup as a data URL
func Generate2FAQRCode(secret, email string) (string, error) {
	// Get the PNG bytes
	pngBytes, err := GenerateQRCodePNG(secret, email)
	if err != nil {
		return "", err
	}
	
	// Convert to base64
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes), nil
}

// Validate2FA validates a 2FA code
func Validate2FA(secret, code string) (bool, error) {
	// Remove spaces from code
	code = strings.ReplaceAll(code, " ", "")
	
	// Validate code using proper TOTP validation with more options
	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      1,       // Allow 1 period skew to account for time drift
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	
	if err != nil {
		return false, err
	}
	
	return valid, nil
}
