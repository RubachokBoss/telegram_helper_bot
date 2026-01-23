# 🌐 Web API Service

REST API микросервис для веб-пользователей с JWT аутентификацией и интеграцией с задачами через gRPC. Использует стандартную библиотеку `net/http` без дополнительных фреймворков.

## 📋 Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Browser   │───▶│   REST API      │───▶│   gRPC Client   │
│   / Mobile App  │    │   (net/http)    │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   JWT Auth      │    │   PostgreSQL    │    │   gRPC          │
│   Service       │    │   (Web Users)   │    │   (Tasks)       │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Быстрый старт

### Запуск через Docker Compose
```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов веб-API
docker-compose logs web-api
```

### Локальный запуск
```bash
# Перейти в директорию сервиса
cd cmd/web-api

# Запустить сервис
go run main.go
```

## 🔐 Аутентификация

Сервис использует JWT токены для аутентификации пользователей.

### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Ответ:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "2024-01-22T10:00:00Z"
  }
}
```

### Вход в систему
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

## 📋 API Endpoints

### Аутентификация

#### `POST /api/v1/auth/register`
Регистрация нового пользователя.

**Тело запроса:**
```json
{
  "email": "string",
  "password": "string (min 6 chars)",
  "first_name": "string",
  "last_name": "string (optional)"
}
```

#### `POST /api/v1/auth/login`
Вход в систему.

**Тело запроса:**
```json
{
  "email": "string",
  "password": "string"
}
```

### Защищенные endpoints (требуют JWT токен)

#### `GET /api/v1/user/profile`
Получение профиля текущего пользователя.

**Заголовки:**
```
Authorization: Bearer <jwt_token>
```

### Работа с задачами

#### `POST /api/v1/tasks`
Создание новой задачи.

**Заголовки:**
```
Authorization: Bearer <jwt_token>
```

**Тело запроса:**
```json
{
  "text": "Описание задачи"
}
```

#### `GET /api/v1/tasks/user/:userId`
Получение задач, назначенных пользователю.

#### `GET /api/v1/tasks/owner/:ownerId`
Получение задач, созданных пользователем (владельцем).

#### `PUT /api/v1/tasks/:taskId/assign/:userId`
Назначение задачи пользователю.

#### `PUT /api/v1/tasks/:taskId/unassign`
Снятие назначения задачи.

#### `DELETE /api/v1/tasks/:taskId`
Выполнение и удаление задачи.

## 🔧 Конфигурация

### Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `WEB_API_PORT` | Порт для REST API | `:8080` |
| `JWT_SECRET` | Секрет для JWT токенов | `your-secret-key-change-in-production` |
| `GRPC_HOST` | Адрес gRPC сервера задач | `notes-service:50051` |
| `POSTGRES_HOST` | Хост PostgreSQL | `localhost` |
| `POSTGRES_PORT` | Порт PostgreSQL | `5432` |

### Файл конфигурации

```yaml
# config/config.yaml
web_api:
  port: ":8080"

jwt:
  secret: "your-secret-key-change-in-production"

grpc:
  port: ":50051"
```

## 📊 Примеры использования

### Полный сценарий работы

```bash
# 1. Регистрация
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"pass123","first_name":"John"}'

# 2. Вход (получение токена)
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"pass123"}' \
  | jq -r '.token')

# 3. Создание задачи
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"Изучить Go"}'

# 4. Получение своих задач
curl -X GET http://localhost:8080/api/v1/tasks/owner/user-id-here \
  -H "Authorization: Bearer $TOKEN"
```

## 🔍 Мониторинг и отладка

### Логи
```bash
# Просмотр логов
docker-compose logs -f web-api

# Логи аутентификации
2024/01/22 10:00:00 🔐 Attempting login for user: john@example.com
2024/01/22 10:00:00 ✅ User logged in successfully: john@example.com
```

### Проверка здоровья
```bash
# Проверка доступности
curl http://localhost:8080/health

# Проверка подключения к gRPC
curl http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN"
```

## 🏗️ Структура проекта

```
cmd/web-api/
├── main.go                    # Точка входа
internal/
├── domain/
│   ├── user.go                # Доменная модель веб-пользователя
│   └── task.go                # Доменная модель задачи (общая)
├── service/
│   ├── auth_service.go        # Сервис аутентификации
│   └── task_service.go        # Сервис задач (общий)
├── repository/postgres/
│   ├── web_user_repository.go # Репозиторий пользователей
│   └── task_repository.go     # Репозиторий задач (общий)
└── delivery/rest/
    ├── server.go              # REST сервер (net/http)
    ├── auth_handlers.go       # Обработчики аутентификации
    └── task_handlers.go       # Обработчики задач
```

## 📦 Зависимости

Проект использует минимальный набор зависимостей:

- **Стандартная библиотека**: `net/http`, `encoding/json`, `context`
- **JWT**: `github.com/golang-jwt/jwt/v5`
- **Криптография**: `golang.org/x/crypto/bcrypt`
- **База данных**: `github.com/lib/pq`
- **gRPC**: `google.golang.org/grpc`
- **Конфигурация**: `gopkg.in/yaml.v2`

## 🔒 Безопасность

- **Пароли:** Хэшируются с помощью bcrypt
- **JWT токены:** Действительны 24 часа
- **CORS:** Настроен для веб-клиентов
- **Валидация:** Проверка входных данных

## 🚀 Развертывание

### Production рекомендации

1. **Измените JWT секрет:**
   ```bash
   export JWT_SECRET="your-production-secret-key"
   ```

2. **Настройте HTTPS** (рекомендуется использовать reverse proxy)

3. **Ограничьте CORS** origins в production

4. **Настройте логирование** в файлы или внешние системы

## 📝 Разработка

### Добавление новых endpoints

1. Добавить маршрут в `server.go` в метод `setupRoutes()`
2. Создать обработчик в соответствующем файле (`auth_handlers.go`, `task_handlers.go`)
3. Обновить документацию

### Тестирование

```bash
# Запуск всех сервисов
docker-compose up -d

# Регистрация пользователя
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'

# Получение токена
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }' | jq -r '.token')

# Создание задачи
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text": "Тестовая задача"}'

# Получение профиля пользователя
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🎯 Следующие шаги

- [ ] Добавить Swagger документацию API
- [ ] Реализовать refresh tokens
- [ ] Добавить rate limiting
- [ ] Создать веб-интерфейс (React/Vue)
- [ ] Добавить email верификацию
- [ ] Реализовать роли пользователей (admin/user)

**API готов к использованию!** 🚀
