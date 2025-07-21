# REST API Marketplace

Это простая релизация REST API маркетплейса на Go. В качестве базы данных используется PostgreSQL.

В репозитории есть тесты
1. Юниты, ими покрыты пакеты auth и post
2. Интеграционные, см. файл integrations
3. Примитивный bash-скрипт, для теста test.sh

Последний можно использовать для быстрого теста всех эндпойнтов. (Но не всех случаев!)

Важное замечание, если бы тут были операции с деньгами, то логичней было бы использовать что-то типа decimal, но для простоты используется float64

## Архитектура проекта

```
rest-api-marketplace/
├── cmd/
│   └── api/
│       └── main.go              # Запуск приложения, HTTP-сервер
├── internal/
│   ├── auth/                    # Регистрация и авторизация
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── storage.go
│   ├── post/                    # Управление объявлениями
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── storage.go
│   └── database/
│       └── database.go          # Настройка PostgreSQL
├── pkg/
│   ├── hash/
│   │   └── hash.go
│   └── jwt/
│       └── token.go
├── tests/                       # Тесты приложения
├── Dockerfile                   # Сборка Docker-образа
├── go.mod, go.sum
└── README.md
```

Взаимодействие между компонентами следующая:
```
HTTP-запрос → handler → service (бизнес-логика) → storage → PostgreSQL
```

Handler — принимает и валидирует HTTP-запросы

Service — выполняет бизнес-логику, сейчас её минимум

Storage — взаимодействует с базой данных

---

## Оглавление

1. [Запуск проекта](#запуск-проекта)
2. [Переменные окружения](#переменные-окружения)
3. [API эндпойнты (кратко)](#api-эндпойнты-кратко)
4. [API эндпойнты (подробно, OpenAPI-like)](#api-эндпойнты-подробно-openapi-like)
5. [Аутентификация](#аутентификация)

---

## Запуск проекта

### 1. Клонирование репозитория

```bash
git clone https://github.com/TemirB/rest-api-marketplace.git
cd rest-api-marketplace
```

### 2. Настройка переменных окружения

Скопируйте `example.env` в `.env` и при необходимости отредактируйте:

```bash
cp example.env .env
```

### 3. Запуск через Docker

Сервисы запустятся на портах:

* **Приложение**: `localhost:8080` (APP\_PORT из .env)
* **PostgreSQL**: `postgres:5432` внутри Docker-сети

---

## Переменные окружения

Файл `example.env` содержит:

```dotenv
APP_NAME=rest-api-marketplace
APP_PORT=8080

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=marketplace

JWT_SECRET=my_jwt_secret
JWT_EXPIRATION=60
```

Отредактируйте под свои нужды.

---

## API эндпойнты (кратко)

| Метод  | Путь          | Описание                         | Авторизация |
| ------ | ------------- | -------------------------------- | ----------- |
| POST   | `/register`   | Регистрация пользователя         | Нет         |
| POST   | `/login`      | Получение JWT токена             | Нет         |
| POST   | `/posts`      | Создание объявления              | Да          |
| GET    | `/posts/feed` | Получение ленты объявлений       | Нет         |
| GET    | `/posts/{id}` | Получение объявления по ID       | Нет         |
| PUT    | `/posts/{id}` | Редактирование объявления        | Да          |
| DELETE | `/posts/{id}` | Удаление объявления              | Да          |

---

## API эндпойнты (подробно, OpenAPI-like)

### 1. POST `/register`

```yaml
Request:
  Content-Type: application/json
  Body:
    login: string (3-50 chars)
    password: string (>=8 chars)

Responses:
  201 Created:
    Body: { login: string }
  400 Bad Request:
    Причины: валидация или некорректный JSON
  409 Conflict:
    Пользователь уже существует
  500 Internal Server Error
```

### 2. POST `/login`

```yaml
Request:
  Content-Type: application/json
  Body:
    login: string
    password: string

Responses:
  200 OK:
    Body: { token: string }
  400 Bad Request
  401 Unauthorized
  405 Method Not Allowed
```

### 3. POST `/posts`

```yaml
Request:
  Content-Type: application/json
  Authorization: Bearer <token>
  Body:
    title: string (1-100 chars)
    description: string (1-2000 chars)
    price: number (>0)
    image_url: string (URL)

Responses:
  201 Created:
    Body: Post object (см. ниже)
  400 Bad Request
  401 Unauthorized
  405 Method Not Allowed
```

### 4. GET `/posts/feed`

```yaml
Request:
  GET /posts/feed
  Query Params:
    min_price, max_price, sort_by, order

Responses:
  200 OK:
    Body: [ Post ]
  405 Method Not Allowed
```

### 5. GET `/posts/{id}`

```yaml
Request:
  GET /posts/{id}

Responses:
  200 OK:
    Body: Post object
  404 Not Found:
    Если объявление не существует
  405 Method Not Allowed
```

### 6. PUT `/posts/{id}`

```yaml
Request:
  Content-Type: application/json
  Authorization: Bearer <token>
  Body (любые поля для обновления):
    title?, description?, price?, image_url?

Responses:
  200 OK:
    Body: обновленный Post object
  400 Bad Request
  401 Unauthorized
  403 Forbidden:
    Если пытаются редактировать чужой пост
  404 Not Found:
    Если объявление не найдено
  405 Method Not Allowed
```

### 7. DELETE `/posts/{id}`

```yaml
Request:
  Authorization: Bearer <token>

Responses:
  204 No Content:
    Успешно удалено
  401 Unauthorized
  403 Forbidden:
    Попытка удалить чужой пост
  404 Not Found
  405 Method Not Allowed
```

**Post object:**

```json
{
  "ID": 1,
  "title": "...",
  "description": "...",
  "price": 123.45,
  "image_url": "...",
  "created_at": "2025-07-21T...Z",
  "owner": "login",
  "is_owner": true|false
}
```

---
