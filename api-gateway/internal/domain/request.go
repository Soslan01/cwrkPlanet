package domain

// HTTP Request/Response DTOs

type RegisterRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	Username    string  `json:"username" binding:"required,min=3,max=50"`
	Password    string  `json:"password" binding:"required,min=6"`
	DisplayName string  `json:"displayName" binding:"required"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type RegisterResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	User         *UserResponse `json:"user"`
}

type LoginResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	User         *UserResponse `json:"user"`
}

type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expires_in"`
}

type LogoutResponse struct {
	Success bool `json:"success"`
}

type UserResponse struct {
	ID          int64   `json:"id"`
	Email       string  `json:"email"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	CreatedAt   int64   `json:"createdAt"`
	UpdatedAt   int64   `json:"updatedAt"`
}

type MeResponse struct {
	User *UserResponse `json:"user"`
}
