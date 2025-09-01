package interfaces

import "time"

// Базовый интерфейс для DTO
type Dto interface {
	GetId() string
}

// Интерфейс для DTO с аудитом
type AuditableDto interface {
	Dto
	GetCreatedAt() time.Time
	GetUpdatedAt() *time.Time
}

// Интерфейс для DTO с мягким удалением
type SoftDeleteDto interface {
	AuditableDto
	GetIsDeleted() bool
	GetDeletedAt() *time.Time
}

// Интерфейс для DTO с константными значениями
type ConstantDto interface {
	SoftDeleteDto
	GetName() string
}
