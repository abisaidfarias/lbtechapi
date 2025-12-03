# 🚀 Quick Start - Deployment a AWS

## 📋 Resumen de tu Infraestructura

Tu aplicación está configurada para deployarse en **AWS** usando:

- **AWS ECR** (Elastic Container Registry) - Para almacenar imágenes Docker
- **AWS CodeBuild** - Para construir las imágenes
- **AWS CodeDeploy** - Para deployar la aplicación
- **AWS Account**: `438758934896`
- **Región**: `us-east-1` (Virginia del Norte)
- **Repositorio ECR**: `lbtechapidev`
- **GitHub**: https://github.com/abisaidfarias/lbtechapi.git

---

## 🎯 Pasos Rápidos para Reactivar el Deployment

### Paso 1: Instalar AWS CLI

```bash
# macOS
brew install awscli

# Verificar instalación
aws --version
```

### Paso 2: Configurar Credenciales AWS

```bash
# Configurar AWS CLI
aws configure

# Te pedirá:
# AWS Access Key ID: [tu-access-key]
# AWS Secret Access Key: [tu-secret-key]
# Default region name: us-east-1
# Default output format: json
```

💡 **¿Dónde obtener las credenciales?**
1. Ir a AWS Console: https://console.aws.amazon.com
2. IAM → Users → Tu usuario → Security credentials
3. Crear nuevo Access Key si es necesario

### Paso 3: Verificar Configuración

```bash
# Ejecutar script de verificación
cd /Users/abisaid.farias/Documents/lbtechapi/lbtechapi
./check-aws-setup.sh
```

Este script te dirá:
- ✅ Qué está configurado correctamente
- ❌ Qué falta por configurar
- 💡 Recomendaciones

### Paso 4: Opciones de Deployment

#### Opción A: Deployment Manual (Recomendado para primera vez)

```bash
# 1. Build de la imagen Docker
docker build -t lbtechapidev:latest .

# 2. Login a ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  438758934896.dkr.ecr.us-east-1.amazonaws.com

# 3. Tag de la imagen
docker tag lbtechapidev:latest \
  438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest

# 4. Push a ECR
docker push 438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest
```

#### Opción B: Deployment Automático (CI/CD)

```bash
# Simplemente push a GitHub
git push origin dev

# CodePipeline se activará automáticamente y:
# 1. CodeBuild construirá la imagen
# 2. Push a ECR
# 3. CodeDeploy desplegará la nueva versión
```

---

## 📁 Archivos Creados para Ayudarte

| Archivo | Descripción |
|---------|-------------|
| `AWS_DEPLOYMENT_GUIDE.md` | Guía completa y detallada de deployment |
| `check-aws-setup.sh` | Script para verificar tu configuración |
| `QUICK_START_DEPLOYMENT.md` | Esta guía rápida |

---

## ⚠️ Cosas Importantes a Saber

### 1. Variables de Entorno

Tu aplicación necesita variables de entorno. Asegúrate de configurarlas en AWS:

```bash
# Localmente (crear .env)
MONGODB_URI=tu-mongo-uri
JWT_SECRET=tu-jwt-secret
AWS_REGION=us-east-1
S3_BUCKET=tu-bucket
```

### 2. Swagger Incluido

✅ Ya agregamos Swagger a tu aplicación
- Endpoint: `/swagger/index.html`
- 46 endpoints documentados
- Asegúrate de que `/docs` esté incluido en la imagen Docker

### 3. Última Actualización

- Último commit en `dev`: 2022-09-06
- Han pasado ~3 años
- Es posible que necesites:
  - Actualizar dependencias de Go
  - Verificar que los servicios AWS sigan activos
  - Revisar costos de AWS

---

## 🔍 Verificaciones Pre-Deployment

### Checklist

- [ ] AWS CLI instalado (`aws --version`)
- [ ] Credenciales AWS configuradas (`aws sts get-caller-identity`)
- [ ] Docker instalado y corriendo (`docker ps`)
- [ ] Repositorio ECR existe en AWS
- [ ] Variables de entorno configuradas
- [ ] Código compila localmente (`go build`)
- [ ] Swagger docs generados (`swag init`)
- [ ] Cambios commiteados en git

### Verificar Servicios AWS Activos

```bash
# Verificar ECR
aws ecr describe-repositories --region us-east-1

# Verificar CodeBuild
aws codebuild list-projects --region us-east-1

# Verificar CodePipeline
aws codepipeline list-pipelines --region us-east-1
```

---

## 💰 Consideraciones de Costos

Después de 3 años, verifica:

1. **ECR Storage**: Cobran por GB de imágenes almacenadas
   - Limpiar imágenes viejas si hay muchas

2. **Servicios Corriendo**: 
   - Verifica si hay EC2, ECS, o Fargate corriendo
   - Detén lo que no uses

3. **CodePipeline/CodeBuild**:
   - Solo cobran cuando se ejecutan

```bash
# Ver costos actuales
# Ir a: https://console.aws.amazon.com/billing/
```

---

## 🆘 Problemas Comunes

### "Repository does not exist"

```bash
# Crear repositorio ECR
aws ecr create-repository \
  --repository-name lbtechapidev \
  --region us-east-1
```

### "No credentials found"

```bash
# Reconfigurar AWS
aws configure
```

### "Docker daemon not running"

```bash
# Iniciar Docker Desktop desde Aplicaciones
```

### "Build failed"

```bash
# Probar build local primero
go mod download
go build

# Verificar que no haya errores
```

---

## 📞 Próximos Pasos

1. **Instalar AWS CLI**: `brew install awscli`
2. **Configurar credenciales**: `aws configure`
3. **Ejecutar verificación**: `./check-aws-setup.sh`
4. **Leer guía completa**: `AWS_DEPLOYMENT_GUIDE.md`
5. **Hacer deployment**: Seguir Opción A o B arriba

---

## 📚 Recursos Útiles

- [AWS Console](https://console.aws.amazon.com)
- [ECR Console](https://console.aws.amazon.com/ecr/repositories?region=us-east-1)
- [CodeBuild Console](https://console.aws.amazon.com/codebuild/home?region=us-east-1)
- [CodePipeline Console](https://console.aws.amazon.com/codesuite/codepipeline/pipelines?region=us-east-1)
- [Tu Repositorio GitHub](https://github.com/abisaidfarias/lbtechapi)

---

**¿Necesitas ayuda?** 
- Lee `AWS_DEPLOYMENT_GUIDE.md` para detalles completos
- Ejecuta `./check-aws-setup.sh` para diagnóstico
- Revisa logs en AWS CloudWatch

**Última actualización**: 2025-12-01

