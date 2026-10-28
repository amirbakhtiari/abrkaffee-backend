package user

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrInvalidNameRequired   = errors.New("name is required")
	ErrInvalidFamilyRequired = errors.New("family is required")
	ErrInvalidMobileRequired = errors.New("mobile is invalid")
)

type Clock interface {
	Now() time.Time
}

type User struct {
	id        uint64
	name      string
	family    string
	email     string
	mobile    string
	birthdate *time.Time
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
}

func NewUser(
	name, family, mobile, email string,
	birthDate *time.Time,
	clock Clock,
) (*User, error) {

	name = strings.TrimSpace(name)
	family = strings.TrimSpace(family)
	email = strings.TrimSpace(email)
	mobile = strings.TrimSpace(mobile)

	if name == "" {
		return nil, ErrInvalidNameRequired
	}

	if family == "" {
		return nil, ErrInvalidFamilyRequired
	}

	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	if !isValidMobile(mobile) {
		return nil, ErrInvalidMobileRequired
	}

	now := clock.Now()

	return &User{
		name:      name,
		family:    family,
		email:     email,
		mobile:    mobile,
		birthdate: birthDate,
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}, nil
}
func (u *User) ID() uint64 {
	return u.id
}
func (u *User) Name() string {
	return u.name
}
func (u *User) Family() string {
	return u.family
}
func (u *User) Email() string {
	return u.email
}
func (u *User) Mobile() string {
	return u.mobile
}
func (u *User) IsActive() bool {
	return u.isActive
}
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return re.MatchString(strings.ToLower(email))
}

func isValidMobile(m string) bool {
	// simple production-safe baseline
	re := regexp.MustCompile(`^[0-9]{10,15}$`)
	return re.MatchString(m)
}

func (u *User) SetID(id uint64) {
	u.id = id
}

func (u *User) Deactivate(clock Clock) {
	u.isActive = false
	u.updatedAt = clock.Now()
}

func (u *User) Activate(clock Clock) {
	u.isActive = true
	u.updatedAt = clock.Now()
}

func (u *User) BirthDate() *time.Time {
	return u.birthdate
}

func Rehydrate(
	id uint64,
	name, family, email, mobile string,
	birthDate *time.Time,
	isActive bool,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		id:        id,
		name:      name,
		family:    family,
		email:     email,
		mobile:    mobile,
		birthdate: birthDate,
		isActive:  isActive,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (u *User) ChangeProfile(
	name string,
	family string,
	email string,
	mobile string,
	birthDate *time.Time,
	clock Clock,
) error {

	name = strings.TrimSpace(name)
	family = strings.TrimSpace(family)
	email = strings.TrimSpace(email)
	mobile = strings.TrimSpace(mobile)

	if name == "" {
		return ErrInvalidNameRequired
	}

	if family == "" {
		return ErrInvalidFamilyRequired
	}

	if !isValidEmail(email) {
		return ErrInvalidEmail
	}

	if !isValidMobile(mobile) {
		return ErrInvalidMobileRequired
	}

	u.name = name
	u.family = family
	u.email = email
	u.mobile = mobile
	u.birthdate = birthDate
	u.updatedAt = clock.Now()

	return nil
}
