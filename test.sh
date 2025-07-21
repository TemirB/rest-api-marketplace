#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"
USER_A="alice_$(date +%s)"
PASS_A="Secret123!"

# Заголовочные наборы
headers_json=("Content-Type: application/json")
headers_none=()

# Универсальная функция запуска теста
# $1 – описание
# $2 – метод
# $3 – URL
# $4 – тело запроса (или "")
# $5 – имя массива заголовков (например: headers_json[@] или headers_auth_a[@])
# $6 – ожидаемый статус
run_test() {
  local desc=$1 method=$2 url=$3 data=$4 hdrs_ref=$5 expect=$6

  # Разворачиваем массив-заголовков по имени
  local hdrs=()
  eval "hdrs=( \"\${${hdrs_ref}}\" )"

  echo -n "Test: $desc ... "
  local args=(-s -o /tmp/response -w "%{http_code}" -X "$method" "$url")
  for h in "${hdrs[@]}"; do
    args+=(-H "$h")
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
    echo "Body:"
    cat /tmp/response
    exit 1
  fi
}

echo "== Starting API tests for Alice ($USER_A) =="

# 1. Регистрация
run_test "Register Alice" \
  POST "$BASE_URL/register" \
  "{\"login\":\"$USER_A\",\"password\":\"$PASS_A\"}" \
  headers_json[@] 201

# 1.1. Короткий логин → 400
run_test "Register short login" \
  POST "$BASE_URL/register" \
  '{"login":"ab","password":"Secret123!"}' \
  headers_json[@] 400

# 1.2. Дубликат → 409
run_test "Register duplicate Alice" \
  POST "$BASE_URL/register" \
  "{\"login\":\"$USER_A\",\"password\":\"$PASS_A\"}" \
  headers_json[@] 409

# 2. Login Alice
run_test "Login Alice" \
  POST "$BASE_URL/login" \
  "{\"login\":\"$USER_A\",\"password\":\"$PASS_A\"}" \
  headers_json[@] 200

TOKEN_A=$(grep -Po '"token"\s*:\s*"\K[^"]+' /tmp/response)
if [[ -z $TOKEN_A ]]; then
  echo "ERROR: no token"
  exit 1
fi
headers_auth_a=("Content-Type: application/json" "Authorization: Bearer $TOKEN_A")

# 2.1. Неверный пароль → 401
run_test "Login wrong pass" \
  POST "$BASE_URL/login" \
  "{\"login\":\"$USER_A\",\"password\":\"WrongPass\"}" \
  headers_json[@] 401

# 3. Create post as Alice
run_test "Create post as Alice" \
  POST "$BASE_URL/posts" \
  '{"title":"Test Item","description":"Desc","price":123.45,"image_url":"https://example.com/img.png"}' \
  headers_auth_a[@] 201

# Извлекаем POST_ID
POST_ID=$(sed -nEn 's/.*"[iI][dD]"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p' /tmp/response)
if [[ -z "$POST_ID" ]]; then
  echo "ERROR: failed to extract post ID"
  exit 1
fi
echo "Created post ID: $POST_ID"

# 3.1. Create without token → 401
run_test "Create post no token" \
  POST "$BASE_URL/posts" \
  '{"title":"Item","description":"Desc","price":1,"image_url":"https://example.com/x.png"}' \
  headers_json[@] 401

# 3.2. Negative price → 400
run_test "Create negative price" \
  POST "$BASE_URL/posts" \
  '{"title":"Item","description":"Desc","price":-5,"image_url":"https://example.com/x.png"}' \
  headers_auth_a[@] 400

# 4. Get feed public → 200
run_test "Get feed public" \
  GET "$BASE_URL/posts/feed" \
  "" \
  headers_none[@] 200

# 5. Filter by owner (public works too) → 200
run_test "Get feed filtered by owner" \
  GET "$BASE_URL/posts/feed?owner=$USER_A" \
  "" \
  headers_none[@] 200

# 6. Get post by ID
run_test "Get post public" \
  GET "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_none[@] 200

run_test "Get post as Alice" \
  GET "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_auth_a[@] 200

# 7. Update post
run_test "Update post as Alice" \
  PUT "$BASE_URL/posts/$POST_ID" \
  '{"title":"NewTitle","description":"NewDesc","price":200,"image_url":"https://example.com/new.png"}' \
  headers_auth_a[@] 204

# 7.1 Verify update
run_test "Verify updated post" \
  GET "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_auth_a[@] 200

# 8. Delete post
run_test "Delete post no token" \
  DELETE "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_json[@] 401

run_test "Delete post as Alice" \
  DELETE "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_auth_a[@] 204

run_test "Get deleted post" \
  GET "$BASE_URL/posts/$POST_ID" \
  "" \
  headers_auth_a[@] 404

# 9. Wrong method / invalid JSON
run_test "Wrong method on register" \
  PUT "$BASE_URL/register" \
  "" \
  headers_none[@] 405

run_test "Invalid JSON on register" \
  POST "$BASE_URL/register" \
  '{"login":"bob","password":}' \
  headers_json[@] 400

# 10. Bob tries to delete Alice's post
echo "== Part 10: Bob unauthorized delete =="
USER_B="bob_$(date +%s)"
PASS_B="Password1!"
run_test "Register Bob" \
  POST "$BASE_URL/register" \
  "{\"login\":\"$USER_B\",\"password\":\"$PASS_B\"}" \
  headers_json[@] 201

run_test "Login Bob" \
  POST "$BASE_URL/login" \
  "{\"login\":\"$USER_B\",\"password\":\"$PASS_B\"}" \
  headers_json[@] 200

TOKEN_B=$(grep -Po '"token"\s*:\s*"\K[^"]+' /tmp/response)
headers_auth_b=("Content-Type: application/json" "Authorization: Bearer $TOKEN_B")

# Alice creates second post for Bob to try delete
run_test "Alice creates second post" \
  POST "$BASE_URL/posts" \
  '{"title":"Second","description":"Desc2","price":50,"image_url":"https://example.com/2.png"}' \
  headers_auth_a[@] 201

id_line=$(grep -m1 '"id"' /tmp/response)
NEW_ID="${id_line//[!0-9]/}"

run_test "Bob deletes Alice’s second post" \
  DELETE "$BASE_URL/posts/$NEW_ID" \
  "" \
  headers_auth_b[@] 401

run_test "Alice deletes second post" \
  DELETE "$BASE_URL/posts/$NEW_ID" \
  "" \
  headers_auth_a[@] 204

echo "== All tests passed! =="
