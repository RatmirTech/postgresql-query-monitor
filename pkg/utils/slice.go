package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dreadew/go-common/pkg/interfaces"
	"github.com/dreadew/go-common/pkg/models"
)

// Фильтрация слайса
func Filter[T interfaces.Entity](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Фильтрация слайса в канал
func FilterToChan[T interfaces.Entity](slice []T, predicate func(T) bool) <-chan T {
	var ch = make(chan T)

	go func() {
		defer close(ch)

		for _, item := range slice {
			if predicate(item) {
				ch <- item
			}
		}
	}()

	return ch
}

// Пагинация сущностей
func Paginate[T interfaces.Entity](slice []T, pagination *models.Pagination) []T {
	if pagination == nil {
		return slice
	}

	return slice[pagination.Page*pagination.PageSize : (pagination.Page+1)*pagination.PageSize]
}

func FilterEntities[T interfaces.Entity](entities []T, filters []models.FilterField) ([]T, error) {
	var result []T

	for _, entity := range entities {
		v := reflect.ValueOf(entity)

		for _, filter := range filters {
			field := v.FieldByName(filter.Field)

			if !field.IsValid() {
				return nil, fmt.Errorf("invalid field name: %s", filter.Field)
			}

			fieldValue := field.String()

			switch filter.Operator {
			case models.OpEquals:
				if fieldValue == filter.Value {
					result = append(result, entity)
				}
			case models.OpNotEquals:
				if fieldValue != filter.Value {
					result = append(result, entity)
				}
			case models.OpContains:
				if strings.Contains(fieldValue, filter.Value) {
					result = append(result, entity)
				}
			case models.OpStartsWith:
				if strings.HasPrefix(fieldValue, filter.Value) {
					result = append(result, entity)
				}
			case models.OpEndsWith:
				if strings.HasSuffix(fieldValue, filter.Value) {
					result = append(result, entity)
				}
			default:
				return nil, fmt.Errorf("invalid operator: %s", filter.Operator)
			}
		}
	}

	return result, nil
}
