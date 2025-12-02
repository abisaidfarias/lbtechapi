#!/bin/bash

# Script para subir secretos a AWS Parameter Store
# Uso: ./scripts/upload-secrets.sh <environment> <env-file>
# Ejemplo: ./scripts/upload-secrets.sh dev .env

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Verificar argumentos
if [ $# -lt 2 ]; then
    echo -e "${RED}❌ Error: Faltan argumentos${NC}"
    echo "Uso: $0 <environment> <env-file>"
    echo "Ejemplo: $0 dev .env"
    exit 1
fi

ENVIRONMENT=$1
ENV_FILE=$2
REGION="us-east-1"
BASE_PATH="/lbtechapi/${ENVIRONMENT}"

# Verificar que el archivo existe
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}❌ Error: El archivo ${ENV_FILE} no existe${NC}"
    exit 1
fi

echo -e "${GREEN}🔐 Subiendo secretos a Parameter Store${NC}"
echo -e "   Ambiente: ${YELLOW}${ENVIRONMENT}${NC}"
echo -e "   Región: ${YELLOW}${REGION}${NC}"
echo -e "   Base Path: ${YELLOW}${BASE_PATH}${NC}"
echo ""

# Contador
SUCCESS=0
FAILED=0

# Leer y procesar cada línea del .env
while IFS='=' read -r key value; do
    # Ignorar líneas vacías y comentarios
    [[ -z "$key" || "$key" =~ ^#.* ]] && continue
    
    # Limpiar espacios
    key=$(echo "$key" | xargs)
    value=$(echo "$value" | xargs)
    
    # Nombre completo del parámetro
    PARAM_NAME="${BASE_PATH}/${key}"
    
    echo -n "   Subiendo ${key}... "
    
    # Subir a Parameter Store
    if aws ssm put-parameter \
        --name "$PARAM_NAME" \
        --value "$value" \
        --type "SecureString" \
        --overwrite \
        --region "$REGION" \
        --output text > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        ((SUCCESS++))
    else
        echo -e "${RED}✗${NC}"
        ((FAILED++))
    fi
    
done < "$ENV_FILE"

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "   ✅ Exitosos: ${GREEN}${SUCCESS}${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "   ❌ Fallidos: ${RED}${FAILED}${NC}"
fi
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Listar parámetros creados
echo -e "${YELLOW}📋 Parámetros en ${BASE_PATH}:${NC}"
aws ssm get-parameters-by-path \
    --path "$BASE_PATH" \
    --region "$REGION" \
    --query 'Parameters[*].Name' \
    --output table

echo ""
echo -e "${GREEN}🎉 ¡Secretos subidos exitosamente!${NC}"
echo ""
echo -e "${YELLOW}💡 Siguiente paso:${NC}"
echo "   1. Configura ENVIRONMENT=${ENVIRONMENT} en Elastic Beanstalk"
echo "   2. Verifica permisos IAM para SSM en el rol EC2"
echo "   3. Deploy de la aplicación"

