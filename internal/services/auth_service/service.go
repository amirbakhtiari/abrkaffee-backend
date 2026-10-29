package auth_service

import (
	"context"
	"errors"
	"time"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/auth"
	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	"github.com/amirbakhtiari/abrcoffee/internal/core/ports"
	userports "github.com/amirbakhtiari/abrcoffee/internal/core/ports/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
	"github.com/amirbakhtiari/abrcoffee/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
)
var _ ports.AuthService = (*AuthService)(nil)

type AuthService struct {
	repo      ports.AuthRepository
	userRepo  userports.UserRepository
	jwtConfig *config.Config
}

func NewAuthService(
	repo ports.AuthRepository,
	userRepo userports.UserRepository,
	jwtConfig *config.Config,
) *AuthService {
	return &AuthService{
		repo:      repo,
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	userID uint64,
	username string,
	password string,
) (*auth.Account, error) {
	profile, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if profile == nil {
		return nil, ErrUserNotFound
	}

	existing, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	existing, err = s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := utils.NewPasswordHash(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	account := auth.NewAccount(
		userID,
		profile,
		username,
		hashedPassword,
		"Staff",
		now,
	)

	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
) (string, error) {
	account, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if account == nil {
		return "", errors.New("invalid credentials")
	}

	if !account.IsActive || account.Profile == nil || !account.Profile.IsActive() {
		return "", errors.New("your account has been deactivated. please contact support")
	}

	if !utils.CheckPasswordHash(password, account.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	token, err := utils.GenerateToken(
		account.UserID,
		account.Username,
		account.Role,
		s.jwtConfig.JWT.Secret,
	)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

func (s *AuthService) UpdatePassword(
	ctx context.Context,
	userID uint64,
	oldPassword, newPassword string,
) error {
	account, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		return ErrUserNotFound
	}

	if !utils.CheckPasswordHash(oldPassword, account.PasswordHash) {
		return ErrInvalidPassword
	}

	newHash, err := utils.NewPasswordHash(newPassword)
	if err != nil {
		return err
	}

	account.PasswordHash = newHash
	account.UpdatedAt = time.Now()

	return s.repo.UpdateAccount(ctx, account)
}

func (s *AuthService) ChangeStatus(
	ctx context.Context,
	userID uint64,
	targetStatus bool,
) error {
	account, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		return ErrUserNotFound
	}

	account.IsActive = targetStatus
	account.UpdatedAt = time.Now()

	return s.repo.UpdateAccount(ctx, account)
}

func (s *AuthService) GetUsersList(
	ctx context.Context,
	page, pageSize int,
) ([]*user.User, paginator.Meta, error) {
	req := paginator.NewRequest(page, pageSize)

	users, total, err := s.userRepo.List(ctx, req)
	if err != nil {
		return nil, paginator.Meta{}, err
	}

	meta := paginator.NewMeta(req.Page, req.PageSize, total)
	return users, meta, nil
}

func (s *AuthService) UpdateProfile(
	ctx context.Context,
	userID uint64,
	name, family, mobile string,
	birthDate *time.Time,
) error {
	profile, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return ErrUserNotFound
	}

	if err := profile.ChangeProfile(
		name,
		family,
		profile.Email(),
		mobile,
		birthDate,
		systemClock{},
	); err != nil {
		return err
	}

	_, err = s.userRepo.Update(ctx, profile)
	return err
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}
