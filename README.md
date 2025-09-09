# 🐘 PostgreSQL Query Monitor CLI (`pgmon`)

CLI-инструмент для сбора и анализа конфигурации PostgreSQL, SQL-файлов и системных метрик. Интегрируется с Vault и внешним API анализа.

---

## 🚀 Установка и настройка

### 1. Переменные окружения

Создайте файл `.env` в корне проекта со следующими переменными:

VAULT_TOKEN=your-vault-token-here
REVIEW_API_URL=https://your-review-api.example.com
ENVIRONMENT=production
VAULT_ADDR=http://localhost:8200


> 💡 `godotenv` автоматически загрузит этот файл при запуске.


---

## 🧭 Команды

### `pgmon csi` — Сбор информации о сервере и конфигурации

Собирает данные о сервере PostgreSQL через Vault и отправляет их на анализ (без метрик).

#### 🏷️ Флаги

| Флаг | Описание | Обязательный | По умолчанию |
|------|----------|--------------|--------------|
| `--vp` | Путь в Vault, где хранятся данные подключения к PostgreSQL | ✅ Да | — |
| `--st` | Является ли задача запущенной по расписанию (scheduler task) | ❌ Нет | `false` |

#### 📌 Примеры

pgmon csi --vp="secret/data/postgres/prod"

pgmon csi --vp="secret/data/postgres/staging" --st=true

---

### `pgmon csf` — Сбор и анализ SQL-файлов

Сканирует директорию на наличие SQL-файлов, фильтрует по режиму и отправляет на анализ.


#### 🏷️ Флаги

| Флаг | Описание | Обязательный | По умолчанию |
|------|----------|--------------|--------------|
| `--dir` | Директория для сканирования | ❌ Нет | `.` (текущая) |
| `--mode` | Режим поиска: `all`, `migrations`, `specific` | ❌ Нет | `all` |
| `--vp` | Поддиректория миграций (используется, если `--mode=migrations`) | ❌ Нет | — |
| `--files` | Список конкретных файлов (используется, если `--mode=specific`) | ❌ Нет | `[]` |
| `--enable-ignore` | Включить игнорирование файлов из списка `--ignore` | ❌ Нет | `false` |
| `--ignore` | Список файлов для игнорирования (имена или пути) | ❌ Нет | `[]` |

#### 📌 Примеры

pgmon csf

pgmon csf --mode=migrations --vp="migrations"

pgmon csf --mode=specific --files="init.sql,users.sql"

pgmon csf --enable-ignore --ignore="deprecated.sql,old_views.sql"

pgmon csf --dir="./sql" --mode=migrations --vp="db/migrations" --enable-ignore --ignore="rollback.sql"

---

### `pgmon csm` — Сбор системных метрик и информации о сервере

Собирает системные метрики (CPU, RAM, диски и т.д.) + информацию о сервере PostgreSQL и отправляет на анализ.

#### 🏷️ Флаги

| Флаг | Описание | Обязательный | По умолчанию |
|------|----------|--------------|--------------|
| `--vp` | Путь в Vault, где хранятся данные подключения к PostgreSQL | ✅ Да | — |
| `--st` | Является ли задача запущенной по расписанию (scheduler task) | ❌ Нет | `false` |

#### 📌 Примеры

pgmon csm --vp="secret/data/postgres/prod"

pgmon csm --vp="secret/data/postgres/prod" --st=true

---

## 🛠️ Разработка

### Раскомментирование отправки SQL-файлов

В команде `csf` отправка на API закомментирована. Чтобы активировать:

1. Раскомментируйте блок кода, начинающийся с:
   // Создаём клиента
   // ctx := context.Background()
   // apiClient := client.NewClient(cfg.ReviewAPI.URL)

2. Убедитесь, что `ReviewAPI.URL` загружается из конфига или `.env`.

3. Убедитесь, что структуры `models.QueryReviewRequest`, `models.BatchReviewRequest`, `models.MigrationReviewRequest` существуют и соответствуют API.

---

## 🧪 Примеры полного использования

pgmon csi --vp="secret/data/pg/main" --st=true

pgmon csf --dir="./sql" --mode=all --enable-ignore --ignore="temp.sql"

pgmon csm --vp="secret/data/pg/main" --st=false

---

## ❗ Ошибки и диагностика

- `Vault path is required` — не указан флаг `--vp` для команд, требующих доступ к Vault.
- `Failed to load config` — проверьте наличие `.env` файла и корректность переменных.
- `Failed to create Vault client` — проверьте доступность Vault и корректность токена.
- `No SQL files found` — в указанной директории нет `.sql` файлов, соответствующих фильтрам.

---

## 📦 Зависимости

- github.com/spf13/cobra — CLI фреймворк.
- github.com/hashicorp/vault/api — клиент Vault.
- github.com/joho/godotenv — загрузка `.env`.
- Внутренние пакеты: `collectors`, `serverinfo`, `sqlfiles`, `review`, `config`.

 ✅ Готово к использованию в CI/CD, скриптах мониторинга и ручном запуске администраторами БД.