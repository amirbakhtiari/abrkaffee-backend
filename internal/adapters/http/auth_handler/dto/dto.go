package dto

type RegisterRequest struct {
	UserID   uint64 `json:"user_id" binding:"required,gt=0"`
	Username string `json:"username" binding:"required,alphanum,min=3,max=30"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
