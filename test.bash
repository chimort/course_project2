#!/bin/bash
set -e

# Настройки
BASE_URL="http://localhost:8080"
USERNAME="testuser"
PASSWORD="testpass"
FIRST_NAME="Test"
LAST_NAME="User"
EMAIL="test@example.com"
AGE=20
GENDER="male"

# Функция для регистрации (игнорирует ошибки дубликатов)
register_user() {
  echo "=== 1. Try to register user ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/register" \
    -H "Content-Type: application/json" \
    -d "{
      \"user\": {
        \"username\": \"$USERNAME\",
        \"first_name\": \"$FIRST_NAME\",
        \"last_name\": \"$LAST_NAME\",
        \"email\": \"$EMAIL\",
        \"password\": \"$PASSWORD\",
        \"age\": $AGE,
        \"gender\": \"$GENDER\",
        \"languages\": [{\"name\":\"en\",\"level\":1}],
        \"interests\": [{\"name\":\"books\"}]
      }
    }")
  echo "Register response: $RESPONSE"
}

# Логин
login_user() {
  echo "=== 2. Login ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/login" \
    -H "Content-Type: application/json" \
    -d "{
      \"username\": \"$USERNAME\",
      \"password\": \"$PASSWORD\"
    }")
  echo "Login response: $RESPONSE"

  ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken // .access_token')
  REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refreshToken // .refresh_token')

  echo "Access token: $ACCESS_TOKEN"
  echo "Refresh token: $REFRESH_TOKEN"
}

# Получаем профиль с access token
get_profile() {
  echo "=== 3. Get profile with access token ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/profile" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "X-Refresh-Token: $REFRESH_TOKEN" \
    -d "{}")
  echo "Profile response: $RESPONSE"
}

# Обновляем токены
refresh_tokens() {
  echo "=== 4. Refresh tokens ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/refresh-token" \
    -H "Content-Type: application/json" \
    -d "{
      \"refresh_token\": \"$REFRESH_TOKEN\"
    }")
  echo "Refresh response: $RESPONSE"

  ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.accessToken // .access_token')
  REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refreshToken // .refresh_token')

  echo "New Access token: $ACCESS_TOKEN"
  echo "New Refresh token: $REFRESH_TOKEN"
}

# Проверяем профиль с новым токеном
get_profile_new_token() {
  echo "=== 5. Get profile with new access token ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/profile" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "X-Refresh-Token: $REFRESH_TOKEN" \
    -d "{}")
  echo "Profile with new token: $RESPONSE"
}

# Основной запуск
register_user || echo "User might already exist, skipping..."
login_user
get_profile
refresh_tokens
get_profile_new_token
