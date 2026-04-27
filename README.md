```markdown
# 🍕 Delivery Service API

Бэкенд для сервиса доставки еды. Управление ресторанами, заказами и балансом пользователей.

![Go Version](https://img.shields.io/badge/Go-1.26.1-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-4169E1?style=flat&logo=postgresql)
![Gin](https://img.shields.io/badge/Gin-1.12.0-00ADD8?style=flat&logo=gin)
![Docker](https://img.shields.io/badge/Docker-24.0+-2496ED?style=flat&logo=docker)
![pgx](https://img.shields.io/badge/pgx-5.9.1-336791?style=flat&logo=postgresql)

---

## 📑 Содержание

- [О проекте](#о-проекте)
- [Стек технологий](#стек-технологий)
- [Архитектура](#архитектура)
- [Запуск](#запуск)
- [Эндпоинты API](#эндпоинты-api)
- [Примеры запросов](#примеры-запросов)
- [Схема БД](#схема-бд)
- [Roadmap](#roadmap)

---

## 📖 О проекте

Проект предоставляет REST API для сервиса доставки еды. Система позволяет пользователям заказывать блюда из ресторанов, пополнять баланс и отслеживать статус заказов.

### Что умеет сейчас

| Функция | Описание |
|---------|----------|
| 👤 **Пользователи** | Регистрация, просмотр профиля, пополнение баланса |
| 🏪 **Рестораны** | Добавление ресторанов и их меню |
| 📦 **Заказы** | Создание, добавление/удаление позиций, подтверждение |
| 💰 **Оплата** | Списание средств с баланса при подтверждении |
| 🐳 **Docker** | Полная контейнеризация |

---

## 🛠 Стек технологий

### Основные компоненты

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| Язык | Go | 1.26.1 |
| HTTP роутер | Gin | 1.12+ |
| Драйвер БД | pgx | v5 |
| База данных | PostgreSQL | 18 |
| Контейнеризация | Docker + Compose | latest |

### Структура слоёв

| Слой | Компонент | Задача |
|------|-----------|--------|
| 1 | Gin Router | Маршрутизация |
| 2 | Handler | Валидация, DTO |
| 3 | Service | Бизнес-логика |
| 4 | Repository | SQL запросы |
| 5 | PostgreSQL | Хранение данных |

---

## 🚀 Запуск

### Требования

```bash
Go 1.26.1         # go version
PostgreSQL 18   # psql --version
Docker           # docker --version (опционально)
Make             # make --version (опционально)
```

### Способ 1: Docker Compose (рекомендуется)

```bash
# Клонируем
git clone https://github.com/kqtagiri/Delivery
cd Delivery

# Настраиваем окружение
cp .env.example .env
# Отредактируйте CONN_STRING под свои нужды

# Запускаем всё одной командой
docker-compose up --build
```

После запуска:
- API: `http://localhost:9111`
- PostgreSQL: `localhost:5432`

### Способ 2: Локальный запуск

```bash
# Создаём базу данных
createdb delivery

# Копируем и заполняем .env
cp .env.example .env

# Запускаем приложение
go run cmd/app/main.go

# Или через Makefile
make run
```

### Проверка работоспособности

```bash
http://localhost:9111/ping
# {"Method":"GET","Status":200,"Message":"Server is open!"}
```

---

## 📡 Эндпоинты API

Базовая URL: `http://localhost:9111`

### 👤 Пользователи (`/users`)

| Метод | Эндпоинт | Действие |
|-------|----------|----------|
| POST | `/users/register` | Создать аккаунт |
| PATCH | `/users/replenish/:email` | Пополнить баланс |
| GET | `/users/:email` | Получить данные |
| GET | `/users` | Список всех |

### 🏪 Рестораны (`/restaurants`)

| Метод | Эндпоинт | Действие |
|-------|----------|----------|
| POST | `/restaurants/create` | Создать ресторан |
| GET | `/restaurants` | Все рестораны |
| GET | `/restaurants/:title` | Меню ресторана |
| POST | `/restaurants/:title` | Добавить блюда |
| DELETE | `/restaurants/:title` | Удалить блюда |

### 📦 Заказы (`/orders`)

| Метод | Эндпоинт | Действие |
|-------|----------|----------|
| POST | `/orders/create` | Создать заказ |
| POST | `/orders/add/:number` | Добавить позиции |
| DELETE | `/orders/deleteitems/:number` | Удалить позиции |
| DELETE | `/orders/delete/:number` | Отменить заказ |
| GET | `/orders/:number` | Информация о заказе |
| GET | `/orders/details/:number` | Детали заказа |
| POST | `/orders/confirm/:number` | Подтвердить оплату |

---

## 📝 Примеры запросов

### 1. Регистрация пользователя

```bash
POST http://localhost:9111/users/register \
  "Content-Type: application/json" \
  '{
    "name": "Алексей Иванов",
    "email": "alex@example.com",
    "address": "ул. Ленина, д. 10, кв. 5"
  }'
```

**Ответ:**
```json
{
  "name": "Алексей Иванов",
  "email": "alex@example.com",
  "address": "ул. Ленина, д. 10, кв. 5",
  "balance": 0
}
```

### 2. Пополнение баланса

```bash
PATCH http://localhost:9111/users/replenish/alex@example.com \
  "Content-Type: application/json" \
  '{"balance": 1500.00}'
```

### 3. Создание ресторана

```bash
POST http://localhost:9111/restaurants/create \
  "Content-Type: application/json" \
  '{
    "title": "Суши Шоп",
    "description": "Лучшие роллы в городе",
    "address": "ул. Пушкина, 15",
    "rating": 4.8
  }'
```

### 4. Добавление блюд в меню

```bash
POST http://localhost:9111/restaurants/Суши%20Шоп \
  "Content-Type: application/json" \
  '[
    {
      "title": "Филадельфия",
      "restaurant_title": "Суши Шоп",
      "composition": "Лосось, сливочный сыр, огурец",
      "time": 15,
      "cost": 450.00
    },
    {
      "title": "Калифорния",
      "restaurant_title": "Суши Шоп",
      "composition": "Краб, авокадо, огурец",
      "time": 12,
      "cost": 390.00
    }
  ]'
```

### 5. Создание заказа

```bash
POST http://localhost:9111/orders/create \
  "Content-Type: application/json" \
  '{
    "email": "alex@example.com",
    "address": "ул. Ленина, д. 10, кв. 5"
  }'
```

**Ответ:** `Order created! Number of your order - 1`

### 6. Добавление блюд в заказ

```bash
POST http://localhost:9111/orders/add/1 \
  "Content-Type: application/json" \
  '[
    {"title": "Филадельфия", "restaurant_title": "Суши Шоп"},
    {"title": "Калифорния", "restaurant_title": "Суши Шоп"}
  ]'
```

### 7. Просмотр заказа

```bash
GET http://localhost:9111/orders/1
```

### 8. Подтверждение и оплата

```bash
POST "http://localhost:9111/orders/confirm/1?email=alex@example.com"
```

---

## 🗄 Схема базы данных

```
┌─────────────┐     ┌─────────────┐
│   users     │     │ restaurant  │
├─────────────┤     ├─────────────┤
│ id (PK)     │     │ id (PK)     │
│ name        │     │ title       │
│ email (U)   │     │ description │
│ address     │     │ address     │
│ balance     │     │ rating      │
└─────────────┘     └──────┬──────┘
                           │
                           ▼
                    ┌─────────────┐
                    │    menu     │
                    ├─────────────┤
                    │ id (PK)     │
                    │ title       │
                    │ rest_title  │
                    │ composition │
                    │ time        │
                    │ cost        │
                    └─────────────┘

┌─────────────┐     ┌─────────────────┐
│   orders    │     │  orders_items   │
├─────────────┤     ├─────────────────┤
│ id (PK)     │◄────│ id (PK)         │
│ number (U)  │     │ number (FK)     │
│ email       │     │ title           │
│ address     │     │ rest_title      │
│ status      │     │ time            │
│ time        │     │ cost            │
│ cost        │     └─────────────────┘
└─────────────┘
```

### Типы статусов заказов

| Статус | Описание |
|--------|----------|
| `Created` | Заказ создан, можно редактировать |
| `Confirmed` | Заказ оплачен, изменения запрещены |
| `Canceled` | Заказ отменён |

---

## 📁 Структура проекта

```
Delivery/
├── cmd/
│   └── app/
│       └── main.go                 # Точка входа
│
├── internal/
│   ├── domain/                     # Бизнес-сущности
│   │   ├── user.go
│   │   ├── order.go
│   │   ├── restaurant.go
│   │   └── item.go
│   │
│   ├── handler/                    # HTTP handlers
│   │   ├── user_handler.go
│   │   ├── order_handler.go
│   │   └── restaurant_handler.go
│   │
│   ├── service/                    # Бизнес-логика
│   │   ├── user_service.go
│   │   ├── order_service.go
│   │   └── restaurant_service.go
│   │
│   ├── repository/                 # Работа с БД
│   │   ├── user_repository.go
│   │   ├── order_repository.go
│   │   └── restaurant_repository.go
│   │
│   └── database/                   # Подключение к БД
│       └── db.go
│
├── .env.example                    # Пример переменных
├── docker-compose.yml              # Docker Compose
├── Dockerfile                      # Docker образ
├── Makefile                        # Утилиты сборки
├── go.mod
└── go.sum
```

---

## 🗺 Roadmap

### Ближайшие планы

- [ ] **JWT аутентификация** — безопасные эндпоинты
- [ ] **Redis кэш** — для меню ресторанов
- [ ] **Swagger документация** — интерактивное API

### В перспективе

- [ ] **WebSocket** — статус заказа в реальном времени
- [ ] **Курьеры** — назначение ближайшего (Redis Geo)
- [ ] **Email уведомления** — Worker pool
- [ ] **Rate limiting** — защита от DDoS
- [ ] **Prometheus + Grafana** — мониторинг
- [ ] **Unit тесты** — покрытие 70%+

---

## 👨‍💻 Автор

**Никита** ([@kqtagiri](https://github.com/kqtagiri))

---

*Последнее обновление: Апрель 2026*
