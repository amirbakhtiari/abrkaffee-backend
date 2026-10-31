package user_service

import (
	"context"
	"time"

	"github.com/amirbakhtiari/abrcoffee/internal/core/domain/user"
	port "github.com/amirbakhtiari/abrcoffee/internal/core/ports/user"
	"github.com/amirbakhtiari/abrcoffee/pkg/paginator"
)

type Service struct {
	repo  port.UserRepository
	clock user.Clock
}

func NewService(repo port.UserRepository, clock user.Clock) port.Service {
	return &Service{repo: repo, clock: clock}
}

func (s *Service) Create(ctx context.Context, name, family, mobile, email string, birthDate *time.Time) (*user.User, error) {
	u, err := user.NewUser(name, family, mobile, email, birthDate, s.clock)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, u)
}

func (s *Service) GetByID(ctx context.Context, id uint64) (*user.User, error) {
	if id == 0 {
		return nil, user.ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, req paginator.Request) ([]*user.User, int64, error) {
	return s.repo.List(ctx, req)
}

func (s *Service) Update(ctx context.Context, id uint64, dto port.UpdateUserDTO) (*user.User, error) {
	if id == 0 {
		return nil, user.ErrInvalidID
	}
	existing, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, user.ErrUserNotFound
	}
	if err := existing.ChangeProfile(dto.Name, dto.Family, dto.Email, dto.Mobile, dto.BirthDate, s.clock); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, existing)
}

func (s *Service) SetActive(ctx context.Context, id uint64, active bool) error {
	if id == 0 {
		return user.ErrInvalidID
	}
	return s.repo.SetActive(ctx, id, active)
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return user.ErrInvalidID
	}
	return s.repo.Delete(ctx, id)
}
