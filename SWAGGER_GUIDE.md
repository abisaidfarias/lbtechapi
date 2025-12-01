# 📚 Guía de Swagger - LBTech API

## ✅ Swagger Configurado Exitosamente

Swagger ha sido instalado y configurado en tu proyecto. La documentación interactiva de tu API está lista para usar.

## 🚀 Acceder a Swagger UI

Una vez que inicies el servidor, puedes acceder a la interfaz de Swagger en:

```
http://localhost:8080/swagger/index.html
```

## 📖 Documentación Actual

La documentación base está configurada con la siguiente información:

- **Título**: LBTech API
- **Versión**: 1.0
- **Descripción**: API para gestión de homologaciones y dispositivos
- **Base Path**: /api/v1
- **Puerto**: 8080
- **Autenticación**: Bearer Token (JWT)

## 🔄 Cómo Actualizar la Documentación

### 1. Agregar anotaciones en tus controladores

Para documentar endpoints específicos, agrega anotaciones en tus archivos de controladores. Ejemplo:

```go
// GetAll godoc
// @Summary Lista todos los usuarios
// @Description Obtiene la lista completa de usuarios del sistema
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} responses.UserResponse
// @Failure 401 {object} models.CustomError
// @Failure 500 {object} models.CustomError
// @Router /users [get]
func (c *userController) Get() gin.HandlerFunc {
    // ... tu código
}
```

### 2. Regenerar la documentación

Cada vez que agregues o modifiques anotaciones, ejecuta:

```bash
swag init
```

Esto actualizará los archivos en la carpeta `/docs`.

### 3. Recompilar (si es necesario)

```bash
go build -o lbtechapi
```

## 📝 Ejemplos de Anotaciones Comunes

### GET Endpoint
```go
// @Summary Descripción corta
// @Description Descripción detallada
// @Tags nombre-tag
// @Accept json
// @Produce json
// @Param id path string true "ID del recurso"
// @Success 200 {object} responses.TuRespuesta
// @Failure 404 {object} models.CustomError
// @Router /ruta/{id} [get]
```

### POST Endpoint
```go
// @Summary Crear nuevo recurso
// @Description Crea un nuevo recurso en el sistema
// @Tags nombre-tag
// @Accept json
// @Produce json
// @Param request body request.TuRequest true "Datos del recurso"
// @Success 201 {object} responses.TuRespuesta
// @Failure 400 {object} models.CustomError
// @Router /ruta [post]
```

### PUT Endpoint
```go
// @Summary Actualizar recurso
// @Description Actualiza un recurso existente
// @Tags nombre-tag
// @Accept json
// @Produce json
// @Param id path string true "ID del recurso"
// @Param request body request.TuRequest true "Datos actualizados"
// @Success 200 {object} responses.TuRespuesta
// @Failure 404 {object} models.CustomError
// @Router /ruta/{id} [put]
```

### DELETE Endpoint
```go
// @Summary Eliminar recurso
// @Description Elimina un recurso del sistema
// @Tags nombre-tag
// @Param id path string true "ID del recurso"
// @Success 204
// @Failure 404 {object} models.CustomError
// @Router /ruta/{id} [delete]
```

## 🔐 Autenticación en Swagger UI

Para probar endpoints protegidos:

1. Haz login usando el endpoint `/api/v1/sign-in`
2. Copia el token JWT de la respuesta
3. Haz clic en el botón "Authorize" en Swagger UI
4. Ingresa: `Bearer TU_TOKEN_AQUI`
5. Haz clic en "Authorize"

Ahora podrás probar los endpoints protegidos.

## 📁 Archivos Generados

Swagger genera automáticamente estos archivos en `/docs`:

- `docs.go` - Documentación en formato Go
- `swagger.json` - Especificación OpenAPI en JSON
- `swagger.yaml` - Especificación OpenAPI en YAML

⚠️ **No edites estos archivos manualmente**, se regeneran con `swag init`.

## 🛠️ Comandos Útiles

```bash
# Instalar swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generar/actualizar documentación
swag init

# Verificar formato de anotaciones
swag fmt

# Ejecutar el servidor
./lbtechapi
# o
go run main.go
```

## 📚 Recursos Adicionales

- [Documentación de swag](https://github.com/swaggo/swag)
- [Ejemplos de anotaciones](https://github.com/swaggo/swag#declarative-comments-format)
- [OpenAPI Specification](https://swagger.io/specification/)

## 🎯 Próximos Pasos

1. Agrega anotaciones a tus controladores
2. Ejecuta `swag init` para regenerar la documentación
3. Inicia el servidor y visita `http://localhost:8080/swagger/index.html`
4. ¡Disfruta de tu API documentada!

