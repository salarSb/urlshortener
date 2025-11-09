package shortener

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNotFound = errors.New("url not found")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&URL{})
}

func (r *Repository) Create(ctx context.Context, u *URL) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *Repository) GetByShortCode(ctx context.Context, code string) (*URL, error) {
	var u URL
	err := r.db.WithContext(ctx).Where("short_code = ?", code).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) IncrementClick(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&URL{}).
		Where("id = ?", id).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", 1)).
		Error
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
