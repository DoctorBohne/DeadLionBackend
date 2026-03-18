package boards

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BoardPoolRepo struct {
	db *gorm.DB
}

func NewBoardPoolRepo(db *gorm.DB) *BoardPoolRepo {
	return &BoardPoolRepo{db: db}
}

// scope pools by user via JOIN boards (ownership check)
func (r *BoardPoolRepo) scoped(ctx context.Context, userID uint) *gorm.DB {
	// IMPORTANT: assumes boards table is "boards" and has columns id (uuid) and user_id (uint)
	// and pools table is "board_pools" with column board_id
	return r.db.WithContext(ctx).
		Model(&models.BoardPool{}).
		Joins("JOIN boards ON boards.id = board_pools.board_id").
		Where("boards.user_id = ?", userID)
}

// Create a pool only if board belongs to user.
// When explicitPosition is true the new lane is inserted at p.Position and
// all existing lanes at that position or higher are shifted up by one so
// there are no duplicates. When false the lane is appended after the current
// last position (max + 1), or placed at 0 for an empty board.
func (r *BoardPoolRepo) Create(ctx context.Context, userID uint, p *models.BoardPool, explicitPosition bool) error {
	var cnt int64
	if err := r.db.WithContext(ctx).
		Table("boards").
		Where("id = ? AND user_id = ?", p.BoardID, userID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return gorm.ErrRecordNotFound
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if explicitPosition {
			// Shift all lanes at the target position and above to make room.
			if err := tx.Model(&models.BoardPool{}).
				Where("board_id = ? AND position >= ?", p.BoardID, p.Position).
				UpdateColumn("position", gorm.Expr("position + 1")).Error; err != nil {
				return err
			}
		} else {
			// Append: place after the current last lane.
			var maxPos *int
			if err := tx.Model(&models.BoardPool{}).
				Where("board_id = ?", p.BoardID).
				Select("MAX(position)").
				Scan(&maxPos).Error; err != nil {
				return err
			}
			if maxPos != nil {
				p.Position = *maxPos + 1
			} else {
				p.Position = 0
			}
		}
		return tx.Create(p).Error
	})
}

func (r *BoardPoolRepo) ListByBoardID(ctx context.Context, userID uint, boardID uuid.UUID) ([]models.BoardPool, error) {
	var items []models.BoardPool
	err := r.scoped(ctx, userID).
		Where("board_pools.board_id = ?", boardID).
		Order("board_pools.position ASC").
		Order("board_pools.created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *BoardPoolRepo) GetByID(ctx context.Context, userID uint, boardID, id uuid.UUID) (*models.BoardPool, error) {
	var p models.BoardPool
	err := r.scoped(ctx, userID).
		Where("board_pools.board_id = ? AND board_pools.id = ?", boardID, id).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *BoardPoolRepo) Update(ctx context.Context, userID uint, boardID, id uuid.UUID, updates map[string]any) (*models.BoardPool, error) {
	tx := r.scoped(ctx, userID).
		Where("board_pools.board_id = ? AND board_pools.id = ?", boardID, id)

	res := tx.Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, userID, boardID, id)
}

func (r *BoardPoolRepo) Delete(ctx context.Context, userID uint, boardID, id uuid.UUID) (bool, error) {
	tx := r.scoped(ctx, userID).
		Where("board_pools.board_id = ? AND board_pools.id = ?", boardID, id)

	res := tx.Delete(&models.BoardPool{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
