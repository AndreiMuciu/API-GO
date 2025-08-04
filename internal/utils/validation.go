package utils

import (
	"regexp"
	"strings"
)

// IsValidEmail verifică dacă emailul are un format valid
func IsValidEmail(email string) bool {
    if email == "" {
        return false
    }
    
    // Regex simplă pentru validarea email-ului
    emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    re := regexp.MustCompile(emailRegex)
    return re.MatchString(email)
}

// IsValidPhone verifică dacă numărul de telefon are un format valid
func IsValidPhone(phone string) bool {
    if phone == "" {
        return true // telefonul e opțional
    }
    
    // Elimină spațiile și caracterele speciale
    cleanPhone := strings.ReplaceAll(phone, " ", "")
    cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
    cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
    cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")
    
    // Verifică format pentru România: +40xxxxxxxxx sau 07xxxxxxxx
    phoneRegex := `^(\+40|0040|0)[67]\d{8}$`
    re := regexp.MustCompile(phoneRegex)
    return re.MatchString(cleanPhone)
}

// ValidateUserInput validează toate câmpurile unui UserInput
func ValidateUserInput(input interface{}) []string {
    var errors []string
    
    // Type assertion pentru a verifica tipul
    switch user := input.(type) {
    case map[string]interface{}:
        // Validare nume
        if name, exists := user["name"]; exists {
            if nameStr, ok := name.(string); ok && nameStr != "" {
                if len(nameStr) < 2 {
                    errors = append(errors, "name must be at least 2 characters")
                }
                if len(nameStr) > 50 {
                    errors = append(errors, "name cannot exceed 50 characters")
                }
            }
        }
        
        // Validare email
        if email, exists := user["email"]; exists {
            if emailStr, ok := email.(string); ok && emailStr != "" {
                if !IsValidEmail(emailStr) {
                    errors = append(errors, "invalid email format")
                }
            }
        }
        
        // Validare telefon
        if phone, exists := user["phone"]; exists {
            if phoneStr, ok := phone.(string); ok && phoneStr != "" {
                if !IsValidPhone(phoneStr) {
                    errors = append(errors, "invalid phone format (use +40xxxxxxxxx or 07xxxxxxxx)")
                }
            }
        }
        
        // Validare parolă
        if password, exists := user["password"]; exists {
            if passStr, ok := password.(string); ok && passStr != "" {
                if len(passStr) < 8 {
                    errors = append(errors, "password must be at least 8 characters")
                }
                if !hasUppercase(passStr) {
                    errors = append(errors, "password must contain at least one uppercase letter")
                }
                if !hasDigit(passStr) {
                    errors = append(errors, "password must contain at least one digit")
                }
            }
        }
    }
    
    return errors
}

// hasUppercase verifică dacă string-ul conține cel puțin o literă mare
func hasUppercase(s string) bool {
    for _, r := range s {
        if r >= 'A' && r <= 'Z' {
            return true
        }
    }
    return false
}

// hasDigit verifică dacă string-ul conține cel puțin o cifră
func hasDigit(s string) bool {
    for _, r := range s {
        if r >= '0' && r <= '9' {
            return true
        }
    }
    return false
}