package errors

import (
	"fmt"
	"strings"
)

// Валидационная ошибка в поле
type FieldError struct {
	Name  string
	Error string
}

// Ошибка валидации
type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	var sb strings.Builder

	for _, fe := range e.Errors {
		sb.WriteString(fmt.Sprintf("Поле \"%s\", ошибка \"%s\"\n", fe.Name, fe.Error))
	}

	return sb.String()
}

// Сервисная ошибка
type ServiceError struct {
	Details string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("service error: %s", e.Details)
}

// Ошибка, отображаемая пользователю
type UserVisibleError struct {
	Details string
}

func (e *UserVisibleError) Error() string {
	return e.Details
}
