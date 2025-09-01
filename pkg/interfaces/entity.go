package interfaces

import "time"

// Интерфейс базовой сущности
type Entity interface {
	GetID() string
}

// Интерфейс сущности с аудитом
type AuditableEntity interface {
	Entity
	GetCreatedAt() time.Time
	GetUpdatedAt() *time.Time
}

// Интерфейс сущности с мягким удалением
type SoftDeleteEntity interface {
	Entity
	AuditableEntity
	GetIsDeleted() bool
	GetDeletedAt() time.Time
}

// Интерфейс сущности с константными значениями
type ConstantEntity interface {
	SoftDeleteEntity
	GetName() string
}
