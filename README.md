# REST API Marketplace

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
# затем откройте .env и задайте свои значения
```

### 3. Запуск через Docker Compose

Убедитесь, что Docker и Docker Compose установлены, затем выполните:

```bash
docker-compose up --build
```

Сервисы запустятся на портах:

* **Приложение**: `localhost:8080` (APP\_PORT из .env)
* **PostgreSQL**: в Docker-сети на `postgres:5432` (не проброшен наружу)

### 4. Остановка

```bash
docker-compose down
```

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

| Метод | Путь          | Описание                   | Авторизация |
| ----- | ------------- | -------------------------- | ----------- |
| POST  | `/register`   | Регистрация пользователя   | Нет         |
| POST  | `/login`      | Получение JWT токена       | Нет         |
| POST  | `/posts`      | Создание объявления        | Да          |
| GET   | `/posts/feed` | Получение ленты объявлений | Нет         |

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
    Content-Type: application/json
    Body: { login: string }
  400 Bad Request:
    Причины: валидация login/password или некорректный JSON
  409 Conflict:
    Пользователь уже существует
  500 Internal Server Error:
    Внутренняя ошибка сервера
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
    Content-Type: application/json
    Body: { token: string }
  400 Bad Request:
    Некорректный JSON
  401 Unauthorized:
    Неправильный логин или пароль
  405 Method Not Allowed:
    Неверный HTTP метод
```

### 3. POST `/posts`

```yaml
Request:
  Content-Type: application/json
  Authorization: Bearer <JWT-token>
  Body:
    title: string (1-100 chars)
    description: string (1-2000 chars)
    price: number (>0)
    image_url: string (URL)

Responses:
  201 Created:
    Content-Type: application/json
    Body: Post object:
      ID, title, description, price, image_url,
      created_at (ISO 8601), owner (login), is_owner: bool
  400 Bad Request:
    Ошибка валидации полей или создание не удалось
  401 Unauthorized:
    Нет или неверный токен
  405 Method Not Allowed:
    Неверный HTTP метод
```

### 4. GET `/posts/feed`

```yaml
Request:
  GET /posts/feed
  Query Params (optional):
    min_price: number
    max_price: number
    sort_by: price | created_at
    order: asc | desc

Responses:
  200 OK:
    Content-Type: application/json
    Body: Array of Post objects (см. выше)
  405 Method Not Allowed:
    Неверный HTTP метод
  500 Internal Server Error:
    Ошибка на сервере
```

---

## Аутентификация

1. **Регистрация**: `POST /register` → логин+пароль → создается пользователь.
2. **Логин**: `POST /login` → возвращает JWT токен.
3. **Защищенные эндпойнты** (`/posts`):

   * Добавьте заголовок:

     ```http
     Authorization: Bearer <token>
     ```
   * Если токен отсутствует или невалиден → 401 Unauthorized.
4. **Публичные эндпойнты**: `/register`, `/login`, `/posts/feed` доступны без токена.

---

*Документация сгенерирована автоматически.*
