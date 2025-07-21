#!/usr/bin/env bash
# test.sh — Автоматизированный запуск набора тестовых запросов к REST API

set -euo pipefail

BASE_URL="http://localhost:8080"
CT="Content-Type: application/json"

# Функция для выполнения теста
# $1 — описание
# $2 — метод
# $3 — URL
# $4 — данные (или пустая строка)
# $5 — массив заголовков
# $6 — ожидаемый код
run_test() {
  local desc=$1 method=$2 url=$3 data=$4 headers=("${!5}") expect=$6

  echo -n "Test: $desc ... "

  local args=(-s -o /tmp/response -w "%{http_code}" -X "$method" "$url")
  for hdr in "${headers[@]}"; do
    args+=(-H "$hdr")
  done
  if [[ -n $data ]]; then
    args+=(-d "$data")
  fi

  local code
  code=$(curl "${args[@]}")
  if [[ $code -eq $expect ]]; then
    echo "OK ($code)"
  else
    echo "FAIL (got $code, expected $expect)"
    echo "Response body:"
    cat /tmp/response
    exit 1
  fi
}

echo "== Starting API tests =="

# Подготовим заголовки
headers_json=("Content-Type: application/json")
headers_none=()

# 1. Регистрация
run_test "Register new user alice" \
  POST "$BASE_URL/register" \
  '{"login":"alice","password":"Secret123"}' \
  headers_json[@] 201

# 1.1. Короткий логин => 400
run_test "Register with short login" \
  POST "$BASE_URL/register" \
  '{"login":"ab","password":"Secret123"}' \
  headers_json[@] 400

# 1.2. Дубликат пользователя => 409 (или 400 в зависимости от реализации)
run_test "Register duplicate alice" \
  POST "$BASE_URL/register" \
  '{"login":"alice","password":"Secret123"}' \
  headers_json[@] 409

# 2. Логин и получение токена
run_test "Login correct credentials" \
  POST "$BASE_URL/login" \
  '{"login":"alice","password":"Secret123"}' \
  headers_json[@] 200

# Извлекаем токен из /tmp/response
TOKEN=$(sed -nE 's/.*"token"\s*:\s*"([^"]+)".*/\1/p' /tmp/response)
if [[ -z "$TOKEN" ]]; then
  echo "ERROR: failed to extract JWT token"
  exit 1
fi
echo "Token: $TOKEN"

# 2.1. Неверный пароль => 401
run_test "Login wrong password" \
  POST "$BASE_URL/login" \
  '{"login":"alice","password":"WrongPass"}' \
  headers_json[@] 401

# 3. Создание поста
headers_auth_json=("Content-Type: application/json" "Authorization: Bearer $TOKEN")
run_test "Create post with token" \
  POST "$BASE_URL/posts" \
  '{"title":"Test Item","description":"Desc","price":123.45,"image_url":"https://example.com/img"}' \
  headers_auth_json[@] 201

# 3.1. Без токена => 401
run_test "Create post without token" \
  POST "$BASE_URL/posts" \
  '{"title":"Test Item","description":"Desc","price":123.45,"image_url":"https://example.com/img"}' \
  headers_json[@] 401

# 3.2. Отрицательная цена => 400
run_test "Create post negative price" \
  POST "$BASE_URL/posts" \
  '{"title":"Test Item","description":"Desc","price":-5,"image_url":"https://example.com/img"}' \
  headers_auth_json[@] 400

# 4. Получение ленты
run_test "Get feed (unprotected)" \
  GET "$BASE_URL/posts/feed" \
  "" \
  headers_none[@] 200

# 5. Получение списка с фильтром по owner (требует авторизации)
run_test "Get posts filtered by owner" \
  GET "$BASE_URL/posts?owner=alice" \
  "" \
  headers_auth_json[@] 200

# 6. Негативные сценарии
# 6.1. Wrong method
run_test "Wrong method on register" \
  PUT "$BASE_URL/register" \
  "" \
  headers_none[@] 405

# 6.2. Invalid JSON
run_test "Invalid JSON on register" \
  POST "$BASE_URL/register" \
  '{"login":"bob","password":}' \
  headers_json[@] 400

echo "== All tests passed! =="
