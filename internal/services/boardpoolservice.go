package services

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BoardPoolRepo interface {
	Create(ctx context.Context, userID uint, p *models.BoardPool, explicitPosition bool) error
	ListByBoardID(ctx context.Context, userID uint, boardID uuid.UUID) ([]models.BoardPool, error)
	GetByID(ctx context.Context, userID uint, boardID, id uuid.UUID) (*models.BoardPool, error)
	Update(ctx context.Context, userID uint, boardID, id uuid.UUID, updates map[string]any) (*models.BoardPool, error)
	Delete(ctx context.Context, userID uint, boardID, id uuid.UUID) (bool, error)
}

type CreateBoardPoolInput struct {
	Issuer   string
	Subject  string
	BoardID  uuid.UUID
	Title    string
	Color    *string
	Position *int
}

type UpdateBoardPoolInput struct {
	Title    *string
	Color    *string
	Position *int
}

type BoardPoolService interface {
	Create(ctx context.Context, in CreateBoardPoolInput) (*models.BoardPool, error)
	List(ctx context.Context, issuer, sub string, boardID uuid.UUID) ([]models.BoardPool, error)
	GetByID(ctx context.Context, issuer, sub string, boardID, id uuid.UUID) (*models.BoardPool, error)
	Update(ctx context.Context, issuer, sub string, boardID, id uuid.UUID, in UpdateBoardPoolInput) (*models.BoardPool, error)
	Delete(ctx context.Context, issuer, sub string, boardID, id uuid.UUID) error
}

type boardPoolService struct {
	users UserLookup
	repo  BoardPoolRepo
}

func NewBoardPoolService(repo BoardPoolRepo, users UserLookup) BoardPoolService {
	return &boardPoolService{users: users, repo: repo}
}

func (s *boardPoolService) userID(ctx context.Context, issuer, sub string) (uint, error) {
	u, err := s.users.FindByIssuerSub(ctx, issuer, sub)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, custom_errors.ErrNotFound
		}
		return 0, err
	}
	return u.ID, nil
}

// allow: "000000", "#000000", "fff", "#fff"
var hexColorRe = regexp.MustCompile(`^#?([0-9a-fA-F]{6}|[0-9a-fA-F]{3})$`)

func normalizeHexColor(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("color must not be empty")
	}
	if !hexColorRe.MatchString(s) {
		return "", errors.New("invalid color (expected hex like 000000 or #000000)")
	}
	s = strings.TrimPrefix(s, "#")
	// expand 3-digit to 6-digit
	if len(s) == 3 {
		r := string(s[0])
		g := string(s[1])
		b := string(s[2])
		s = r + r + g + g + b + b
	}
	return strings.ToLower(s), nil
}

// optional: ensure position is non-negative
func validatePosition(p int) error {
	if p < 0 {
		return errors.New("position must be >= 0")
	}
	// optional upper bound, remove if not needed
	if p > 1_000_000 {
		return errors.New("position too large: " + strconv.Itoa(p))
	}
	return nil
}

func (s *boardPoolService) Create(ctx context.Context, in CreateBoardPoolInput) (*models.BoardPool, error) {
	uid, err := s.userID(ctx, in.Issuer, in.Subject)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, errors.New("title must not be empty")
	}

	p := models.BoardPool{
		BoardID:  in.BoardID,
		Title:    title,
		Position: 0,
		// Color default is handled by DB default, but we can set explicitly if provided
	}

	if in.Position != nil {
		if err := validatePosition(*in.Position); err != nil {
			return nil, err
		}
		p.Position = *in.Position
	}

	if in.Color != nil {
		c, err := normalizeHexColor(*in.Color)
		if err != nil {
			return nil, err
		}
		p.Color = c
	}

	if err := s.repo.Create(ctx, uid, &p, in.Position != nil); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// board not found / not owned
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *boardPoolService) List(ctx context.Context, issuer, sub string, boardID uuid.UUID) ([]models.BoardPool, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListByBoardID(ctx, uid, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return items, nil
}

func (s *boardPoolService) GetByID(ctx context.Context, issuer, sub string, boardID, id uuid.UUID) (*models.BoardPool, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}
	p, err := s.repo.GetByID(ctx, uid, boardID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *boardPoolService) Update(ctx context.Context, issuer, sub string, boardID, id uuid.UUID, in UpdateBoardPoolInput) (*models.BoardPool, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}

	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return nil, errors.New("title must not be empty")
		}
		updates["title"] = t
	}

	if in.Color != nil {
		c, err := normalizeHexColor(*in.Color)
		if err != nil {
			return nil, err
		}
		updates["color"] = c
	}

	if in.Position != nil {
		if err := validatePosition(*in.Position); err != nil {
			return nil, err
		}
		updates["position"] = *in.Position
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	p, err := s.repo.Update(ctx, uid, boardID, id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *boardPoolService) Delete(ctx context.Context, issuer, sub string, boardID, id uuid.UUID) error {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return err
	}

	ok, err := s.repo.Delete(ctx, uid, boardID, id)
	if err != nil {
		return err
	}
	if !ok {
		return custom_errors.ErrNotFound
	}
	return nil
}
