package dto

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=3,max=50"`
	Family   string `json:"family" binding:"required,min=3,max=50"`
	Mobile   string `json:"mobile" binding:"required,numeric,len=11"`
	Username string `json:"username" binding:"required,alphanum,min=3,max=30"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
