## Reward System

Простой сервис на Go для управления пользователями, их целевыми заданиями (ввод реферального кода, подписка на TG/X и т.п.) и начислением поинтов. Есть JWT‑авторизация, лидерборд, PostgreSQL с миграциями и Docker Compose для локального запуска.

Репозиторий: `https://github.com/NetPo4ki/reward-system`

### Возможности
- Создание пользователя (тестовый эндпоинт) и выдача JWT токена
- Выполнение заданий с начислением поинтов (один раз на пользователя для каждого задания)
- Установка реферера (один раз, нельзя указать самого себя)
- Просмотр статуса пользователя: профиль, баланс, список выполненных задач
- Лидерборд по балансу
- Авторизация через Bearer JWT для всех бизнес‑эндпоинтов

### Технологии
- Go 1.23, `chi` (HTTP роутер)
- JWT (Bearer) — `github.com/golang-jwt/jwt/v5`
- PostgreSQL + SQL миграции (`golang-migrate`)
- Docker и Docker Compose

### Структура проекта (кратко)
- `cmd/server` — входная точка приложения
- `internal/config` — загрузка конфигурации из env
- `internal/db` — подключение к Postgres (pgx pool)
- `internal/models` — доменные модели
- `internal/repo` — репозитории: users, tasks, user_tasks
- `internal/auth` — JWT helper’ы и middleware
- `internal/handlers` — HTTP‑обработчики (auth, users)
- `internal/server` — сборка роутера и middleware
- `migrations` — SQL‑миграции
- `deployments/docker/Dockerfile` — Dockerfile
- `docker-compose.yml` — docker‑компоуз стэк

### Схема данных
- `users`
  - `id` (bigserial, PK)
  - `username` (text, unique)
  - `referrer_id` (bigint, FK -> users.id, nullable)
  - `created_at` (timestamptz, default now())
- `tasks`
  - `id` (bigserial, PK)
  - `code` (text, unique) — напр., `subscribe_tg`, `follow_twitter`, `enter_referral`
  - `name` (text)
  - `points` (int, >= 0)
  - `active` (bool, default true)
  - `created_at` (timestamptz, default now())
- `user_tasks`
  - `user_id` (bigint, FK -> users.id)
  - `task_id` (bigint, FK -> tasks.id)
  - `points_awarded` (int, >= 0)
  - `completed_at` (timestamptz, default now())
  - `UNIQUE (user_id, task_id)` — одно выполнение задачи на пользователя

Баланс пользователя = `SUM(points_awarded)` из `user_tasks`.

### Переменные окружения (.env / .env.example)
Скопируйте `.env.example` в `.env`.

```bash
PORT=8080
JWT_SECRET=change_me

POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=rewards

DATABASE_URL=postgres://postgres:postgres@db:5432/rewards?sslmode=disable
```

### Запуск через Docker Compose
```bash
cp .env.example .env
docker compose up --build
```
Стек поднимет:
- `db` — PostgreSQL с healthcheck
- `migrate` — применит SQL‑миграции к БД
- `app` — Go‑сервис, стартует после успешных миграций

Проверка:
```bash
curl http://localhost:8080/healthz
# ok
```

### Аутентификация
- Публичный эндпоинт: `POST /auth/signup` — создаёт пользователя и выдаёт JWT.
- Все остальные бизнес‑эндпоинты — под Bearer JWT.

### Эндпоинты
- `POST /auth/signup` — регистрация и выдача JWT.
- `GET /users/{id}/status` — профиль, баланс, выполненные задания (только для владельца токена).
- `GET /users/leaderboard?limit=50` — топ пользователей по балансу.
- `POST /users/{id}/task/complete` — выполнение задачи по `task_code` (только для владельца токена).
- `POST /users/{id}/referrer` — установка `referrer_id` (один раз, не сам себя).

Стандартная ошибка (JSON):
```json
{ "error": "message", "code": "optional_code" }
```
Коды: 400 (bad request), 401 (unauthorized), 403 (forbidden), 404 (not found), 409 (conflict), 500 (internal).

### Postman
1) Импортируйте Environment `reward-system-local` (переменные: `baseUrl`, `token`, `userId`, `referrerId`).  
2) Импортируйте Collection `Reward System API`.  
3) Запустите `Auth / Signup (issue JWT)` — скрипт сохранит `token` и `userId` в окружение.  
4) Выполняйте запросы из раздела `Users`: `status`, `leaderboard`, `task/complete`, `referrer`.

### Примечания
- SQL миграции seed’ят базовые задания: `subscribe_tg`, `follow_twitter`, `enter_referral`.

