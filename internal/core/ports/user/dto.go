package user

import "time"

type UpdateUserDTO struct {
	Name      string
	Family    string
	Email     string
	Mobile    string
	BirthDate *time.Time
}
