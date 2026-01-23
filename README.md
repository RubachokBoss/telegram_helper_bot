# Telegram Task Manager Bot 🤖

Микросервисное приложение для управления задачами через Telegram с использованием Go, gRPC, PostgreSQL и Docker.

## 📋 Оглавление

- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Технологический стек](#технологический-стек)
- [Быстрый старт](#быстрый-старт)
- [Настройка окружения](#настройка-окружения)
- [Команды бота](#команды-бота)
- [API документация](#api-документация)
- [Разработка](#разработка)
- [Деплой](#деплой)
- [Безопасность](#безопасность)
- [Troubleshooting](#troubleshooting)

## 🚀 Возможности

### Для пользователей:
- ✅ **Создание задач** - быстрое создание новых задач через Telegram
- ✅ **Назначение исполнителей** - делегирование задач другим пользователям
- ✅ **Просмотр задач** - разделение на "мои задачи" и "созданные мной"
- ✅ **Отслеживание статуса** - мониторинг выполнения задач
- ✅ **Завершение задач** - отметка о выполнении и автоматическое удаление

### Для бизнеса:
- 🏢 **Командная работа** - идеально для небольших команд до 10 человек
- 📊 **Учет задач** - автоматическое ведение истории задач
- 🔔 **Уведомления** - мгновенные уведомления через Telegram
- 📱 **Мобильность** - работа из любого места через мобильное приложение
- 🎯 **Простота** - интуитивно понятный интерфейс без обучения

### Использование случаев:
- **Персональное использование** - управление личными задачами
- **Семейные дела** - распределение домашних обязанностей
- **Учебные группы** - координация учебных проектов
- **Небольшие команды** - управление рабочими задачами
- **Фриланс** - отслеживание проектов для клиентов

## 🏗️ Архитектура

### Компоненты системы:

```
┌─────────────────┐    gRPC    ┌──────────────────┐    HTTP     ┌──────────────┐
│   Telegram Bot  │ ◄───────── │  Notes Service   │ ◄───────── │  PostgreSQL  │
│  (Go + TB API)  │            │ (Go + gRPC)      │            │   Database   │
└─────────────────┘            └──────────────────┘            └──────────────┘
│                               │                               │
│ Docker Container              │ Docker Container              │ Docker Container
└───────────────────────────────┼───────────────────────────────┘
│                               │
│                               │ Redis Cache
│                               │ (In-memory)
└───────────────────────────────┼───────────────────────────────
│
┌─────┴─────┐
│ Docker    │
│ Network   │
└───────────┘
```

### Микросервисы:

1. **Telegram Bot** (`telegram-bot`)
   - Обработка сообщений от пользователей
   - Валидация команд
   - Форматирование ответов
   - Коммуникация с gRPC сервером

2. **Notes Service** (`notes-service`) 
   - Бизнес-логика управления задачами
   - gRPC API сервер
   - Работа с базой данных
   - Валидация данных

3. **PostgreSQL Database**
   - Хранение задач и метаданных
   - Индексы для быстрого поиска
   - Надежное хранение данных

4. **Redis Cache**
   - Кэширование часто запрашиваемых данных
   - Ускорение чтения списков задач пользователей
   - Автоматическая инвалидация при изменениях

### Поток данных:

```
Пользователь → Telegram → Bot API → Telegram Bot → gRPC → Notes Service → Redis/PostgreSQL
│
Ответ ← Форматирование ← Telegram Bot ← gRPC ← Notes Service ← Кэш/База данных ←┘
```

**Кэширование:**
- **Cache-aside паттерн**: проверка Redis → PostgreSQL (если нет в кэше)
- **TTL стратегия**: 10 мин для списков задач, 30 мин для индивидуальных задач
- **Инвалидация**: автоматическая очистка кэша при изменениях данных

## 🛠 Технологический стек

### Бэкенд:
- **Go 1.19** - основной язык программирования
- **gRPC** - межсервисная коммуникация
- **Protocol Buffers** - сериализация данных
- **PostgreSQL** - реляционная база данных
- **Docker** - контейнеризация
- **Docker Compose** - оркестрация контейнеров

### Библиотеки:
- `go-telegram-bot-api` - интеграция с Telegram API
- `lib/pq` - драйвер PostgreSQL для Go
- `google.golang.org/grpc` - gRPC фреймворк
- `google/uuid` - генерация уникальных идентификаторов
- `gopkg.in/yaml.v2` - парсинг YAML конфигураций
- `github.com/redis/go-redis/v9` - Redis клиент для кэширования

### Инфраструктура:
- **Многостадийные Docker образы** - оптимизация размера
- **Docker сети** - изолированная коммуникация
- **Health checks** - мониторинг состояния сервисов
- **Retry логика** - устойчивость к временным сбоям

## 🚀 Быстрый старт

### Предварительные требования:
- Docker 20.10+
- Docker Compose 2.0+
- Telegram аккаунт

### 1. Клонирование репозитория:
```bash
git clone https://github.com/your-username/telegram-task-bot.git
cd telegram-task-bot
```

### 2. Настройка окружения:
```bash
# Копируем пример конфигурации
cp .env.example .env

# Редактируем файл .env
nano .env  # или используйте любой текстовый редактор
```

### 3. Создание Telegram бота:
1. Откройте Telegram и найдите @BotFather
2. Отправьте команду `/newbot`
3. Следуйте инструкциям для создания бота
4. Скопируйте полученный токен в файл `.env`

### 4. Запуск приложения:
```bash
# Запуск всех сервисов
docker-compose up --build

# Или в фоновом режиме
docker-compose up -d --build
```

### 5. Проверка работы:
```bash
# Просмотр логов
docker-compose logs -f

# Проверка статуса контейнеров
docker-compose ps
```

### 6. Начало работы с ботом:
1. Найдите вашего бота в Telegram по имени
2. Отправьте команду `/tasks` для просмотра доступных команд
3. Создайте первую задачу с помощью `/new`

## ⚙️ Настройка окружения

Файл `.env` должен содержать:

```env
# Database Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=task_manager

# Telegram Bot
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMnOpQRstUvWxYz123456789

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Service Configuration (опционально)
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
GRPC_HOST=notes-service:50051
```

## 💬 Команды бота

### Основные команды:

| Команда | Описание | Пример |
|---------|-----------|---------|
| `/tasks` | Показать все доступные команды | `/tasks` |
| `/new` | Создать новую задачу | `/new Купить молоко` |
| `/my` | Показать задачи назначенные на меня | `/my` |
| `/owner` | Показать задачи созданные мной | `/owner` |

### Команды управления задачами:

| Действие | Формат | Описание |
|----------|---------|-----------|
| Назначить задачу | `/assign_<task_id>` | Взять задачу в работу |
| Снять задачу | `/unassign_<task_id>` | Отказаться от задачи |
| Завершить задачу | `/resolve_<task_id>` | Отметить как выполненную |

### Пример рабочего процесса:

```
👤 Пользователь: /new Подготовить отчет по проекту
🤖 Бот: Задача создана! ID: abc123

👤 Пользователь: /owner
🤖 Бот: Задачи созданные вами:
       Задача #1: Подготовить отчет по проекту [ID: abc123]

👤 Коллега: /assign_abc123  
🤖 Бот: Задача назначена! Исполнитель: colleague

👤 Коллега: /resolve_abc123
🤖 Бот: Задача выполнена и удалена!
```

## 🔌 API документация

### gRPC сервис - TaskService

#### Создание задачи:
```protobuf
rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse);

message CreateTaskRequest {
    string text = 1;
    string owner_id = 2;
}

message CreateTaskResponse {
    Task task = 1;
}
```

#### Назначение задачи:
```protobuf
rpc AssignTask(AssignTaskRequest) returns (AssignTaskResponse);

message AssignTaskRequest {
    string task_id = 1;
    string user_id = 2;
}
```

#### Получение задач пользователя:
```protobuf
rpc GetUserTasks(GetUserTasksRequest) returns (GetUserTasksResponse);

message GetUserTasksRequest {
    string user_id = 1;
}
```

### Модель данных Task:

```go
type Task struct {
    ID         string    `json:"id"`
    Text       string    `json:"text"`
    OwnerID    string    `json:"owner_id"`    // Создатель задачи
    AssignedID string    `json:"assigned_id"` // Исполнитель
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

## 🛠 Разработка

### Локальная разработка без Docker:

```bash
# Установка зависимостей
go mod download

# Запуск базы данных
docker-compose up postgres -d

# Запуск notes-service
cd cmd/notes-service && go run main.go

# Запуск telegram-bot (в отдельном терминале)
cd cmd/telegram-bot && go run main.go
```

### Структура проекта:

```
telegram_helper_bot/
├── cmd/                    # Точки входа приложения
│   ├── notes-service/      # gRPC сервер
│   └── telegram-bot/       # Telegram бот
├── config/                 # Конфигурация
├── internal/               # Внутренние пакеты
│   ├── delivery/           # Обработчики запросов
│   │   ├── grpc/           # gRPC хендлеры
│   │   └── telegram/       # Telegram хендлеры
│   ├── domain/             # Бизнес-сущности
│   ├── repository/         # Работа с данными
│   └── service/            # Бизнес-логика
├── pkg/                    # Внешние пакеты
│   └── pb/                 # Protobuf файлы
├── docker-compose.yml      # Docker композ
├── Dockerfile.notes        # Образ notes-service
├── Dockerfile.bot          # Образ telegram-bot
└── init.sql               # Инициализация БД
```

### Добавление новых функций:

1. **Новая gRPC endpoint**:
    - Добавить метод в `pkg/pb/task.proto`
    - Сгенерировать код: `protoc --go_out=. --go-grpc_out=. pkg/pb/*.proto`
    - Реализовать хендлер в `internal/delivery/grpc/handler.go`
    - Добавить метод в сервис `internal/service/task_service.go`

2. **Новая команда бота**:
    - Добавить обработчик в `internal/delivery/telegram/bot.go`
    - Создать функцию в `internal/delivery/telegram/handlers.go`

## 🚀 Деплой

### Продакшен конфигурация:

Создай `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    env_file: .env.prod
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app-network
    restart: always

  notes-service:
    build:
      context: .
      dockerfile: Dockerfile.notes
    env_file: .env.prod
    environment:
      - POSTGRES_HOST=postgres
    networks:
      - app-network
    restart: always
    deploy:
      replicas: 2

  telegram-bot:
    build:
      context: .
      dockerfile: Dockerfile.bot
    env_file: .env.prod
    environment:
      - GRPC_HOST=notes-service:50051
    networks:
      - app-network
    restart: always

networks:
  app-network:

volumes:
  postgres_data:
```

### Деплой на сервер:

```bash
# Копируем файлы на сервер
scp -r . user@server:/opt/telegram-bot/

# На сервере
cd /opt/telegram-bot
docker-compose -f docker-compose.prod.yml up -d
```

## 🔒 Безопасность

### Защищенные практики:

1. **Секреты**:
    - Токены хранятся только в переменных окружения
    - `.env` файлы в `.gitignore`
    - Использование Docker Secrets в продакшене

2. **Валидация**:
    - Проверка входных данных на всех уровнях
    - SQL injection protection через подготовленные запросы

3. **Сетевая безопасность**:
    - Изолированные Docker сети
    - Только необходимые порты открыты
    - gRPC с TLS в продакшене

### Рекомендации для продакшена:

- Использовать HTTPS для gRPC
- Настроить firewall правила
- Регулярно обновлять зависимости
- Мониторинг логов на подозрительную активность

## 🐛 Troubleshooting

### Частые проблемы:

**Бот не отвечает:**
```bash
# Проверяем логи бота
docker-compose logs telegram-bot

# Проверяем подключение к gRPC
docker-compose exec telegram-bot ping notes-service
```

**Ошибки базы данных:**
```bash
# Проверяем подключение к БД
docker-compose exec postgres psql -U postgres -d task_manager

# Проверяем таблицы
docker-compose exec postgres psql -U postgres -d task_manager -c "\dt"
```

**Проблемы с gRPC:**
```bash
# Проверяем порт gRPC
docker-compose exec notes-service netstat -tlnp | grep 50051
```

### Мониторинг:

```bash
# Статус контейнеров
docker-compose ps

# Использование ресурсов
docker stats

# Логи в реальном времени
docker-compose logs -f --tail=50
```

## 🤝 Contributing

Мы приветствуем вклад в развитие проекта!

1. Форкните репозиторий
2. Создайте feature ветку: `git checkout -b feature/amazing-feature`
3. Закоммитьте изменения: `git commit -m 'Add amazing feature'`
4. Запушьте ветку: `git push origin feature/amazing-feature`
5. Откройте Pull Request

## 📄 Лицензия

Этот проект распространяется под MIT License - смотрите файл [LICENSE](LICENSE) для деталей.

## 👨‍💻 Автор

**Ваше Имя**
- Telegram: [@glebb98](https://t.me/glebb98)

## 🙏 Благодарности

- Команде Telegram за прекрасный Bot API
- Сообществу Go за отличные библиотеки
- Docker сообществу за инструменты контейнеризации

---

**⭐ Если этот проект был полезен, поставьте звезду на GitHub!**
