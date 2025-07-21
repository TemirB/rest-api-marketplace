#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"
# Уникальный логин, чтобы тесты не конфликтовали с прошлым запуском
USER="alice_$(date +%s)"

# Заголовочные наборы
headers_json=("Content-Type: application/json")
headers_none=()
# headers_auth_json будем формировать после логина

# Функция выполнения теста
# $1 — описание
# $2 — метод
# $3 — URL
# $4 — данные (строка или "")
# $5 — имя переменной-с-именем-массива заголовков (напр. headers_json)
# $6 — ожидаемый HTTP-код
run_test() {
  local desc=$1 method=$2 url=$3 data=$4 hdrs_name=$5 expect=$6

  # Достаём массив заголовков по имени
  local headers=()
  eval "headers=(\"\${${hdrs_name}[@]}\")"

  echo -n "Test: $desc ... "

  # Формируем curl
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

echo "== Starting API tests for user $USER =="

# 1. Регистрация
run_test "Register new user $USER" \
  POST "$BASE_URL/register" \
  "{\"login\":\"$USER\",\"password\":\"Secret123\"}" \
  headers_json 201

# 1.1. Слишком короткий логин => 400
run_test "Register short login" \
  POST "$BASE_URL/register" \
  '{"login":"ab","password":"Secret123"}' \
  headers_json 400

# 1.2. Дубликат => 409
run_test "Register duplicate $USER" \
  POST "$BASE_URL/register" \
  "{\"login\":\"$USER\",\"password\":\"Secret123\"}" \
  headers_json 409

# 2. Логин
run_test "Login correct credentials" \
  POST "$BASE_URL/login" \
  "{\"login\":\"$USER\",\"password\":\"Secret123\"}" \
  headers_json 200

# Извлекаем токен
TOKEN=$(sed -nE 's/.*"token"\s*:\s*"([^"]+)".*/\1/p' /tmp/response)
if [[ -z "$TOKEN" ]]; then
  echo "ERROR: failed to extract JWT token"
  exit 1
fi
echo "Token: $TOKEN"

# Заголовки с авторизацией
headers_auth_json=("Content-Type: application/json" "Authorization: Bearer $TOKEN")

# 2.1. Неверный пароль => 401
run_test "Login wrong password" \
  POST "$BASE_URL/login" \
  "{\"login\":\"$USER\",\"password\":\"WrongPass\"}" \
  headers_json 401

# 3. Создание поста
run_test "Create post with token" \
  POST "$BASE_URL/posts" \
  '{"title":"Test Item","description":"Desc","price":123.45,"image_url":"https://example.com/img"}' \
  headers_auth_json 201

# Извлекаем ID созданного поста
POST_ID=$(sed -nEn 's/.*"[iI][dD]"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p' /tmp/response)
if [[ -z "$POST_ID" ]]; then
  echo "ERROR: failed to extract post ID"
  exit 1
fi
echo "Created post ID: $POST_ID"

# 3.1. Без токена => 401
run_test "Create post without token" \
  POST "$BASE_URL/posts" \
  '{"title":"Test Item","description":"Desc","price":123.45,"image_url":"https://example.com/img"}' \
  headers_json 401

# 3.2. Отрицательная цена => 400
run_test "Create post negative price" \
  POST "$BASE_URL/posts" \
  '{"title":"Test Item","description":"Desc","price":-5,"image_url":"https://example.com/img"}' \
  headers_auth_json 400

# 4. Получение ленты (публичный эндпоинт)
run_test "Get feed (unprotected)" \
  GET "$BASE_URL/posts/feed" \
  "" \
  headers_none 200

# 5. Фильтр по owner (требует авторизации)
run_test "Get posts filtered by owner" \
  GET "$BASE_URL/posts/feed?owner=$USER" \
  "" \
  headers_auth_json 200

# 6. Получение одного поста по ID
run_test "Get post without token" \
  GET "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_none 401

run_test "Get post with token" \
  GET "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_auth_json 200

# 7. Обновление поста
run_test "Update post without token" \
  PUT "$BASE_URL/posts/$POST_ID" \
  '{"title":"NewTitle","description":"NewDesc","price":200,"image_url":"https://example.com/new.png"}' \
  headers_json 401

run_test "Update post with token" \
  PUT "$BASE_URL/posts/$POST_ID" \
  '{"title":"NewTitle","description":"NewDesc","price":200,"image_url":"https://example.com/new.png"}' \
  headers_auth_json 204

# Проверим, что изменения сохранились
run_test "Get updated post" \
  GET "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_auth_json 200

# 8. Удаление поста
run_test "Delete post without token" \
  DELETE "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_json 401

run_test "Delete post with token" \
  DELETE "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_auth_json 204

# После удаления — должен быть 404
run_test "Get deleted post" \
  GET "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_auth_json 404

# 9. Негативные сценарии на регистр и логин
run_test "Wrong method on register" \
  PUT "$BASE_URL/register" \
  "" \
  headers_none 405

run_test "Invalid JSON on register" \
  POST "$BASE_URL/register" \
  '{"login":"bob","password":}' \
  headers_json 400

echo "== All tests passed! =="

# 10 fixes

# ФИД без авторизации
echo "Public feed:"
curl -s "$BASE_URL/posts/feed" | jq .

# ФИД с авторизацией
echo "Alice's feed:"
curl -s -H "Authorization: Bearer $TOKEN_ALICE" "$BASE_URL/posts/feed" | jq .

# GET поста без токена
echo "Get post #$ID1 public:"
curl -s "$BASE_URL/posts/$ID1" | jq .

# GET поста с токеном Alice
echo "Get post #$ID1 as Alice:"
curl -s -H "Authorization: Bearer $TOKEN_ALICE" "$BASE_URL/posts/$ID1" | jq .

# Некорректный DELETE поста чужим пользователем (Bob)
echo "Delete #$ID2 as Bob (should fail):"
curl -s -X DELETE -H "Authorization: Bearer $TOKEN_BOB" \
     "$BASE_URL/posts/$ID2" -w "\nStatus: %{http_code}\n"

# Успешный DELETE поста Alice
echo "Delete #$ID2 as Alice:"
curl -s -X DELETE -H "Authorization: Bearer $TOKEN_ALICE" \
     "$BASE_URL/posts/$ID2" -w "\nStatus: %{http_code}\n"

# Проверка, что пост удалён
echo "Get deleted #$ID2 (should 404):"
curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/posts/$ID2"
echo
