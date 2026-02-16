#!/bin/bash
set -e

BASE_URL="http://localhost:8080"
USERNAME="testuser"
PASSWORD="testpass"
FIRST_NAME="Test"
LAST_NAME="User"
EMAIL="test@example.com"
AGE=20
GENDER="male"

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

join_queue() {
  echo "=== 3. Add user to matching queue ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/matching/join" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{\"username\":\"$USERNAME\",\"mode\":1}")
  echo "JoinQueue response: $RESPONSE"
}

leave_queue() {
  echo "=== 4. Remove user from matching queue ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/matching/leave" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{\"username\":\"$USERNAME\"}")
  echo "LeaveQueue response: $RESPONSE"
}

list_queue() {
  echo "=== 5. List users in matching queue ==="
  RESPONSE=$(curl -s -X GET "$BASE_URL/v1/matching/list" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
  echo "ListQueue response: $RESPONSE"
}

get_profile() {
  echo "=== 6. Get profile with access token ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/profile" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "X-Refresh-Token: $REFRESH_TOKEN" \
    -d "{}")
  echo "Profile response: $RESPONSE"
}

refresh_tokens() {
  echo "=== 7. Refresh tokens ==="
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

update_profile() {
  echo "=== 8. Update profile ==="
  RESPONSE=$(curl -s -X PATCH "$BASE_URL/v1/profile" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
      \"user\": {
        \"first_name\": \"UpdatedFirst\",
        \"last_name\": \"UpdatedLast\",
        \"age\": 25,
        \"languages\": [{\"name\":\"en\",\"level\":1}],
        \"interests\": [{\"name\":\"books\"},{\"name\":\"movies\"}]
      }
    }")
  echo "Update profile response: $RESPONSE"
}

update_profile_partial() {
  echo "=== 9. Partial update profile ==="
  RESPONSE=$(curl -s -X PATCH "$BASE_URL/v1/profile" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -d "{
      \"user\": {
        \"first_name\": \"PartiallyUpdated\",
        \"age\": 30,
        \"interests\": [{\"name\":\"music\"}]
      }
    }")
  echo "Partial update response: $RESPONSE"
}

get_profile_after_update() {
  echo "=== 10. Get profile after update ==="
  RESPONSE=$(curl -s -X POST "$BASE_URL/v1/profile" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "X-Refresh-Token: $REFRESH_TOKEN" \
    -d "{}")
  echo "Profile after update: $RESPONSE"
}

# Основной запуск
register_user || echo "User might already exist, skipping..."
login_user
join_queue
list_queue
leave_queue
get_profile
refresh_tokens
update_profile
get_profile_after_update
update_profile_partial
get_profile_after_update
