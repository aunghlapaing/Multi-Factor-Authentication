package handlers

import (
	"html/template"
	"net/http"
)

// Helper function to render 2FA verification page
func render2FAPage(w http.ResponseWriter, errorMsg string, isSetup bool) {
	var templateFile string
	if isSetup {
		templateFile = "templates/setup-2fa.html"
	} else {
		templateFile = "templates/verify-2fa.html"
	}
	
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Error": errorMsg,
	}
	
	tmpl.Execute(w, data)
}
