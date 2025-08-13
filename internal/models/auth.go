package models

// AuthSignUpRequest reprezintă payload-ul pentru /auth/signup
type AuthSignUpRequest struct {
    Name            string `json:"name"`
    Email           string `json:"email"`
    Password        string `json:"password"`
    PasswordConfirm string `json:"passwordConfirm"`
    Phone           string `json:"phone"`
}

// AuthLoginRequest reprezintă payload-ul pentru /auth/login
type AuthLoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// AuthResponse reprezintă răspunsul standard pentru autentificare
type AuthResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Phone string `json:"phone,omitempty"`
}
