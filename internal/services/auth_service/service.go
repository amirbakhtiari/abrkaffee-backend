package auth_service

import (
	"context"
	"errors"
	"time"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
	"github.com/amirbakhtiari/abrcoffee/pkg/utils"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
)
var _ ports.AuthService = (*AuthService)(nil)

type AuthService struct {
	repo      ports.AuthRepository
	jwtConfig *config.Config
}

func NewAuthService(repo ports.AuthRepository, jwtConfig *config.Config) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtConfig: jwtConfig,
	}
}
func (s *AuthService) Register(ctx context.Context, name, family, username, password, mobile string) (*user.User, error) {
	existingUser, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	existingUser, err = s.repo.GetByMobile(ctx, mobile)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}
	hashedPassword, err := utils.NewPasswordHash(password)
	if err != nil {
		return nil, err
	}
	userEntity, err := user.NewUser(name, family, username, hashedPassword, mobile, "", nil)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateUser(ctx, userEntity); err != nil {
		return nil, err
	}
	return userEntity, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	var user *user.User
	var err error

	user, err = s.repo.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return "", errors.New("invalid credentials")
	}
	if !user.IsActive {
		return "", errors.New("your account has been deactivated. please contact support")
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid credentials")
	}
	token, err := utils.GenerateToken(user.ID, user.Username, "Admin", s.jwtConfig.JWT.Secret)
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	return token, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uint64, name, family, mobile string, birthDate *time.Time) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.Name = name
	user.Family = family
	user.Birthdate = birthDate
	user.Mobile = mobile
	return s.repo.UpdateUser(ctx, user)
}

func (s *AuthService) UpdatePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}

	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		return ErrInvalidPassword
	}

	newHashedPassword, err := utils.NewPasswordHash(newPassword)
	if err != nil {
		return err
	}
	user.Password = newHashedPassword
	return s.repo.UpdateUser(ctx, user)
}

func (s *AuthService) ChangeStatus(ctx context.Context, userID uint64, targetStatus bool) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}
	user.IsActive = targetStatus
	user.UpdatedAt = time.Now()
	return s.repo.UpdateUser(ctx, user)
}

func (s *AuthService) GetUsersList(ctx context.Context, page, pageSize int) ([]*user.User, paginator.Pagination, error) {
	p := paginator.Translate(page, pageSize)
	user, totalRows, err := s.repo.ListUsers(ctx, p.Page, p.PageSize)
	if err != nil {
		return nil, p, err
	}
	p.CreateResponse(totalRows)
	return user, p, nil
}
