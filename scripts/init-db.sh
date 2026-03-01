#!/bin/bash

# Script de inicialización de base de datos para tenant-service
# Ejecuta las migraciones y opcionalmente inserta datos de prueba

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuración de BD (usa variables de entorno o defaults)
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5435}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-tenant_db}

echo -e "${GREEN}🚀 Inicializando base de datos del Tenant Service${NC}"
echo "Host: $DB_HOST:$DB_PORT"
echo "Database: $DB_NAME"
echo ""

# Verificar que PostgreSQL esté disponible
echo -e "${YELLOW}⏳ Verificando conexión a PostgreSQL...${NC}"
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  echo "PostgreSQL no está disponible - esperando..."
  sleep 2
done

echo -e "${GREEN}✅ PostgreSQL está disponible${NC}"
echo ""

# Ejecutar migraciones
echo -e "${YELLOW}📦 Ejecutando migraciones...${NC}"

# Migración 1: Crear tabla
echo "Ejecutando: 001_create_tenant_config_table.sql"
PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f migrations/001_create_tenant_config_table.sql

echo -e "${GREEN}✅ Migraciones ejecutadas exitosamente${NC}"
echo ""

# Preguntar si insertar datos de prueba
read -p "¿Deseas insertar datos de prueba? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}📝 Insertando datos de prueba...${NC}"
    
    # Insertar configuraciones de ejemplo
    PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<EOF
-- Datos de prueba para tenant-service
-- Nota: Este tenant debe existir en iam-service

-- Tenant de prueba 1
INSERT INTO tenant_config (tenant_id, config_key, config_value)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'catalog.stock_policy', 'IGNORE_STOCK'),
    ('00000000-0000-0000-0000-000000000001', 'catalog.auto_publish', 'true')
ON CONFLICT (tenant_id, config_key) 
DO UPDATE SET 
    config_value = EXCLUDED.config_value,
    updated_at = now();

-- Tenant de prueba 2
INSERT INTO tenant_config (tenant_id, config_key, config_value)
VALUES 
    ('00000000-0000-0000-0000-000000000002', 'catalog.stock_policy', 'VALIDATE_STOCK'),
    ('00000000-0000-0000-0000-000000000002', 'catalog.auto_publish', 'false')
ON CONFLICT (tenant_id, config_key) 
DO UPDATE SET 
    config_value = EXCLUDED.config_value,
    updated_at = now();

SELECT 'Datos de prueba insertados correctamente' as status;
EOF
    
    echo -e "${GREEN}✅ Datos de prueba insertados${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Inicialización completada!${NC}"
echo ""
echo "Puedes verificar los datos con:"
echo "  psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c 'SELECT * FROM tenant_config;'"
