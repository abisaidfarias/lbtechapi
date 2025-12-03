#!/bin/bash

# Script para verificar la configuración de AWS para LBTech API
# Creado: 2025-12-01

echo "🔍 Verificando configuración de AWS para LBTech API"
echo "=================================================="
echo ""

# Colores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variables
AWS_REGION="us-east-1"
AWS_ACCOUNT_ID="438758934896"
ECR_REPO="lbtechapidev"
CONTAINER_NAME="lbtechcontainer-dev"

# 1. Verificar AWS CLI
echo "1️⃣  Verificando AWS CLI..."
if command -v aws &> /dev/null; then
    AWS_VERSION=$(aws --version)
    echo -e "${GREEN}✅ AWS CLI instalado: $AWS_VERSION${NC}"
else
    echo -e "${RED}❌ AWS CLI no instalado${NC}"
    echo "   Instalar: brew install awscli"
    exit 1
fi
echo ""

# 2. Verificar credenciales AWS
echo "2️⃣  Verificando credenciales AWS..."
CALLER_IDENTITY=$(aws sts get-caller-identity 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Credenciales AWS configuradas${NC}"
    CURRENT_ACCOUNT=$(echo $CALLER_IDENTITY | jq -r '.Account' 2>/dev/null || echo "No se pudo obtener")
    CURRENT_USER=$(echo $CALLER_IDENTITY | jq -r '.Arn' 2>/dev/null || echo "No se pudo obtener")
    echo "   Account ID: $CURRENT_ACCOUNT"
    echo "   User/Role: $CURRENT_USER"
    
    if [ "$CURRENT_ACCOUNT" != "$AWS_ACCOUNT_ID" ]; then
        echo -e "${YELLOW}⚠️  ADVERTENCIA: Account ID diferente al esperado${NC}"
        echo "   Esperado: $AWS_ACCOUNT_ID"
        echo "   Actual: $CURRENT_ACCOUNT"
    fi
else
    echo -e "${RED}❌ No se pudieron verificar credenciales AWS${NC}"
    echo "   Configurar: aws configure"
    exit 1
fi
echo ""

# 3. Verificar región configurada
echo "3️⃣  Verificando región AWS..."
CONFIGURED_REGION=$(aws configure get region)
if [ "$CONFIGURED_REGION" == "$AWS_REGION" ]; then
    echo -e "${GREEN}✅ Región correcta: $AWS_REGION${NC}"
else
    echo -e "${YELLOW}⚠️  Región diferente: $CONFIGURED_REGION (esperada: $AWS_REGION)${NC}"
    echo "   Cambiar: aws configure set region $AWS_REGION"
fi
echo ""

# 4. Verificar repositorio ECR
echo "4️⃣  Verificando repositorio ECR..."
ECR_CHECK=$(aws ecr describe-repositories --repository-names $ECR_REPO --region $AWS_REGION 2>&1)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Repositorio ECR existe: $ECR_REPO${NC}"
    
    # Obtener URI del repositorio
    ECR_URI=$(echo $ECR_CHECK | jq -r '.repositories[0].repositoryUri' 2>/dev/null)
    echo "   URI: $ECR_URI"
    
    # Contar imágenes
    IMAGE_COUNT=$(aws ecr list-images --repository-name $ECR_REPO --region $AWS_REGION 2>&1 | jq '.imageIds | length' 2>/dev/null || echo "0")
    echo "   Imágenes: $IMAGE_COUNT"
else
    echo -e "${RED}❌ Repositorio ECR no existe: $ECR_REPO${NC}"
    echo "   Crear con: aws ecr create-repository --repository-name $ECR_REPO --region $AWS_REGION"
fi
echo ""

# 5. Verificar Docker
echo "5️⃣  Verificando Docker..."
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version)
    echo -e "${GREEN}✅ Docker instalado: $DOCKER_VERSION${NC}"
    
    # Verificar si Docker está corriendo
    if docker info &> /dev/null; then
        echo -e "${GREEN}✅ Docker daemon corriendo${NC}"
    else
        echo -e "${RED}❌ Docker daemon no está corriendo${NC}"
        echo "   Iniciar Docker Desktop"
    fi
else
    echo -e "${RED}❌ Docker no instalado${NC}"
    echo "   Instalar Docker Desktop"
fi
echo ""

# 6. Verificar CodeBuild
echo "6️⃣  Verificando proyectos CodeBuild..."
CODEBUILD_PROJECTS=$(aws codebuild list-projects --region $AWS_REGION 2>&1)
if [ $? -eq 0 ]; then
    PROJECT_COUNT=$(echo $CODEBUILD_PROJECTS | jq '.projects | length' 2>/dev/null || echo "0")
    echo -e "${GREEN}✅ CodeBuild accesible${NC}"
    echo "   Proyectos encontrados: $PROJECT_COUNT"
    
    if [ "$PROJECT_COUNT" -gt 0 ]; then
        echo "   Proyectos:"
        echo $CODEBUILD_PROJECTS | jq -r '.projects[]' 2>/dev/null | while read project; do
            echo "     - $project"
        done
    fi
else
    echo -e "${YELLOW}⚠️  No se pudo acceder a CodeBuild${NC}"
fi
echo ""

# 7. Verificar CodePipeline
echo "7️⃣  Verificando pipelines..."
PIPELINES=$(aws codepipeline list-pipelines --region $AWS_REGION 2>&1)
if [ $? -eq 0 ]; then
    PIPELINE_COUNT=$(echo $PIPELINES | jq '.pipelines | length' 2>/dev/null || echo "0")
    echo -e "${GREEN}✅ CodePipeline accesible${NC}"
    echo "   Pipelines encontrados: $PIPELINE_COUNT"
    
    if [ "$PIPELINE_COUNT" -gt 0 ]; then
        echo "   Pipelines:"
        echo $PIPELINES | jq -r '.pipelines[].name' 2>/dev/null | while read pipeline; do
            echo "     - $pipeline"
        done
    fi
else
    echo -e "${YELLOW}⚠️  No se pudo acceder a CodePipeline${NC}"
fi
echo ""

# 8. Verificar archivo .env
echo "8️⃣  Verificando configuración local..."
if [ -f ".env" ]; then
    echo -e "${GREEN}✅ Archivo .env existe${NC}"
else
    echo -e "${YELLOW}⚠️  Archivo .env no encontrado${NC}"
    echo "   Crear archivo .env con las variables necesarias"
fi
echo ""

# 9. Verificar Dockerfile
echo "9️⃣  Verificando Dockerfile..."
if [ -f "Dockerfile" ]; then
    echo -e "${GREEN}✅ Dockerfile existe${NC}"
else
    echo -e "${RED}❌ Dockerfile no encontrado${NC}"
fi
echo ""

# 10. Verificar archivos de deployment
echo "🔟 Verificando archivos de deployment..."
FILES=("buildspec_build.yml" "appspec.yml" "Dockerrun.aws.json")
for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ $file${NC}"
    else
        echo -e "${RED}❌ $file no encontrado${NC}"
    fi
done
echo ""

# Resumen final
echo "=================================================="
echo "📊 RESUMEN"
echo "=================================================="
echo ""
echo "Configuración detectada:"
echo "  • AWS Account: $CURRENT_ACCOUNT"
echo "  • Región: $AWS_REGION"
echo "  • ECR Repo: $ECR_REPO"
echo "  • Container: $CONTAINER_NAME"
echo ""
echo "Próximos pasos sugeridos:"
echo "  1. Revisar AWS_DEPLOYMENT_GUIDE.md"
echo "  2. Actualizar Dockerfile si es necesario"
echo "  3. Configurar variables en CodeBuild"
echo "  4. Probar build local: docker build -t $ECR_REPO:latest ."
echo "  5. Push a ECR o activar pipeline con git push"
echo ""
echo "=================================================="

