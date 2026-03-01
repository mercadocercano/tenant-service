#!/bin/bash

# Script de prueba para el endpoint POST /api/v1/tenant/config
# Verifica el comportamiento de escritura (upsert)

set -e

BASE_URL="http://localhost:8120"
TENANT_ID="00000000-0000-0000-0000-000000000001"

echo "🧪 Test de Write Path - Tenant Config"
echo "======================================"
echo ""

# Test 1: INSERT - Crear nueva configuración
echo "📝 Test 1: INSERT - Crear nueva configuración"
echo "----------------------------------------------"
RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tenant/config" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "test.write.policy",
    "value": "TEST_VALUE_1"
  }')

echo "Response: ${RESPONSE}"
echo ""

# Verificar que se creó
echo "✅ Verificando que se creó..."
GET_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/tenant/config/test.write.policy" \
  -H "X-Tenant-ID: ${TENANT_ID}")

echo "GET Response: ${GET_RESPONSE}"
echo ""

# Test 2: UPDATE - Actualizar configuración existente
echo "📝 Test 2: UPDATE - Actualizar configuración existente"
echo "-------------------------------------------------------"
RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tenant/config" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "test.write.policy",
    "value": "TEST_VALUE_2_UPDATED"
  }')

echo "Response: ${RESPONSE}"
echo ""

# Verificar que se actualizó
echo "✅ Verificando que se actualizó..."
GET_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/tenant/config/test.write.policy" \
  -H "X-Tenant-ID: ${TENANT_ID}")

echo "GET Response: ${GET_RESPONSE}"
echo ""

# Test 3: Validación - Key vacío
echo "📝 Test 3: Validación - Key vacío (debe fallar)"
echo "------------------------------------------------"
RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tenant/config" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "",
    "value": "TEST_VALUE"
  }')

echo "Response: ${RESPONSE}"
echo ""

# Test 4: Validación - Value vacío
echo "📝 Test 4: Validación - Value vacío (debe fallar)"
echo "--------------------------------------------------"
RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tenant/config" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "test.key",
    "value": ""
  }')

echo "Response: ${RESPONSE}"
echo ""

# Test 5: Validación - Sin X-Tenant-ID
echo "📝 Test 5: Validación - Sin X-Tenant-ID (debe fallar)"
echo "------------------------------------------------------"
RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tenant/config" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "test.key",
    "value": "TEST_VALUE"
  }')

echo "Response: ${RESPONSE}"
echo ""

echo "✅ Tests completados!"
echo ""
echo "💡 Para limpiar los datos de prueba, ejecuta:"
echo "   psql -h localhost -p 5435 -U postgres -d tenant_db -c \"DELETE FROM tenant_config WHERE config_key = 'test.write.policy';\""
