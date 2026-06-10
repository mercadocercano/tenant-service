#!/usr/bin/env bash
# E2E smoke tests contra el tenant-service corriendo con lab-postgres.
# Requiere: docker compose up -d (ver docker-compose.yml) + JWT_TOKEN válido.
#
# Para obtener JWT_TOKEN:
#   TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
#     -H 'Content-Type: application/json' \
#     -d '{"email":"admin@test.com","password":"password"}' | jq -r .token)
#   JWT_TOKEN=$TOKEN bash scripts/e2e.sh
#
# Nota: SQLite no es viable porque el servicio usa JSONB y ON CONFLICT de PostgreSQL.
# El script usa el lab-postgres ya levantado vía lab-network.

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8125}"
TENANT_ID="${TENANT_ID:-550e8400-e29b-41d4-a716-446655440000}"
JWT_TOKEN="${JWT_TOKEN:-}"
PASS=0
FAIL=0

check() {
  local name="$1"; local expected_status="$2"; local actual_status="$3"
  if [ "$actual_status" -eq "$expected_status" ]; then
    echo "  PASS: $name"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $name — esperado $expected_status, obtenido $actual_status"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== E2E Tenant Service: $BASE_URL ==="

# 1. Health check (sin auth)
status=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
check "GET /health" 200 "$status"

# 2. Sin token → 401
status=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-Tenant-ID: $TENANT_ID" \
  "$BASE_URL/api/v1/tenant/settings")
check "GET /settings sin token → 401" 401 "$status"

if [ -z "$JWT_TOKEN" ]; then
  echo ""
  echo "JWT_TOKEN no definido — saltando tests autenticados."
  echo "Pasar JWT_TOKEN=<token> para el suite completo."
else
  AUTH=(-H "Authorization: Bearer $JWT_TOKEN" -H "X-Tenant-ID: $TENANT_ID")

  # 3. Bootstrap (idempotente)
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${AUTH[@]}" "$BASE_URL/api/v1/tenant/bootstrap")
  check "POST /bootstrap (idempotente)" 200 "$status" || true

  # 4. Set config
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${AUTH[@]}" \
    -H "Content-Type: application/json" \
    -d '{"key":"catalog.stock_policy","value":"IGNORE"}' \
    "$BASE_URL/api/v1/tenant/config")
  check "POST /config" 200 "$status"

  # 5. Get config
  status=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH[@]}" \
    "$BASE_URL/api/v1/tenant/config/catalog.stock_policy")
  check "GET /config/catalog.stock_policy" 200 "$status"

  # 6. Get settings
  status=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH[@]}" "$BASE_URL/api/v1/tenant/settings")
  check "GET /settings" 200 "$status"

  # 7. Get points of sale
  status=$(curl -s -o /dev/null -w "%{http_code}" "${AUTH[@]}" "$BASE_URL/api/v1/tenant/points-of-sale")
  check "GET /points-of-sale" 200 "$status"

  # 8. Create point of sale
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${AUTH[@]}" \
    -H "Content-Type: application/json" \
    -d '{"code":99,"description":"E2E Test POS","is_fiscal_enabled":true,"default_invoice_type":"B"}' \
    "$BASE_URL/api/v1/tenant/points-of-sale")
  check "POST /points-of-sale" 201 "$status"
fi

echo ""
echo "=== Resultado: PASS=$PASS FAIL=$FAIL ==="
[ "$FAIL" -eq 0 ]
