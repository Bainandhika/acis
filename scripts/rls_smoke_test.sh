#!/usr/bin/env bash
# ==============================================================================
# ACIS - Row-Level Security (RLS) & Auth Negative Smoke Test Suite
# ==============================================================================
# Usage:
#   API_BASE_URL="http://localhost:8080" \
#   USER_A_TOKEN="eyJhbGciOi..." \
#   USER_B_FAMILY_ID="00000000-0000-0000-0000-000000000000" \
#   BOT_SECRET="test_bot_secret" \
#   ./scripts/rls_smoke_test.sh
# ==============================================================================

set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
USER_A_TOKEN="${USER_A_TOKEN:-}"
USER_B_FAMILY_ID="${USER_B_FAMILY_ID:-}"
BOT_SECRET="${BOT_SECRET:-}"

echo "============================================================"
echo " Starting ACIS RLS & Auth Smoke Tests"
echo " Target API: ${API_BASE_URL}"
echo "============================================================"

FAILED_TESTS=0

run_test() {
  local test_name="$1"
  local expected_status="$2"
  local http_method="$3"
  local endpoint="$4"
  local token="$5"
  local extra_headers="${6:-}"
  local body="${7:-}"

  echo -n "Running: ${test_name} ... "

  local curl_cmd=(curl -s -o /dev/null -w "%{http_code}" -X "${http_method}")
  
  if [ -n "${token}" ]; then
    curl_cmd+=(-H "Authorization: Bearer ${token}")
  fi

  if [ -n "${extra_headers}" ]; then
    curl_cmd+=(-H "${extra_headers}")
  fi

  if [ -n "${body}" ]; then
    curl_cmd+=(-H "Content-Type: application/json" -d "${body}")
  fi

  curl_cmd+=("${API_BASE_URL}${endpoint}")

  local status_code
  status_code=$("${curl_cmd[@]}")

  if [ "${status_code}" -eq "${expected_status}" ]; then
    echo "✅ PASS (HTTP ${status_code})"
  else
    echo "❌ FAIL (Expected HTTP ${expected_status}, got ${status_code})"
    FAILED_TESTS=$((FAILED_TESTS + 1))
  fi
}

# ------------------------------------------------------------------------------
# Test 1: No token on authenticated endpoint -> 401 Unauthorized
# ------------------------------------------------------------------------------
run_test "1. No token -> 401 on /api/v1/auth/me" 401 "GET" "/api/v1/auth/me" ""

# ------------------------------------------------------------------------------
# Test 2: Valid User A token -> 200 OK on /api/v1/family/me
# ------------------------------------------------------------------------------
if [ -n "${USER_A_TOKEN}" ]; then
  run_test "2. User A token -> 200 on own families (/api/v1/family/me)" 200 "GET" "/api/v1/family/me" "${USER_A_TOKEN}"
else
  echo "⚠️ Skipping Test 2 (USER_A_TOKEN not set)"
fi

# ------------------------------------------------------------------------------
# Test 3: User A token + User B family_id header -> 403 Forbidden / 404 Not Found
# RLS / FamilyContextMiddleware blocks cross-tenant access to wallets/transactions
# ------------------------------------------------------------------------------
if [ -n "${USER_A_TOKEN}" ] && [ -n "${USER_B_FAMILY_ID}" ]; then
  echo -n "Running: 3. User A token + User B family_id -> isolated (403/404 on /api/v1/wallets) ... "
  status_code=$(curl -s -o /dev/null -w "%{http_code}" -X GET \
    -H "Authorization: Bearer ${USER_A_TOKEN}" \
    -H "X-Family-ID: ${USER_B_FAMILY_ID}" \
    "${API_BASE_URL}/api/v1/wallets")
  
  if [ "${status_code}" -eq 403 ] || [ "${status_code}" -eq 404 ]; then
    echo "✅ PASS (HTTP ${status_code})"
  else
    echo "❌ FAIL (Expected HTTP 403 or 404, got ${status_code})"
    FAILED_TESTS=$((FAILED_TESTS + 1))
  fi
else
  echo "⚠️ Skipping Test 3 (USER_A_TOKEN and/or USER_B_FAMILY_ID not set)"
fi

# ------------------------------------------------------------------------------
# Test 4: /internal/* without BOT_INTERNAL_SECRET -> 401 Unauthorized
# ------------------------------------------------------------------------------
run_test "4. /api/v1/internal/telegram/link without BOT_INTERNAL_SECRET -> 401" 401 "POST" "/api/v1/internal/telegram/link" "" "" '{"code":"TEST01","chat_id":12345}'

# Test 4b: /api/v1/bot/family without BOT_INTERNAL_SECRET -> 401 Unauthorized
run_test "4b. /api/v1/bot/family without BOT_INTERNAL_SECRET -> 401" 401 "GET" "/api/v1/bot/family?chat_id=12345" ""

echo "============================================================"
if [ "${FAILED_TESTS}" -eq 0 ]; then
  echo "🎉 All RLS & Security smoke test cases passed successfully!"
  exit 0
else
  echo "💥 ${FAILED_TESTS} test case(s) failed."
  exit 1
fi
