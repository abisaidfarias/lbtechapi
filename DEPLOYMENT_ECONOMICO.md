# 💰 Deployment Económico - Sin Aumentar Costos

## 🎯 Objetivo

Actualizar tu aplicación **SIN activar servicios costosos** ni aumentar tu factura de AWS.

---

## 📋 PRIMERO: Verifica Qué Está Corriendo

```bash
# Instalar AWS CLI (si no lo tienes)
brew install awscli

# Configurar credenciales
aws configure
# Región: us-east-1
# Las demás credenciales las obtienes de AWS Console → IAM

# Ejecutar script de verificación (SOLO LECTURA)
chmod +x check-aws-current-state.sh
./check-aws-current-state.sh
```

Este script te dirá:
- ✅ Qué servicios están corriendo (y costando dinero)
- 💰 Estimación de costos mensuales
- 📦 Qué está almacenado en ECR
- 🔧 Si hay pipelines configurados

---

## 🚀 Opciones de Deployment (de más barata a más cara)

### Opción 1: Solo Actualizar Imagen en ECR (Casi Gratis) 🟢

**Costo**: ~$0.10/mes (solo almacenamiento)  
**Ideal si**: Ya tienes un ambiente corriendo que usa la imagen de ECR

```bash
# 1. Build local
docker build -t lbtechapidev:latest .

# 2. Login a ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  438758934896.dkr.ecr.us-east-1.amazonaws.com

# 3. Tag
docker tag lbtechapidev:latest \
  438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest

# 4. Push
docker push 438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest
```

Luego:
- Si usas ECS/Fargate: Reinicia el servicio desde AWS Console
- Si usas Elastic Beanstalk: Redeploy desde AWS Console
- Si usas EC2: SSH y haz `docker pull` + restart

---

### Opción 2: GitHub + DockerHub (Gratis) 🟢

**Costo**: $0  
**Ideal si**: Quieres evitar AWS por completo para el deployment

```bash
# 1. Crear cuenta en Docker Hub (gratis)
# hub.docker.com

# 2. Login a Docker Hub
docker login

# 3. Tag con tu usuario de Docker Hub
docker tag lbtechapidev:latest tu-usuario/lbtechapi:latest

# 4. Push a Docker Hub
docker push tu-usuario/lbtechapi:latest
```

Actualizar `Dockerrun.aws.json`:
```json
{
    "AWSEBDockerrunVersion":"1",
    "Image":{
        "Name":"tu-usuario/lbtechapi:latest"
    },
    "Ports":[
        {
            "ContainerPort":"8080"
        }
    ]
}
```

---

### Opción 3: Elastic Beanstalk Single Instance 🟡

**Costo**: ~$10-15/mes (gratis primer año con Free Tier)  
**Ideal si**: Ya tienes un environment de EB corriendo

**NO CREAR NUEVO**, solo actualizar el existente:

```bash
# 1. Instalar EB CLI
pip install awsebcli

# 2. Inicializar (usa el environment existente)
eb init

# 3. Verificar status (NO CREA NADA)
eb status

# 4. Deploy (solo actualiza)
eb deploy
```

---

### Opción 4: Mantener Todo Como Está ⚪

**Costo**: El que ya tienes  
**Si no quieres cambiar nada**:

```bash
# Solo verifica qué está corriendo
./check-aws-current-state.sh

# Ve a AWS Console y verifica manualmente
# https://console.aws.amazon.com
```

---

## ❌ EVITA ESTO (Aumentará Costos)

### NO hagas:

1. **NO crear nuevos recursos**
   - No crees nuevas instancias EC2
   - No crees nuevos environments de Elastic Beanstalk
   - No crees clusters nuevos de ECS

2. **NO uses servicios premium**
   - No uses RDS (usa MongoDB Atlas Free Tier)
   - No uses Load Balancers
   - No uses NAT Gateways
   - No uses instancias grandes (max t2.micro)

3. **NO dejes CodePipeline activado** si no lo necesitas
   - Cuesta $1/mes por pipeline
   - Solo actívalo cuando lo necesites

4. **NO acumules imágenes en ECR**
   - Limpia imágenes viejas regularmente
   - Solo mantén las últimas 3-5 versiones

---

## 💡 Estrategia Recomendada para Dev

### Setup Económico Ideal:

```
┌─────────────────────────────────────┐
│  GRATIS / CASI GRATIS               │
├─────────────────────────────────────┤
│ • GitHub (gratis)                   │
│ • MongoDB Atlas (Free Tier 512MB)   │
│ • AWS ECR (storage mínimo)          │
│ • Elastic Beanstalk t2.micro        │
│   (Free Tier 1er año)               │
│                                      │
│ TOTAL: $0-10/mes                    │
└─────────────────────────────────────┘
```

---

## 🔍 Verificar Costos Actuales

### Ver tu factura AWS:

1. Ir a: https://console.aws.amazon.com/billing/
2. Click en "Bills" en el menú izquierdo
3. Ver el mes actual
4. Ver desglose por servicio

### Configurar alertas de costo:

```bash
# Crear alarma de billing (te avisa si gastas más de $10)
aws cloudwatch put-metric-alarm \
  --alarm-name billing-alarm \
  --alarm-description "Alert if monthly bill exceeds $10" \
  --metric-name EstimatedCharges \
  --namespace AWS/Billing \
  --statistic Maximum \
  --period 21600 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 1
```

---

## 📊 Plan de Acción Sugerido

### Paso 1: Evaluar (Sin Costo)
```bash
# Ver qué tienes actualmente
./check-aws-current-state.sh

# Ver costos en AWS Console
# https://console.aws.amazon.com/billing/
```

### Paso 2: Decidir

**Si gastas $0-5/mes**: 
- ✅ Solo actualiza imagen en ECR (Opción 1)

**Si gastas $10-20/mes**:
- ✅ Usa Elastic Beanstalk existente (Opción 3)
- 🔍 Verifica si puedes optimizar

**Si gastas >$20/mes**:
- ⚠️ Revisa qué está corriendo
- 🛑 Considera apagar servicios no usados
- 💡 Migra a servicios más baratos

### Paso 3: Actualizar (Bajo Costo)

```bash
# Build y push a ECR
docker build -t lbtechapidev:latest .
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  438758934896.dkr.ecr.us-east-1.amazonaws.com
docker tag lbtechapidev:latest \
  438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest
docker push 438758934896.dkr.ecr.us-east-1.amazonaws.com/lbtechapidev:latest

# Luego actualiza tu servicio desde AWS Console
```

---

## 🛑 Apagar Servicios No Usados (Ahorrar Dinero)

### Si encuentras servicios corriendo que no usas:

```bash
# Detener instancia EC2 (no la elimina)
aws ec2 stop-instances --instance-ids i-xxxxx --region us-east-1

# Terminar environment de Elastic Beanstalk
eb terminate environment-name

# Detener servicio ECS
aws ecs update-service \
  --cluster cluster-name \
  --service service-name \
  --desired-count 0 \
  --region us-east-1
```

⚠️ **CUIDADO**: Asegúrate de que no los necesitas antes de detenerlos

---

## 📝 Checklist Antes de Deployment

- [ ] Ejecuté `./check-aws-current-state.sh`
- [ ] Revisé mi factura actual en AWS Console
- [ ] Sé qué servicios están corriendo
- [ ] Elegí la opción más económica para mi caso
- [ ] Tengo backup de datos importantes
- [ ] No voy a crear recursos nuevos innecesarios

---

## 🆘 Si Algo Sale Mal

### "Mi factura subió mucho"

1. Ir a AWS Console → Billing → Bills
2. Ver qué servicio está costando
3. Detenerlo inmediatamente
4. Contactar soporte AWS si es necesario

### "Borré algo por error"

- ECR: Las imágenes se pueden re-pushear
- EB Environment: Se puede recrear
- EC2: Si tenías snapshot, se puede restaurar

### "No sé qué está corriendo"

```bash
# Ejecuta esto y mándame el output
./check-aws-current-state.sh > aws-status.txt
cat aws-status.txt
```

---

## 💬 Resumen

**Para deployment sin aumentar costos**:

1. ✅ Usa `./check-aws-current-state.sh` primero
2. ✅ Solo actualiza imagen en ECR
3. ✅ Reutiliza recursos existentes
4. ❌ NO crees recursos nuevos
5. 💰 Monitorea tu billing mensualmente

**Costo objetivo para dev**: $0-10/mes

---

**Creado**: 2025-12-01  
**Actualizado**: Para ambiente de desarrollo económico

