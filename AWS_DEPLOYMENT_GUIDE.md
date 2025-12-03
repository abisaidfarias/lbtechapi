# 🚀 Guía de Deployment AWS - LBTech API

## 📋 Configuración Actual

### 🏗️ Infraestructura AWS Identificada

| Componente | Configuración |
|------------|---------------|
| **Región AWS** | `us-east-1` |
| **AWS Account ID** | `438758934896` |
| **ECR Repository** | `lbtechapidev` |
| **Container Name** | `lbtechcontainer-dev` |
| **Puerto** | `8080` |
| **Repositorio Git** | https://github.com/abisaidfarias/lbtechapi.git |

### 🔧 Servicios AWS Utilizados

1. **Amazon ECR (Elastic Container Registry)**
   - Almacena las imágenes Docker
   - URI: `438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest`

2. **AWS CodeBuild**
   - Construye la imagen Docker
   - Archivo: `buildspec_build.yml`
   - Push a ECR automático

3. **AWS CodeDeploy**
   - Deployment de la aplicación
   - Archivo: `appspec.yml`
   - Scripts en `/codedeploy/`

4. **AWS Elastic Beanstalk** (opcional)
   - Archivo: `Dockerrun.aws.json`
   - Apunta a imagen de Docker Hub (desactualizado)

---

## 📁 Archivos de Configuración

### 1. `Dockerfile`
```dockerfile
FROM golang:latest
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
ENV PORT 8080
RUN go build
RUN find . -name "*.go" -type f -delete
EXPOSE $PORT
CMD ["./lbtechapi"]
```

### 2. `buildspec_build.yml` (AWS CodeBuild)
- Login a ECR
- Build de imagen Docker
- Tag de imagen
- Push a ECR
- Genera `imagedefinitions.json`

### 3. `appspec.yml` (AWS CodeDeploy)
Scripts de deployment:
- `BeforeInstall.sh` - (vacío actualmente)
- `ApplicationStop.sh` - Detiene la app
- `ApplicationStart.sh` - Inicia la app
- `ValidateService.sh` - Valida que funcione (curl localhost:8080)

### 4. `Dockerrun.aws.json` (Elastic Beanstalk)
⚠️ **DESACTUALIZADO** - Apunta a Docker Hub en lugar de ECR

---

## ✅ Verificación Pre-Deployment

### 1. Verificar Acceso AWS

```bash
# Verificar credenciales AWS
aws sts get-caller-identity

# Debe mostrar:
# Account: 438758934896
# Region: us-east-1
```

### 2. Verificar Acceso a ECR

```bash
# Login a ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  438758934896.dkr.ecr.us-east-1.amazonaws.com
```

### 3. Verificar Repositorio ECR Existe

```bash
# Listar repositorios
aws ecr describe-repositories --region us-east-1

# Buscar: lbtechapidev
```

### 4. Verificar CodePipeline/CodeBuild

```bash
# Listar pipelines
aws codepipeline list-pipelines --region us-east-1

# Listar proyectos CodeBuild
aws codebuild list-projects --region us-east-1
```

---

## 🚀 Pasos para Deployment

### Opción 1: Deployment Manual (Recomendado para primera vez)

#### 1. **Build Local de la Imagen**

```bash
# En el directorio del proyecto
cd /Users/abisaid.farias/Documents/lbtechapi/lbtechapi

# Build de la imagen
docker build -t lbtechapidev:latest .

# Verificar que funcionó
docker images | grep lbtechapidev
```

#### 2. **Login a AWS ECR**

```bash
# Login
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  438758934896.dkr.ecr.us-east-1.amazonaws.com
```

#### 3. **Tag y Push a ECR**

```bash
# Tag de la imagen
docker tag lbtechapidev:latest \
  438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest

# Push a ECR
docker push 438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest
```

#### 4. **Verificar en AWS Console**
- Ir a ECR en us-east-1
- Verificar que la imagen se subió correctamente

---

### Opción 2: Deployment Automático (CI/CD)

#### 1. **Configurar Variables de Entorno en CodeBuild**

Variables necesarias:
- `AWS_DEFAULT_REGION`: us-east-1
- `AWS_ACCOUNT_ID`: 438758934896
- `IMAGE_REPO_NAME`: lbtechapidev
- `IMAGE_TAG`: latest (o usar $CODEBUILD_RESOLVED_SOURCE_VERSION)
- `CONTAINER_NAME`: lbtechcontainer-dev
- `DockerFilePath`: Dockerfile

#### 2. **Push a GitHub**

```bash
# Asegúrate de estar en rama dev
git branch

# Push tus cambios
git push origin dev
```

#### 3. **CodePipeline se Activa Automáticamente**
- Detecta cambios en GitHub
- Ejecuta CodeBuild
- Build de imagen Docker
- Push a ECR
- CodeDeploy despliega la nueva versión

---

## 🔍 Verificación Post-Deployment

### 1. Verificar Imagen en ECR

```bash
aws ecr describe-images \
  --repository-name lbtechapidev \
  --region us-east-1
```

### 2. Verificar Deployment (si usas ECS/Fargate)

```bash
# Listar clusters
aws ecs list-clusters --region us-east-1

# Describir servicios
aws ecs list-services --cluster <tu-cluster> --region us-east-1
```

### 3. Probar la API

```bash
# Obtener URL del servicio (desde AWS Console)
# Luego probar:
curl https://tu-api-url.amazonaws.com/health

# Probar Swagger
curl https://tu-api-url.amazonaws.com/swagger/index.html
```

---

## ⚠️ Problemas Comunes y Soluciones

### 1. **Error: "no basic auth credentials"**
```bash
# Solución: Re-login a ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  438758934896.dkr.ecr.us-east-1.amazonaws.com
```

### 2. **Error: "repository does not exist"**
```bash
# Crear repositorio ECR
aws ecr create-repository \
  --repository-name lbtechapidev \
  --region us-east-1
```

### 3. **Error: "AccessDeniedException"**
- Verificar credenciales AWS
- Verificar permisos IAM para ECR, CodeBuild, CodeDeploy
- Necesitas permisos: ECR Full Access, CodeBuild, CodeDeploy

### 4. **Dockerfile no compila**
```bash
# Verificar que Go esté instalado en local
go version

# Build local primero
go mod download
go build
```

---

## 🔄 Actualización del Dockerfile (Recomendado)

El Dockerfile actual usa `golang:latest`. Recomiendo usar una versión específica:

```dockerfile
# Cambiar de:
FROM golang:latest

# A:
FROM golang:1.25-alpine

# Y usar multi-stage build para imagen más pequeña:
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o lbtechapi

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/lbtechapi .
COPY --from=builder /app/docs ./docs
EXPOSE 8080
CMD ["./lbtechapi"]
```

---

## 📊 Monitoreo y Logs

### CloudWatch Logs
```bash
# Ver logs del grupo
aws logs describe-log-groups --region us-east-1

# Ver streams
aws logs describe-log-streams \
  --log-group-name /aws/codebuild/lbtech-api \
  --region us-east-1
```

### Métricas
- CloudWatch Metrics para ECR
- CloudWatch Metrics para CodeBuild
- Application Insights (si está configurado)

---

## 🔐 Variables de Entorno Necesarias

Crear archivo `.env` (ya ignorado en git):

```bash
# MongoDB
MONGODB_URI=<tu-mongo-uri>
MONGODB_DATABASE=<tu-database>

# JWT
JWT_SECRET=<tu-jwt-secret>

# AWS
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<tu-access-key>
AWS_SECRET_ACCESS_KEY=<tu-secret-key>

# S3 (si usas storage)
S3_BUCKET=<tu-bucket>

# Email (si usas notificaciones)
SMTP_HOST=<smtp-host>
SMTP_PORT=<smtp-port>
SMTP_USER=<smtp-user>
SMTP_PASSWORD=<smtp-password>
```

---

## 📋 Checklist Pre-Deployment

- [ ] AWS CLI instalado y configurado
- [ ] Credenciales AWS verificadas (`aws sts get-caller-identity`)
- [ ] Docker instalado y corriendo
- [ ] Código compilado localmente sin errores
- [ ] Variables de entorno configuradas
- [ ] ECR repository existe
- [ ] Permisos IAM correctos
- [ ] CodePipeline configurado (si aplica)
- [ ] MongoDB accesible desde AWS
- [ ] Swagger docs generados (`swag init`)
- [ ] Tests pasando (si los hay)
- [ ] .env configurado (NO commitear)

---

## 🆘 Contactos y Recursos

### AWS Console Links (us-east-1)
- ECR: https://console.aws.amazon.com/ecr/repositories?region=us-east-1
- CodeBuild: https://console.aws.amazon.com/codebuild/home?region=us-east-1
- CodePipeline: https://console.aws.amazon.com/codesuite/codepipeline/pipelines?region=us-east-1
- CloudWatch: https://console.aws.amazon.com/cloudwatch/home?region=us-east-1

### Documentación
- [AWS ECR User Guide](https://docs.aws.amazon.com/ecr/)
- [AWS CodeBuild](https://docs.aws.amazon.com/codebuild/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)

---

## 📝 Notas

1. **Última actualización**: Hace 3 años (2022-09-06)
2. **Rama activa**: `dev` (tiene commits más recientes)
3. **Swagger**: Recién agregado - asegúrate de incluir `/docs` en la imagen
4. **Go version**: Actualizada a 1.25.4
5. **Imagen actual**: Incluye Swagger UI completo (46 endpoints documentados)

---

**Creado**: 2025-12-01  
**Por**: Cursor AI Assistant  
**Proyecto**: LBTech API

