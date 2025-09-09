package sqlfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SQLFile представляет собой SQL-файл с его именем и содержимым
type SQLFile struct {
	Title   string
	Content string
	Path    string
}

// SearchMode определяет режим поиска
type SearchMode int

const (
	AllSQLFiles    SearchMode = iota // Все SQL файлы, кроме миграций
	MigrationsOnly                   // Только миграции
	SpecificFiles                    // Конкретные файлы
)

// SearchConfig конфигурация поиска
type SearchConfig struct {
	RootPath          string
	Mode              SearchMode
	MigrationsPath    string   // Путь к папке миграций (для MigrationsOnly)
	SpecificFileNames []string // Имена файлов для поиска (для SpecificFiles)
	EnableIgnoreList  bool
	IgnoreFiles       []string
}

// normalizePath нормализует путь, убирая слеш в конце и добавляя ./ если нужно
func normalizePath(path string) string {
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "./") && !filepath.IsAbs(path) {
		path = "./" + path
	}
	return path
}

// hasSameBaseName проверяет, есть ли файлы с одинаковыми именами
func hasSameBaseName(files []SQLFile, filename string) bool {
	for _, file := range files {
		if file.Title == filename {
			return true
		}
	}
	return false
}

// makeUniqueTitle создает уникальное имя файла при конфликте
func makeUniqueTitle(files []SQLFile, filename, fullPath string) string {
	if !hasSameBaseName(files, filename) {
		return filename
	}

	dir := filepath.Dir(fullPath)
	relDir := strings.TrimPrefix(dir, filepath.Dir(fullPath)+"/")
	relDir = strings.ReplaceAll(relDir, "/", "_")

	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	ext := filepath.Ext(filename)

	return fmt.Sprintf("%s_%s%s", name, relDir, ext)
}

// CollectSQLFiles запускает сбор файлов в зависимости от режима
func CollectSQLFiles(config SearchConfig) ([]SQLFile, error) {
	switch config.Mode {
	case MigrationsOnly:
		return collectMigrations(config)
	case SpecificFiles:
		return collectSpecificFiles(config)
	default:
		return collectAllSQLFiles(config)
	}
}

// collectMigrations собирает только файлы миграций
func collectMigrations(config SearchConfig) ([]SQLFile, error) {
	var files []SQLFile

	migrationsPath := config.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = filepath.Join(config.RootPath, "migrations")
	} else {
		migrationsPath = normalizePath(migrationsPath)
		if !filepath.IsAbs(migrationsPath) {
			migrationsPath = filepath.Join(config.RootPath, migrationsPath)
		}
	}

	err := filepath.Walk(migrationsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			if config.EnableIgnoreList && contains(config.IgnoreFiles, info.Name()) {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			title := makeUniqueTitle(files, filepath.Base(path), path)
			files = append(files, SQLFile{
				Title:   title,
				Content: string(content),
				Path:    path,
			})
		}
		return nil
	})

	return files, err
}

// collectSpecificFiles собирает только указанные файлы
func collectSpecificFiles(config SearchConfig) ([]SQLFile, error) {
	var files []SQLFile

	targetFiles := make(map[string]bool)
	for _, filename := range config.SpecificFileNames {
		targetFiles[strings.ToLower(filename)] = true
	}

	err := filepath.Walk(config.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			if strings.Contains(strings.ToLower(path), "migrations") ||
				strings.Contains(strings.ToLower(info.Name()), "migration") {
				return nil
			}

			if targetFiles[strings.ToLower(info.Name())] {
				if config.EnableIgnoreList && contains(config.IgnoreFiles, info.Name()) {
					return nil
				}

				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				title := makeUniqueTitle(files, filepath.Base(path), path)
				files = append(files, SQLFile{
					Title:   title,
					Content: string(content),
					Path:    path,
				})
			}
		}
		return nil
	})

	return files, err
}

// collectAllSQLFiles собирает все SQL файлы, кроме миграций
func collectAllSQLFiles(config SearchConfig) ([]SQLFile, error) {
	var files []SQLFile

	err := filepath.Walk(config.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			if strings.Contains(strings.ToLower(path), "migrations") ||
				strings.Contains(strings.ToLower(info.Name()), "migration") {
				return nil
			}

			if config.EnableIgnoreList && contains(config.IgnoreFiles, info.Name()) {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			title := makeUniqueTitle(files, filepath.Base(path), path)
			files = append(files, SQLFile{
				Title:   title,
				Content: string(content),
				Path:    path,
			})
		}
		return nil
	})

	return files, err
}

// contains проверяет наличие элемента в срезе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
