package models

import "time"

// Базовая сущность
type BaseEntity struct {
	ID string `json:"id" db:"id"`
}

func (b *BaseEntity) GetID() string {
	return b.ID
}

// Базовая сущность с аудитом
type BaseAuditableEntity struct {
	BaseEntity
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

func (b *BaseAuditableEntity) GetCreatedAt() time.Time {
	return b.CreatedAt
}

// Базовая сущность с мягким удалением
type BaseSoftDeleteEntity struct {
	BaseAuditableEntity
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`
	DeletedAt time.Time `json:"deleted_at" db:"deleted_at"`
}

func (b *BaseSoftDeleteEntity) GetIsDeleted() bool {
	return b.IsDeleted
}

// Базовая сущность с константными значениями
type BaseConstantEntity struct {
	BaseSoftDeleteEntity
	Name string `json:"name" db:"name"`
}

func (b *BaseConstantEntity) GetName() string {
	return b.Name
}
