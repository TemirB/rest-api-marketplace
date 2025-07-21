# REST API Marketplace

Это простая релизация REST API маркетплейса на Go. В качестве базы данных используется PostgreSQL.

В репозитории есть тесты
1. Юниты, ими покрыты пакеты auth и post
2. Интеграционные, см. файл integrations
3. Примитивный bash-скрипт, для теста test.sh

Последний можно использовать для быстрого теста всех эндпойнтов. (Но не всех случаев!)

---

## Оглавление

1. [Запуск проекта](#запуск-проекта)
2. [API эндпойнты (кратко)](#api-эндпойнты-кратко)
3. [API эндпойнты (подробно)](#api-эндпойнты-подробно)
4. [Аутентификация](#аутентификация)

---

## Запуск проекта

### 1. Клонирование репозитория

```bash
git clone https://github.com/TemirB/rest-api-marketplace.git
cd rest-api-marketplace
```
---

### 2. Настройка переменных окружения

Скопируйте `example.env` в `.env` и при необходимости отредактируйте:

```bash
cp example.env .env
```
---

### 3. Запуск 

Сервисы запустятся на портах:

* **Приложение**: `localhost:8080` (APP\_PORT из .env)
* **PostgreSQL**: `postgres:5432` внутри Docker-сети

## API эндпойнты (кратко)

| Метод  | Путь          | Описание                         | Авторизация |
| ------ | ------------- | -------------------------------- | ----------- |
| POST   | `/register`   | Регистрация пользователя         | Нет         |
| POST   | `/login`      | Получение JWT токена             | Нет         |
| POST   | `/posts`      | Создание объявления              | Да          |
| GET    | `/posts/feed` | Получение ленты объявлений       | Нет         |
| GET    | `/posts/{id}` | Получение объявления по ID       | Нет         |
| PUT    | `/posts/{id}` | Редактирование своего объявления | Да          |
| DELETE | `/posts/{id}` | Удаление своего объявления       | Да          |

---

## API эндпойнты (подробно)

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

Поле is_owner = true, заполнятеся только при запросе всех постов (и автор сопадает с логином запрашивающего пользователя), и при запросе по айди

---

## Аутентификация

1. **Регистрация**: POST `/register` → логин+пароль → создается пользователь.
2. **Логин**: POST `/login` → возвращает JWT токен.
3. **Защищенные эндпойнты** (`/posts`, `/posts/{id}` для PUT/DELETE):

   * Заголовок: `Authorization: Bearer <token>`
   * При отсутствии/невалидности токена — 401 Unauthorized.
4. **Публичные эндпойнты**: `/register`, `/login`, `/posts/feed`, `/posts/{id}` доступны без токена.

---

*Документация обновлена: добавлены операции CRUD для Posts.*
