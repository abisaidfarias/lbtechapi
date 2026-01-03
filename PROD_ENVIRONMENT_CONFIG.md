# 🚀 Configuración del Ambiente de PRODUCCIÓN

> **Fecha de creación:** Diciembre 2024  
> **Estado actual:** Funcionando ✅

---

## 📋 Resumen de Infraestructura

| Componente | Valor | Estado |
|------------|-------|--------|
| **Dominio** | `lbonetrack.com` | ✅ Activo |
| **API URL** | `https://api.lbonetrack.com` | ✅ |
| **SSL/HTTPS** | Wildcard `*.lbonetrack.com` | ✅ |
| **MongoDB** | Atlas - ProdCluster | ✅ |
| **Pipeline** | GitHub Actions (manual) | ✅ |

---

## 🔐 Credenciales y ARNs

### AWS - Route 53
```
Hosted Zone ID: Z103529712YQQ4D5PZZ0H
Dominio: lbonetrack.com
```

### AWS - Certificado SSL (ACM)
```
Certificate ARN: arn:aws:acm:us-east-1:438758934896:certificate/16402f4c-d385-4939-9f96-a77c72c80e7e
Dominios cubiertos: lbonetrack.com, *.lbonetrack.com
Región: us-east-1
```

### MongoDB Atlas - PROD
```
Cluster: ProdCluster
Database: lbtechprod
URI: mongodb+srv://USERNAME:***@prodcluster.osehmw9.mongodb.net/lbtechprod?retryWrites=true&w=majority
Usuario: [Ver AWS Parameter Store: /lbtechapi/prod/MONGO_URI]
Password: [Ver AWS Parameter Store: /lbtechapi/prod/MONGO_URI]
⚠️ CREDENCIALES: Almacenadas en AWS Systems Manager Parameter Store
```

### Docker Hub
```
Imagen: abisaid/lbtechapi:prod-latest
```

---

## ⚙️ Variables de Ambiente (Elastic Beanstalk)

```bash
ENVIRONMENT=prod
USE_LOCAL_ENV=true
MONGO_URI=[Ver AWS Parameter Store: /lbtechapi/prod/MONGO_URI]
MONGO_DB=lbtechprod
SECRET_KEY=[copiar de DEV]
EMAIL_FROM=soporte@lbtechnology-la.com
EMAIL_PASSWORD=[copiar de DEV]
EMAIL_PORT=587
SMTP_CLIENTE=smtp.office365.com
MONTH_INTERVAL=1
```

---

## 🔄 Comandos para Recrear PROD

### Paso 1: Crear ambiente Elastic Beanstalk

```bash
aws elasticbeanstalk create-environment \
  --application-name lbtechapi \
  --environment-name lbtechapi-prod \
  --description "Produccion - LB Tech API" \
  --solution-stack-name "64bit Amazon Linux 2 v3.4.17 running Docker" \
  --option-settings \
    Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=LoadBalanced \
    Namespace=aws:autoscaling:launchconfiguration,OptionName=InstanceType,Value=t3.micro \
    Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value=aws-elasticbeanstalk-ec2-role \
    Namespace=aws:elasticbeanstalk:environment,OptionName=ServiceRole,Value=aws-elasticbeanstalk-service-role \
    Namespace=aws:elasticbeanstalk:application:environment,OptionName=ENVIRONMENT,Value=prod \
    Namespace=aws:elasticbeanstalk:application:environment,OptionName=USE_LOCAL_ENV,Value=true
```

### Paso 2: Esperar a que esté Ready (~5-10 min)

```bash
# Verificar estado
aws elasticbeanstalk describe-environments --environment-names lbtechapi-prod \
  --query "Environments[0].{Status:Status,Health:Health,URL:EndpointURL}" --output table
```

### Paso 3: Configurar Health Check

```bash
aws elasticbeanstalk update-environment \
  --environment-name lbtechapi-prod \
  --option-settings \
    Namespace=aws:elasticbeanstalk:environment:process:default,OptionName=HealthCheckPath,Value=/health \
    Namespace=aws:elasticbeanstalk:environment:process:default,OptionName=MatcherHTTPCode,Value=200
```

### Paso 4: Configurar variables de ambiente

```bash
# Obtener variables de DEV y aplicar a PROD
# (ejecutar desde el proyecto con el script Python que usamos antes)

# O manualmente crear archivo JSON:
cat > /tmp/prod-vars.json << 'EOF'
[
  {"Namespace": "aws:elasticbeanstalk:application:environment", "OptionName": "MONGO_URI", "Value": "[OBTENER DE AWS PARAMETER STORE: /lbtechapi/prod/MONGO_URI]"},
  {"Namespace": "aws:elasticbeanstalk:application:environment", "OptionName": "MONGO_DB", "Value": "lbtechprod"},
  {"Namespace": "aws:elasticbeanstalk:application:environment", "OptionName": "ENVIRONMENT", "Value": "prod"},
  {"Namespace": "aws:elasticbeanstalk:application:environment", "OptionName": "USE_LOCAL_ENV", "Value": "true"},
  {"Namespace": "aws:elasticbeanstalk:application:environment", "OptionName": "EMAIL_PORT", "Value": "587"},
  {"Namespace": "aws:elasticbeanstalk:application:environment", "OptionName": "MONTH_INTERVAL", "Value": "1"}
]
EOF

aws elasticbeanstalk update-environment \
  --environment-name lbtechapi-prod \
  --option-settings file:///tmp/prod-vars.json
```

### Paso 5: Configurar HTTPS con el certificado SSL

```bash
CERT_ARN="arn:aws:acm:us-east-1:438758934896:certificate/16402f4c-d385-4939-9f96-a77c72c80e7e"

aws elasticbeanstalk update-environment \
  --environment-name lbtechapi-prod \
  --option-settings \
    "Namespace=aws:elb:listener:443,OptionName=ListenerProtocol,Value=HTTPS" \
    "Namespace=aws:elb:listener:443,OptionName=InstancePort,Value=80" \
    "Namespace=aws:elb:listener:443,OptionName=InstanceProtocol,Value=HTTP" \
    "Namespace=aws:elb:listener:443,OptionName=SSLCertificateId,Value=$CERT_ARN" \
    "Namespace=aws:elb:listener:443,OptionName=ListenerEnabled,Value=true"
```

### Paso 6: Actualizar DNS (Route 53)

```bash
# Obtener la URL del nuevo Load Balancer
NEW_LB_URL=$(aws elasticbeanstalk describe-environments --environment-names lbtechapi-prod \
  --query "Environments[0].EndpointURL" --output text)

echo "Nuevo Load Balancer: $NEW_LB_URL"

# Actualizar DNS
HOSTED_ZONE_ID="Z103529712YQQ4D5PZZ0H"

aws route53 change-resource-record-sets \
  --hosted-zone-id $HOSTED_ZONE_ID \
  --change-batch "{
    \"Changes\": [{
      \"Action\": \"UPSERT\",
      \"ResourceRecordSet\": {
        \"Name\": \"api.lbonetrack.com\",
        \"Type\": \"CNAME\",
        \"TTL\": 300,
        \"ResourceRecords\": [{\"Value\": \"$NEW_LB_URL\"}]
      }
    }]
  }"
```

### Paso 7: Hacer deploy

```bash
cd /ruta/al/proyecto
eb deploy lbtechapi-prod
```

---

## 🗄️ Configuración de MongoDB Atlas

### Cluster Info
```
Nombre: ProdCluster
Tier: M0 (Free)
Provider: AWS
Región: N. Virginia (us-east-1)
```

### Network Access
```
IP: 0.0.0.0/0 (Allow from anywhere)
```

### Usuario Admin
```
Email: abisaidfarias@gmail.com
Password: [CREDENCIAL PRIVADA - No almacenar en código]
⚠️ NUNCA commitear contraseñas en archivos
```

---

## 🔄 Pipeline CI/CD

### Archivo: `.github/workflows/deploy-prod.yml`

- **Trigger:** Manual (workflow_dispatch)
- **Requiere:** Escribir "DEPLOY" para confirmar
- **Imagen Docker:** `abisaid/lbtechapi:prod-latest`

### Secretos necesarios en GitHub:
```
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
DOCKERHUB_USERNAME
DOCKERHUB_TOKEN
```

---

## 💰 Costos Estimados

| Componente | Costo/mes |
|------------|-----------|
| EC2 t3.micro | ~$8 |
| Load Balancer | ~$18 |
| MongoDB Atlas M0 | $0 |
| Dominio (anual/12) | ~$1.25 |
| **Total PROD** | **~$27/mes** |

---

## 🛑 Comando para Terminar PROD (Pausar)

```bash
aws elasticbeanstalk terminate-environment --environment-name lbtechapi-prod
```

> ⚠️ Esto elimina el ambiente pero mantiene: dominio, SSL, MongoDB, código, pipeline.

---

## ✅ Verificación post-recreación

```bash
# Verificar estado
aws elasticbeanstalk describe-environments --environment-names lbtechapi-prod

# Probar health
curl https://api.lbonetrack.com/health

# Probar login
curl -X POST https://api.lbonetrack.com/api/v1/sign-in \
  -H "Content-Type: application/json" \
  -d '{"email":"abisaidfarias@gmail.com","password":"@Mipassword123"}'
```

---

## 📞 Notas Adicionales

- El ambiente DEV sigue funcionando en: `http://api.testkoi.com`
- MongoDB DEV: `lbtechdev` (cluster separado)
- Si hay problemas con el login, verificar que el usuario y sus referencias (profile, company) existan en MongoDB PROD

---

*Última actualización: Diciembre 2024*

