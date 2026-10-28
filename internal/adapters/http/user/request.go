package user

type CreateUserRequest struct {
	Name      string `json:"name" binding:"required"`
	Family    string `json:"family" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Mobile    string `json:"mobile" binding:"required"`
	BirthDate string `json:"birth_date"`
}

type UpdateUserRequest struct {
	Name      string `json:"name" binding:"required"`
	Family    string `json:"family" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Mobile    string `json:"mobile" binding:"required"`
	BirthDate string `json:"birth_date"`
}

type ListUsersRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
