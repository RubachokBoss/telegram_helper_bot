# ✅ Общий кеш Redis для всех микросервисов

## 🎯 Что было сделано

Теперь **Redis используется как общий кеш для всех микросервисов**:
- ✅ `notes-service` - использует Redis для кеширования
- ✅ `telegram-bot` - использует общий Redis кеш для чтения задач
- ✅ `web-api` - использует общий Redis кеш для чтения задач

## 📋 Изменения

### 1. Docker Compose (`docker-compose.yml`)
- Добавлены переменные окружения `REDIS_HOST` и `REDIS_PORT` для `telegram-bot` и `web-api`
- Добавлены зависимости от Redis для обоих сервисов

### 2. Telegram Bot (`cmd/telegram-bot/main.go`)
- Добавлено подключение к Redis
- Создан кеш-клиент для чтения задач из общего кеша
- Обновлены обработчики `getUserTasks` и `getOwnerTasks` для использования кеша

### 3. Web API (`cmd/web-api/main.go`)
- Добавлено подключение к Redis
- Создан кеш-клиент для чтения задач из общего кеша
- Обновлены обработчики `getUserTasks` и `getOwnerTasks` для использования кеша

### 4. Новый кеш-клиент (`internal/pkg/cache/task_cache_client.go`)
- Создан общий клиент для работы с кешем задач
- Методы: `GetTask`, `GetUserTasks`, `GetOwnerTasks`
- Методы инвалидации: `InvalidateTask`, `InvalidateUserCache`, `InvalidateOwnerCache`

## 🚀 Как это работает

### Архитектура:
```
┌─────────────────┐         ┌─────────────────┐         ┌─────────────────┐
│  Telegram Bot   │         │  Notes Service  │         │    Web API      │
└────────┬────────┘         └────────┬────────┘         └────────┬────────┘
         │                            │                            │
         └────────────────────────────┼────────────────────────────┘
                                      │
                              ┌───────▼────────┐
                              │   Redis Cache  │
                              │  ОБЩИЙ КЕШ    │
                              └────────────────┘
```

### Поток данных:

1. **Создание задачи:**
   - Telegram Bot → gRPC → Notes Service → PostgreSQL + Redis (кеширование)

2. **Чтение задач:**
   - Telegram Bot/Web API → **Сначала проверяет Redis кеш** → Если есть, возвращает из кеша
   - Если нет в кеше → gRPC → Notes Service → PostgreSQL → Кеширование в Redis

## ✅ Преимущества

1. **Разделяемое состояние** - все сервисы видят один и тот же кеш
2. **Быстрые ответы** - Telegram Bot и Web API могут читать из кеша без обращения к gRPC
3. **Меньше нагрузки** - меньше запросов к PostgreSQL и gRPC
4. **Масштабирование** - можно масштабировать сервисы, кеш остается общим

## 🔧 Использование

### Запуск:
```bash
docker-compose up -d
```

Все сервисы автоматически подключатся к общему Redis кешу.

### Проверка:
```bash
# Проверить подключение к Redis
docker-compose exec redis redis-cli ping

# Посмотреть ключи в кеше
docker-compose exec redis redis-cli KEYS "*"
```

## 📝 Логирование

При использовании кеша в логах будет:
- `✅ User tasks found in shared cache for user: {userID}` - данные из кеша
- `✅ Owner tasks found in shared cache for owner: {ownerID}` - данные из кеша

В ответах Web API будет поле `source`:
- `"source": "cache"` - данные из кеша
- `"source": "database"` - данные из базы данных

## ⚠️ Важно

- Если Redis недоступен, сервисы продолжают работать через gRPC (graceful degradation)
- Кеш автоматически инвалидируется при изменениях в `notes-service`
- TTL для кеша: 10 минут для списков задач, 30 минут для отдельных задач

