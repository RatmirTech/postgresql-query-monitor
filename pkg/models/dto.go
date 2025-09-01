package models

import "time"

// Базовая структура для DTO
type BaseDto struct {
	Id string `json:"id"`
}

func (b *BaseDto) GetId() string {
	return b.Id
}

// Базовая структура для DTO с аудитом
type BaseAuditableDto struct {
	BaseDto
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func (b *BaseAuditableDto) GetCreatedAt() time.Time {
	return b.CreatedAt
}

func (b *BaseAuditableDto) GetUpdatedAt() *time.Time {
	return b.UpdatedAt
}

// Базовая структура для DTO с мягким удалением
type BaseSoftDeleteDto struct {
	BaseAuditableDto
	IsDeleted bool       `json:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (b *BaseSoftDeleteDto) GetIsDeleted() bool {
	return b.IsDeleted
}

func (b *BaseSoftDeleteDto) GetDeletedAt() *time.Time {
	return b.DeletedAt
}

// Базовая структура для DTO с константными значениями
type BaseConstantDto struct {
	BaseSoftDeleteDto
	Name string `json:"name"`
}

func (b *BaseConstantDto) GetName() string {
	return b.Name
}

// Структура для пагинации
type Pagination struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

// Структура для фильтрации
type FilterField struct {
	Field    string   `json:"field" query:"field"`
	Operator Operator `json:"operator" query:"operator"`
	Value    string
}
