# 🚀 Redis Кэширование в Telegram Helper Bot

## 📋 Обзор архитектуры

Проект использует **паттерн декоратора** для кэширования:
```go
// В main.go
baseRepo := postgres.NewTaskRepository(db)          // Базовый репозиторий
taskRepo := redisRepo.NewTaskCacheRepository(redisClient, baseRepo) // + Redis кэш
```

## 🛠️ Основные компоненты

### 1. `taskCacheRepository` (декоратор)
**Файл:** `internal/repository/redis/task_cache_repository.go`

### 2. Вспомогательные функции
**Файл:** `internal/pkg/database/redis.go`

### 3. Ключи и TTL константы
```go
// Ключи кэша
UserTasksKey(userID)    // "user_tasks:{userID}"
OwnerTasksKey(ownerID)  // "owner_tasks:{ownerID}"
TaskKey(taskID)         // "task:{taskID}"

// Время жизни
UserTasksTTL  = 10 * time.Minute  // Списки задач пользователя
OwnerTasksTTL = 10 * time.Minute  // Списки задач владельца
TaskTTL       = 30 * time.Minute  // Индивидуальные задачи
```

---

## 🎯 Руководство по использованию

### **READ операции (самые частые)**

#### ✅ Получение списка задач пользователя
```go
func (r *taskCacheRepository) FindByUserID(userID string) ([]*domain.Task, error) {
    ctx := context.Background()
    cacheKey := database.UserTasksKey(userID)

    // ШАГ 1: Сначала проверяем кэш
    if cached, err := r.getCachedTasks(ctx, cacheKey); err == nil && cached != nil {
        log.Printf("✅ Кэш HIT для пользователя: %s", userID)
        return cached, nil
    }

    // ШАГ 2: Если нет в кэше - идем в базу
    tasks, err := r.base.FindByUserID(userID)
    if err != nil {
        return nil, err
    }

    // ШАГ 3: Кэшируем результат
    r.cacheTasks(ctx, cacheKey, tasks, database.UserTasksTTL)
    return tasks, nil
}
```

#### ✅ Получение индивидуальной задачи
```go
func (r *taskCacheRepository) FindByID(id string) (*domain.Task, error) {
    ctx := context.Background()
    cacheKey := database.TaskKey(id)

    // ШАГ 1: Проверяем кэш
    if cached, err := r.getCachedTask(ctx, cacheKey); err == nil && cached != nil {
        return cached, nil
    }

    // ШАГ 2: База данных
    task, err := r.base.FindByID(id)
    if err != nil || task == nil {
        return task, err
    }

    // ШАГ 3: Кэшируем
    r.cacheTask(ctx, cacheKey, task, database.TaskTTL)
    return task, nil
}
```

### **WRITE операции (с инвалидацией кэша)**

#### ✅ Создание задачи (умная инвалидация)
```go
func (r *taskCacheRepository) Create(task *domain.Task) error {
    // ШАГ 1: Пишем в базу
    err := r.base.Create(task)
    if err != nil {
        return err
    }

    // ШАГ 2: Пытаемся обновить кэш вместо полной инвалидации
    err = r.updateOwnerCacheWithNewTask(ctx, task)
    if err != nil {
        // Fallback: инвалидируем кэш если не удалось обновить
        r.invalidateOwnerCache(ctx, task.OwnerID)
    }
    return nil
}

func (r *taskCacheRepository) updateOwnerCacheWithNewTask(ctx context.Context, task *domain.Task) error {
    cacheKey := database.OwnerTasksKey(task.OwnerID)

    // Если кэша нет - ничего не делаем
    exists, _ := r.client.Exists(ctx, cacheKey).Result()
    if exists == 0 {
        return nil
    }

    // Получаем существующие задачи и добавляем новую в начало
    cachedTasks, err := r.getCachedTasks(ctx, cacheKey)
    if err != nil || cachedTasks == nil {
        return err
    }

    updatedTasks := append([]*domain.Task{task}, cachedTasks...)
    return r.cacheTasks(ctx, cacheKey, updatedTasks, database.OwnerTasksTTL)
}
```

#### ✅ Обновление задачи
```go
func (r *taskCacheRepository) Update(task *domain.Task) error {
    // ШАГ 1: Обновляем в базе
    err := r.base.Update(task)
    if err != nil {
        return err
    }

    ctx := context.Background()

    // ШАГ 2: Чистим кэш конкретной задачи
    taskKey := database.TaskKey(task.ID)
    r.client.Del(ctx, taskKey)

    // ШАГ 3: Чистим кэш связанных пользователей
    r.invalidateOwnerCache(ctx, task.OwnerID)
    if task.AssignedID != "" {
        r.invalidateUserCache(ctx, task.AssignedID)
    }

    return nil
}
```

#### ✅ Удаление задачи
```go
func (r *taskCacheRepository) Delete(task *domain.Task) error {
    // ШАГ 1: Удаляем из базы
    err := r.base.Delete(task)
    if err != nil {
        return err
    }

    ctx := context.Background()

    // ШАГ 2: Чистим весь связанный кэш
    taskKey := database.TaskKey(task.ID)
    r.client.Del(ctx, taskKey)
    r.invalidateOwnerCache(ctx, task.OwnerID)
    if task.AssignedID != "" {
        r.invalidateUserCache(ctx, task.AssignedID)
    }

    return nil
}
```

---

## 🛠️ Вспомогательные методы (Helper Methods)

### **Кэширование данных**
```go
// Кэшировать одну задачу
func (r *taskCacheRepository) cacheTask(ctx context.Context, key string, task *domain.Task, ttl time.Duration) {
    if task == nil {
        return
    }
    data, _ := json.Marshal(task)
    r.client.Set(ctx, key, data, ttl).Err()
}

// Кэшировать список задач
func (r *taskCacheRepository) cacheTasks(ctx context.Context, key string, tasks []*domain.Task, ttl time.Duration) {
    data, _ := json.Marshal(tasks)
    r.client.Set(ctx, key, data, ttl).Err()
}
```

### **Чтение из кэша**
```go
// Прочитать одну задачу
func (r *taskCacheRepository) getCachedTask(ctx context.Context, key string) (*domain.Task, error) {
    data, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return nil, err
    }
    var task domain.Task
    json.Unmarshal([]byte(data), &task)
    return &task, nil
}

// Прочитать список задач
func (r *taskCacheRepository) getCachedTasks(ctx context.Context, key string) ([]*domain.Task, error) {
    data, err := r.client.Get(ctx, key).Result()
    if err != nil {
        return nil, err
    }
    var tasks []*domain.Task
    json.Unmarshal([]byte(data), &tasks)
    return tasks, nil
}
```

### **Инвалидация кэша**
```go
// Очистить кэш пользователя
func (r *taskCacheRepository) invalidateUserCache(ctx context.Context, userID string) {
    key := database.UserTasksKey(userID)
    r.client.Del(ctx, key)
}

// Очистить кэш владельца
func (r *taskCacheRepository) invalidateOwnerCache(ctx context.Context, ownerID string) {
    key := database.OwnerTasksKey(ownerID)
    r.client.Del(ctx, key)
}
```

---

## 📊 Алгоритм использования (чек-лист)

### Для любой READ операции:
1. ✅ **Сначала проверь кэш** (`getCachedTask` / `getCachedTasks`)
2. ✅ **Если есть в кэше** → верни результат (CACHE HIT)
3. ✅ **Если нет в кэше** → запроси из базы данных
4. ✅ **Закэшируй результат** (`cacheTask` / `cacheTasks`)

### Для любой WRITE операции:
1. ✅ **Сделай изменения в базе данных** (через `base` репозиторий)
2. ✅ **Инвалидируй связанный кэш** (`invalidateUserCache` / `invalidateOwnerCache`)
3. ✅ **При обновлении/удалении** также удали конкретный ключ задачи

---

## 🎯 Лучшие практики

### ✅ Правильное использование TTL:
- **Индивидуальные задачи:** 30 минут (редко меняются)
- **Списки задач:** 10 минут (часто обновляются)

### ✅ Умная инвалидация кэша при изменениях:
```go
// СТАРЫЙ ПОДХОД: Полная инвалидация
func (r *taskCacheRepository) Create(task *domain.Task) error {
    err := r.base.Create(task)
    if err != nil {
        return err
    }
    r.invalidateOwnerCache(ctx, task.OwnerID) // ❌ Всегда удаляем кэш
    return nil
}

// НОВЫЙ ПОДХОД: Умное обновление кэша
func (r *taskCacheRepository) Create(task *domain.Task) error {
    err := r.base.Create(task)
    if err != nil {
        return err
    }
    // ✅ Пытаемся обновить существующий кэш
    err = r.updateOwnerCacheWithNewTask(ctx, task)
    if err != nil {
        // Fallback: инвалидируем если не удалось обновить
        r.invalidateOwnerCache(ctx, task.OwnerID)
    }
    return nil
}
```

### ✅ Graceful degradation:
- Если Redis недоступен → работай с PostgreSQL напрямую
- Логируй все операции кэширования

---

## 🎯 Преимущества умной инвалидации кэша

### 🚀 **Производительность:**
- **Быстрее:** Не нужно делать запрос к базе при следующем чтении
- **Меньше нагрузки:** База данных вызывается реже
- **Лучший UX:** Пользователи сразу видят новые задачи

### 🔒 **Надежность:**
- **Fallback стратегия:** Если обновление кэша не удалось - инвалидируем полностью
- **Консистентность:** Данные всегда актуальны
- **Graceful degradation:** Работает даже при проблемах с Redis

### 📊 **Примеры сценариев:**

```bash
# Сценарий 1: Кэш существует
# ✅ Добавляем задачу к кэшу - БД не трогаем при следующем чтении

# Сценарий 2: Кэша нет
# ✅ Ничего не делаем - кэш обновится при следующем запросе

# Сценарий 3: Ошибка обновления кэша
# ✅ Fallback: инвалидируем кэш - гарантируем консистентность
```

### 🎯 **Когда это особенно эффективно:**
- Частые создания задач одними и теми же пользователями
- Большой размер кэша (много задач у пользователя)
- Высокая нагрузка на чтение списков задач

---

## 🚨 Распространенные ошибки

### ❌ НЕ ДЕЛАТЬ:
```go
// Ошибка: кэшируем до записи в базу
func (r *taskCacheRepository) Create(task *domain.Task) error {
    r.cacheTask(ctx, key, task, ttl) // ❌ Рано!
    return r.base.Create(task)
}

// Ошибка: не инвалидируем кэш при обновлении
func (r *taskCacheRepository) Update(task *domain.Task) error {
    err := r.base.Update(task)
    // ❌ Забыли invalidateOwnerCache()
    return err
}
```

### ✅ ПРАВИЛЬНО:
```go
// Сначала пишем в базу, потом кэшируем/инвалидируем
func (r *taskCacheRepository) Create(task *domain.Task) error {
    err := r.base.Create(task)      // 1. Сначала база
    if err != nil {
        return err
    }
    r.invalidateOwnerCache(ctx, task.OwnerID) // 2. Потом инвалидация
    return nil
}
```

---

## 📈 Мониторинг и отладка

### Логи кэширования:
```
✅ Task cached: 12345
✅ User tasks found in cache for user: user123
🗑️ Invalidated owner cache: owner_tasks:user123
```

### Проверка работоспособности:
```bash
# Подключиться к Redis
docker exec -it telegram-helper-bot-redis-1 redis-cli

# Посмотреть ключи
KEYS *

# Посмотреть значение
GET "user_tasks:user123"
```

---

## 🎉 Итого

**3 простых правила:**
1. **READ:** Кэш → База → Кэшируй
2. **WRITE:** База → Инвалидируй кэш
3. **Всегда:** Логируй и обрабатывай ошибки

**Пример полной реализации:**
```go
func (r *taskCacheRepository) FindByUserID(userID string) ([]*domain.Task, error) {
    // 1. Проверяем кэш
    if cached := r.getCachedTasks(cacheKey); cached != nil {
        return cached, nil
    }

    // 2. Если нет - база
    tasks, err := r.base.FindByUserID(userID)

    // 3. Кэшируем и возвращаем
    r.cacheTasks(cacheKey, tasks, UserTasksTTL)
    return tasks, err
}
```

Теперь Redis кэширование в вашем проекте работает эффективно! 🚀
