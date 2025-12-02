# 🔐 Secrets Management

## Arquitectura de Ambientes

La aplicación soporta múltiples ambientes con secretos almacenados de forma segura en **AWS Systems Manager Parameter Store**.

### Ambientes Configurados

- **DEV** (`/lbtechapi/dev/`) - Ambiente de desarrollo
- **PROD** (`/lbtechapi/prod/`) - Ambiente de producción (futuro)

---

## 📦 Secretos por Ambiente

Cada ambiente tiene los siguientes secretos encriptados:

```
/lbtechapi/{environment}/
├── SECRET_KEY          - Clave para JWT
├── MONGO_URI           - URI de conexión a MongoDB
├── MONGO_DB            - Nombre de la base de datos
├── EMAIL_FROM          - Email remitente para notificaciones
├── EMAIL_PASSWORD      - Password del email
├── EMAIL_PORT          - Puerto SMTP
├── SMTP_CLIENTE        - Host SMTP
└── MONTH_INTERVAL      - Intervalo de meses para reportes
```

---

## 🚀 Uso en la Aplicación

### Cargar Secretos al Inicio

En `main.go`:

```go
import "github.com/abisaidfarias/lbtechapi/config"

func main() {
    // Cargar secretos al inicio
    secrets, err := config.LoadSecrets()
    if err != nil {
        log.Fatal("Failed to load secrets:", err)
    }
    
    // Los secretos están ahora disponibles en toda la app
}
```

### Obtener un Secreto

```go
import "github.com/abisaidfarias/lbtechapi/config"

// Opción 1: Obtener el struct completo
secrets := config.Get()
mongoURI := secrets.MongoURI

// Opción 2: Obtener un valor específico
secretKey := config.GetValue("SECRET_KEY")
```

---

## 🌍 Configuración de Ambientes

### En AWS Elastic Beanstalk

Configura la variable de ambiente `ENVIRONMENT`:

```bash
# Para DEV
aws elasticbeanstalk update-environment \
  --environment-name lbtechapi \
  --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=ENVIRONMENT,Value=dev

# Para PROD (futuro)
aws elasticbeanstalk update-environment \
  --environment-name lbtechapi-prod \
  --option-settings Namespace=aws:elasticbeanstalk:application:environment,OptionName=ENVIRONMENT,Value=prod
```

### Desarrollo Local

Para desarrollo local, usa `.env`:

```bash
# Crear .env local
cp .env.example .env

# Configurar para usar .env local
export USE_LOCAL_ENV=true
```

---

## 🔧 Gestión de Secretos

### Crear Secretos en Parameter Store (DEV)

```bash
aws ssm put-parameter \
  --name "/lbtechapi/dev/SECRET_KEY" \
  --value "tu-secret-key" \
  --type "SecureString" \
  --region us-east-1
```

### Actualizar un Secreto

```bash
aws ssm put-parameter \
  --name "/lbtechapi/dev/MONGO_URI" \
  --value "nueva-uri" \
  --type "SecureString" \
  --overwrite \
  --region us-east-1
```

### Listar Secretos de un Ambiente

```bash
aws ssm get-parameters-by-path \
  --path "/lbtechapi/dev" \
  --with-decryption \
  --region us-east-1
```

### Ver un Secreto Específico

```bash
aws ssm get-parameter \
  --name "/lbtechapi/dev/SECRET_KEY" \
  --with-decryption \
  --region us-east-1
```

---

## 🔒 Seguridad

### Ventajas

✅ **Encriptación en reposo** - Los secretos se encriptan automáticamente con AWS KMS  
✅ **Auditoría** - Todos los accesos se registran en CloudTrail  
✅ **No en código** - Los secretos nunca están en el repositorio  
✅ **IAM control** - Permisos granulares por ambiente  
✅ **Rotación fácil** - Actualiza secretos sin redeployar  

### Permisos IAM Necesarios

El rol EC2 de Elastic Beanstalk necesita estos permisos:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters",
        "ssm:GetParametersByPath"
      ],
      "Resource": "arn:aws:ssm:us-east-1:*:parameter/lbtechapi/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt"
      ],
      "Resource": "*"
    }
  ]
}
```

---

## 📝 Migración desde .env

### Antes (usando .env):

```go
secretKey := os.Getenv("SECRET_KEY")
```

### Después (usando Parameter Store):

```go
secretKey := config.GetValue("SECRET_KEY")
```

---

## 🐛 Troubleshooting

### Error: "Secrets not loaded"

**Causa**: `LoadSecrets()` no fue llamado en `main.go`  
**Solución**: Asegúrate de llamar `config.LoadSecrets()` al inicio de la aplicación

### Error: "Failed to get parameter"

**Causa**: El parámetro no existe en Parameter Store o faltan permisos IAM  
**Solución**:
1. Verifica que el parámetro existe: `aws ssm get-parameter --name "/lbtechapi/dev/SECRET_KEY"`
2. Verifica permisos IAM del rol EC2

### Error: "Unable to load AWS SDK config"

**Causa**: La aplicación no puede autenticarse con AWS  
**Solución**: 
- En AWS: Verifica que el rol EC2 esté configurado
- En local: Configura credenciales con `aws configure`

---

## 🔄 Rotación de Secretos

Para rotar un secreto sin downtime:

1. **Actualiza en Parameter Store:**
   ```bash
   aws ssm put-parameter \
     --name "/lbtechapi/dev/SECRET_KEY" \
     --value "nuevo-valor" \
     --type "SecureString" \
     --overwrite
   ```

2. **Reinicia la aplicación:**
   ```bash
   eb deploy
   ```

Los secretos se cargan al inicio de la aplicación.

---

## 📊 Costos

- **Parameter Store (Standard)**: GRATIS
- **KMS Encryption**: ~$1/mes por key
- **Total estimado**: ~$1-2/mes

---

## 🎯 Best Practices

1. ✅ **Nunca** hagas commit de `.env` al repositorio
2. ✅ Usa secretos diferentes para cada ambiente
3. ✅ Rota secretos regularmente (cada 90 días)
4. ✅ Usa nombres descriptivos para los parámetros
5. ✅ Documenta qué hace cada secreto
6. ✅ Revisa logs de CloudTrail para accesos sospechosos

---

## 🆕 Agregar un Nuevo Secreto

### Paso 1: Agregar a `config/secrets.go`

```go
type Secrets struct {
    // ... existing fields
    NewSecret string // Add new field
}

// Update params map in LoadSecrets
params := map[string]*string{
    // ... existing params
    "NEW_SECRET": &secrets.NewSecret,
}

// Update GetValue switch
case "NEW_SECRET":
    return secrets.NewSecret
```

### Paso 2: Crear en Parameter Store

```bash
aws ssm put-parameter \
  --name "/lbtechapi/dev/NEW_SECRET" \
  --value "valor" \
  --type "SecureString" \
  --region us-east-1
```

### Paso 3: Deploy

```bash
eb deploy
```

---

**Actualizado**: Diciembre 2025  
**Ambiente**: DEV  
**Región**: us-east-1

