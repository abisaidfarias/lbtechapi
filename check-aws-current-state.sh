#!/bin/bash

# Script para SOLO VERIFICAR el estado actual de AWS
# NO HACE CAMBIOS - Solo consulta
# Creado: 2025-12-01

echo "🔍 Verificando Estado Actual de AWS (SOLO LECTURA)"
echo "=================================================="
echo ""

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

AWS_REGION="us-east-1"
ECR_REPO="lbtechapidev"

echo "⚠️  Este script SOLO CONSULTA, no hace cambios"
echo ""

# Función para verificar si AWS CLI está instalado
check_aws_cli() {
    if ! command -v aws &> /dev/null; then
        echo -e "${RED}❌ AWS CLI no instalado${NC}"
        echo ""
        echo "Para instalar:"
        echo "  brew install awscli"
        echo ""
        echo "Después configura credenciales:"
        echo "  aws configure"
        exit 1
    fi
}

# Función para verificar credenciales
check_credentials() {
    if ! aws sts get-caller-identity &> /dev/null; then
        echo -e "${RED}❌ Credenciales AWS no configuradas${NC}"
        echo ""
        echo "Configurar con: aws configure"
        exit 1
    fi
}

check_aws_cli
check_credentials

echo -e "${GREEN}✅ AWS CLI configurado correctamente${NC}"
echo ""

# 1. Ver qué servicios están CORRIENDO (y costando dinero)
echo "💰 SERVICIOS ACTIVOS (que pueden generar costos)"
echo "=================================================="
echo ""

# EC2 Instances
echo "🖥️  EC2 Instances:"
INSTANCES=$(aws ec2 describe-instances --region $AWS_REGION \
    --filters "Name=instance-state-name,Values=running" \
    --query 'Reservations[*].Instances[*].[InstanceId,InstanceType,State.Name,Tags[?Key==`Name`].Value|[0]]' \
    --output text 2>/dev/null)

if [ -z "$INSTANCES" ]; then
    echo -e "  ${GREEN}✅ No hay instancias EC2 corriendo${NC}"
else
    echo -e "  ${YELLOW}⚠️  HAY INSTANCIAS CORRIENDO (generan costo):${NC}"
    echo "$INSTANCES" | while read line; do
        echo "     $line"
    done
fi
echo ""

# ECS Services
echo "🐳 ECS/Fargate Services:"
CLUSTERS=$(aws ecs list-clusters --region $AWS_REGION --query 'clusterArns[*]' --output text 2>/dev/null)

if [ -z "$CLUSTERS" ]; then
    echo -e "  ${GREEN}✅ No hay clusters ECS${NC}"
else
    echo -e "  ${BLUE}ℹ️  Clusters ECS encontrados:${NC}"
    for cluster in $CLUSTERS; do
        cluster_name=$(basename $cluster)
        echo "     Cluster: $cluster_name"
        
        SERVICES=$(aws ecs list-services --cluster $cluster --region $AWS_REGION --query 'serviceArns[*]' --output text 2>/dev/null)
        if [ ! -z "$SERVICES" ]; then
            echo -e "     ${YELLOW}⚠️  Servicios corriendo:${NC}"
            for service in $SERVICES; do
                service_name=$(basename $service)
                TASKS=$(aws ecs describe-services --cluster $cluster --services $service --region $AWS_REGION \
                    --query 'services[0].[serviceName,runningCount,desiredCount]' --output text 2>/dev/null)
                echo "       - $TASKS"
            done
        fi
    done
fi
echo ""

# Elastic Beanstalk
echo "🌱 Elastic Beanstalk Environments:"
EB_APPS=$(aws elasticbeanstalk describe-applications --region $AWS_REGION \
    --query 'Applications[*].ApplicationName' --output text 2>/dev/null)

if [ -z "$EB_APPS" ]; then
    echo -e "  ${GREEN}✅ No hay aplicaciones Elastic Beanstalk${NC}"
else
    echo -e "  ${BLUE}ℹ️  Aplicaciones encontradas:${NC}"
    for app in $EB_APPS; do
        echo "     App: $app"
        ENVS=$(aws elasticbeanstalk describe-environments --application-name $app --region $AWS_REGION \
            --query 'Environments[*].[EnvironmentName,Status,Health]' --output text 2>/dev/null)
        if [ ! -z "$ENVS" ]; then
            echo "$ENVS" | while read env; do
                echo "       - $env"
            done
        fi
    done
fi
echo ""

# Lambda Functions
echo "⚡ Lambda Functions:"
LAMBDAS=$(aws lambda list-functions --region $AWS_REGION --query 'Functions[*].FunctionName' --output text 2>/dev/null)
LAMBDA_COUNT=$(echo $LAMBDAS | wc -w | xargs)
echo -e "  ${BLUE}ℹ️  Funciones Lambda: $LAMBDA_COUNT${NC}"
if [ "$LAMBDA_COUNT" -gt 0 ] && [ "$LAMBDA_COUNT" -lt 10 ]; then
    echo "$LAMBDAS" | tr '\t' '\n' | while read lambda; do
        echo "     - $lambda"
    done
fi
echo ""

# RDS Databases
echo "🗄️  RDS Databases:"
RDS=$(aws rds describe-db-instances --region $AWS_REGION \
    --query 'DBInstances[*].[DBInstanceIdentifier,DBInstanceClass,DBInstanceStatus]' \
    --output text 2>/dev/null)

if [ -z "$RDS" ]; then
    echo -e "  ${GREEN}✅ No hay bases de datos RDS${NC}"
else
    echo -e "  ${YELLOW}⚠️  BASES DE DATOS CORRIENDO (generan costo):${NC}"
    echo "$RDS" | while read line; do
        echo "     $line"
    done
fi
echo ""

echo ""
echo "📦 RECURSOS DE ALMACENAMIENTO"
echo "=================================================="
echo ""

# ECR - Imágenes almacenadas
echo "🐳 ECR (Elastic Container Registry):"
if aws ecr describe-repositories --repository-names $ECR_REPO --region $AWS_REGION &> /dev/null; then
    IMAGES=$(aws ecr list-images --repository-name $ECR_REPO --region $AWS_REGION \
        --query 'imageIds[*]' --output json 2>/dev/null)
    IMAGE_COUNT=$(echo $IMAGES | jq 'length' 2>/dev/null || echo "0")
    
    echo -e "  ${BLUE}ℹ️  Repositorio: $ECR_REPO${NC}"
    echo "     Imágenes almacenadas: $IMAGE_COUNT"
    
    # Calcular tamaño aproximado
    SIZE=$(aws ecr describe-images --repository-name $ECR_REPO --region $AWS_REGION \
        --query 'sum(imageDetails[*].imageSizeInBytes)' --output text 2>/dev/null || echo "0")
    SIZE_MB=$((SIZE / 1024 / 1024))
    echo "     Tamaño total: ~${SIZE_MB}MB"
    
    if [ $SIZE_MB -gt 1000 ]; then
        echo -e "     ${YELLOW}⚠️  Considera limpiar imágenes viejas${NC}"
    fi
else
    echo -e "  ${GREEN}✅ No hay repositorio ECR o está vacío${NC}"
fi
echo ""

# S3 Buckets
echo "🪣 S3 Buckets:"
S3_BUCKETS=$(aws s3 ls 2>/dev/null | wc -l)
echo -e "  ${BLUE}ℹ️  Buckets S3: $S3_BUCKETS${NC}"
echo ""

echo ""
echo "🔧 CI/CD PIPELINE (solo se cobra cuando se ejecuta)"
echo "=================================================="
echo ""

# CodePipeline
echo "🔄 CodePipeline:"
PIPELINES=$(aws codepipeline list-pipelines --region $AWS_REGION \
    --query 'pipelines[*].name' --output text 2>/dev/null)

if [ -z "$PIPELINES" ]; then
    echo -e "  ${GREEN}✅ No hay pipelines configurados${NC}"
else
    echo -e "  ${BLUE}ℹ️  Pipelines encontrados:${NC}"
    for pipeline in $PIPELINES; do
        echo "     - $pipeline"
        LAST_EXEC=$(aws codepipeline get-pipeline-state --name $pipeline --region $AWS_REGION \
            --query 'updated' --output text 2>/dev/null)
        echo "       Última ejecución: $LAST_EXEC"
    done
fi
echo ""

# CodeBuild
echo "🏗️  CodeBuild Projects:"
BUILD_PROJECTS=$(aws codebuild list-projects --region $AWS_REGION \
    --query 'projects[*]' --output text 2>/dev/null)

if [ -z "$BUILD_PROJECTS" ]; then
    echo -e "  ${GREEN}✅ No hay proyectos CodeBuild${NC}"
else
    echo -e "  ${BLUE}ℹ️  Proyectos encontrados:${NC}"
    echo "$BUILD_PROJECTS" | tr '\t' '\n' | while read project; do
        echo "     - $project"
    done
fi
echo ""

echo ""
echo "💵 ESTIMACIÓN DE COSTOS MENSUALES"
echo "=================================================="
echo ""
echo "Basado en los recursos encontrados:"
echo ""

# Calcular estimación aproximada
TOTAL_COST=0

# EC2
if [ ! -z "$INSTANCES" ]; then
    INSTANCE_COUNT=$(echo "$INSTANCES" | wc -l)
    echo "  • EC2 Instances ($INSTANCE_COUNT): ~\$10-50/mes cada una"
    TOTAL_COST=$((TOTAL_COST + INSTANCE_COUNT * 30))
fi

# ECS/Fargate
if [ ! -z "$CLUSTERS" ]; then
    echo "  • ECS/Fargate: ~\$20-100/mes (depende del uso)"
    TOTAL_COST=$((TOTAL_COST + 50))
fi

# RDS
if [ ! -z "$RDS" ]; then
    RDS_COUNT=$(echo "$RDS" | wc -l)
    echo "  • RDS ($RDS_COUNT): ~\$15-100/mes cada una"
    TOTAL_COST=$((TOTAL_COST + RDS_COUNT * 50))
fi

# ECR
if [ $IMAGE_COUNT -gt 0 ]; then
    ECR_COST=$(echo "scale=2; $SIZE_MB * 0.0001" | bc 2>/dev/null || echo "0.01")
    echo "  • ECR Storage (${SIZE_MB}MB): ~\$$ECR_COST/mes"
fi

# CodePipeline
if [ ! -z "$PIPELINES" ]; then
    echo "  • CodePipeline: \$1/mes por pipeline activo"
    TOTAL_COST=$((TOTAL_COST + 1))
fi

echo ""
if [ $TOTAL_COST -eq 0 ]; then
    echo -e "${GREEN}✅ COSTO ESTIMADO: ~\$0-5/mes (Free Tier)${NC}"
    echo ""
    echo "Solo pagas por:"
    echo "  - ECR storage (muy bajo)"
    echo "  - Ejecuciones de CodeBuild (cuando corren)"
    echo "  - Data transfer (mínimo)"
else
    echo -e "${YELLOW}⚠️  COSTO ESTIMADO: ~\$$TOTAL_COST/mes${NC}"
fi

echo ""
echo "=================================================="
echo "📊 RESUMEN"
echo "=================================================="
echo ""
echo "Para NO generar costos adicionales:"
echo ""
echo "✅ OPCIONES ECONÓMICAS para deployment:"
echo ""
echo "1. 🐳 SOLO actualizar imagen en ECR (casi gratis)"
echo "   - Build local → Push a ECR"
echo "   - Costo: ~\$0.10/mes por almacenamiento"
echo ""
echo "2. ⚡ Usar Lambda + API Gateway (Free Tier generoso)"
echo "   - 1M requests gratis/mes"
echo "   - Bueno para tráfico bajo/medio"
echo ""
echo "3. 🌱 Elastic Beanstalk single instance (dev)"
echo "   - Si ya existe, solo actualizar"
echo "   - ~\$10-15/mes con t2.micro (Free Tier 1er año)"
echo ""
echo "❌ EVITAR (más costoso):"
echo "   - EC2 grandes (t2.medium+)"
echo "   - RDS (usa MongoDB Atlas gratis)"
echo "   - Load Balancers"
echo "   - NAT Gateways"
echo ""
echo "💡 RECOMENDACIÓN:"
echo "   Ve a AWS Console → Billing Dashboard"
echo "   para ver costos reales actuales"
echo ""
echo "=================================================="

