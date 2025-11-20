# 🔄 Migración de Proyecto - Cheos Café Backend

## 📋 Contexto General del Proyecto

Este es un **backend completo para e-commerce de Cheos Café**, desarrollado en **Go 1.21+** usando **Gin Framework** y **Firebase Firestore** como base de datos NoSQL. El proyecto está en fase de desarrollo activo y ya cuenta con funcionalidad completa de autenticación, productos, órdenes, ubicaciones, galería, reseñas y códigos de descuento.

### Información Importante del Proyecto:
- **Nombre:** Cheos Café Backend
- **Lenguaje:** Go (Golang)
- **Framework Web:** Gin
- **Base de Datos:** Firebase Firestore (NoSQL)
- **Autenticación:** JWT con cookies (24 horas de expiración)
- **Patrón de Arquitectura:** Repository-Service-Handler (3 capas)
- **Puerto:** 8080
- **Versión API:** v1
- **Rama Git:** main

---

## 🎯 Estado Actual del Desarrollo

### ✅ Completado Recientemente:

#### 1. **Endpoints de Gestión de Usuarios (Administrador)**
Se implementaron 3 nuevos endpoints para administradores:
- `GET /api/v1/users` - Obtener todos los usuarios
- `PUT /api/v1/users/:id` - Actualizar cualquier usuario por ID
- `DELETE /api/v1/users/:id` - Eliminar usuario por ID

**Archivos modificados:**
- `internal/repository/user_repository.go` - Métodos: `GetAll()`, `UpdateByID()`, `Delete()`
- `internal/services/auth_service.go` - Métodos: `GetAllUsers()`, `UpdateUserByID()`, `DeleteUser()`
- `internal/models/user.go` - Nuevo DTO: `UpdateUserByIDRequest`
- `internal/handlers/auth_handler.go` - Handlers: `GetAllUsers()`, `UpdateUserByID()`, `DeleteUser()`
- `cmd/api/main.go` - 3 nuevas rutas admin-only

**Características:**
- Validación de email único al actualizar
- Hashing de contraseña con bcrypt
- Actualización parcial usando punteros
- Solo accesible para usuarios con rol ADMIN

#### 2. **Campo map_iframe para Google Maps en Ubicaciones**
Se agregó la capacidad de almacenar iframes de Google Maps en las ubicaciones:

**Archivos modificados:**
- `internal/models/location.go` - Campo `MapIframe string`
- `internal/services/location_service.go` - Soporte en Create y Update
- `internal/repository/location_repository.go` - Agregado al método `Update()`

**Uso:**
Los administradores pueden pegar el código iframe de Google Maps para mostrar la ubicación exacta.

#### 3. **Cambio CRÍTICO: TODOS los DELETE son ahora Hard Delete**
Se cambió **TODOS** los endpoints DELETE de Soft Delete a Hard Delete (eliminación física):

**Repositorios modificados:**
- ✅ `product_repository.go` - DELETE elimina físicamente
- ✅ `user_repository.go` - DELETE elimina físicamente
- ✅ `location_repository.go` - DELETE elimina físicamente
- ✅ `gallery_repository.go` - DELETE elimina físicamente
- ✅ `discount_repository.go` - DELETE elimina físicamente
- ✅ `review_repository.go` - Ya tenía Hard Delete

**IMPORTANTE:** Ahora cuando eliminas cualquier registro, se **borra completamente de Firebase** y no se puede recuperar.

---

## 📂 Estructura del Proyecto

```
GoBackend_Cheos/
├── cmd/
│   └── api/
│       └── main.go                 # Punto de entrada, configuración de rutas
├── internal/
│   ├── config/
│   │   └── config.go               # Configuración (JWT, Firebase, etc.)
│   ├── database/
│   │   ├── firebase.go             # Conexión a Firestore
│   │   └── redis.go                # Conexión a Redis (cache)
│   ├── handlers/                   # HTTP handlers (controladores)
│   │   ├── auth_handler.go         # Login, registro, gestión usuarios
│   │   ├── product_handler.go      # CRUD productos
│   │   ├── order_handler.go        # Gestión de órdenes
│   │   ├── location_handler.go     # CRUD ubicaciones
│   │   ├── gallery_handler.go      # Galería de imágenes
│   │   ├── review_handler.go       # Reseñas de productos
│   │   └── discount_handler.go     # Códigos de descuento
│   ├── middleware/
│   │   ├── auth.go                 # AuthMiddleware (JWT)
│   │   ├── cors.go                 # CORS
│   │   └── admin.go                # RequireAdmin
│   ├── models/                     # Modelos de datos y DTOs
│   │   ├── user.go
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── location.go
│   │   ├── gallery.go
│   │   ├── review.go
│   │   └── discount.go
│   ├── repository/                 # Capa de acceso a datos
│   │   ├── user_repository.go
│   │   ├── product_repository.go
│   │   ├── order_repository.go
│   │   ├── location_repository.go
│   │   ├── gallery_repository.go
│   │   ├── review_repository.go
│   │   └── discount_repository.go
│   ├── services/                   # Lógica de negocio
│   │   ├── auth_service.go
│   │   ├── product_service.go
│   │   ├── order_service.go
│   │   ├── location_service.go
│   │   ├── gallery_service.go
│   │   ├── review_service.go
│   │   └── discount_service.go
│   └── utils/
│       ├── jwt.go                  # Generación y validación JWT
│       ├── password.go             # Hashing bcrypt
│       ├── response.go             # Respuestas HTTP estandarizadas
│       └── validator.go            # Validación de structs
├── .env                            # Variables de entorno
├── go.mod
├── go.sum
├── CambiosAHacer.md               # Lista de tareas pendientes
└── MigracionClaude.md             # Este archivo
```

---

## 🔧 Tecnologías y Dependencias

### Stack Tecnológico:
- **Go 1.21+**
- **Gin Web Framework** - HTTP routing y middleware
- **Firebase Admin SDK** - Firestore database
- **JWT (golang-jwt/jwt)** - Autenticación
- **bcrypt** - Hashing de contraseñas
- **UUID** - Identificadores únicos
- **Redis (opcional)** - Cache (actualmente no conectado)

### Variables de Entorno (.env):
```env
PORT=8080
JWT_SECRET=tu-secret-key
FIREBASE_CREDENTIALS_PATH=path/to/serviceAccountKey.json
REDIS_HOST=localhost:6379
```

---

## 🚀 API Endpoints Disponibles

### 🔐 Autenticación (`/api/v1/auth`)
| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| POST | `/register` | Registrar nuevo usuario | Público |
| POST | `/login` | Iniciar sesión | Público |
| POST | `/refresh` | Refrescar token | Público |
| POST | `/logout` | Cerrar sesión | Público |

### 👤 Usuarios (`/api/v1/users`)
| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| GET | `/me` | Obtener perfil propio | Usuario |
| PUT | `/me` | Actualizar perfil propio | Usuario |
| **GET** | **`/`** | **Obtener todos los usuarios** | **Admin** |
| **PUT** | **`/:id`** | **Actualizar usuario por ID** | **Admin** |
| **DELETE** | **`/:id`** | **Eliminar usuario (hard delete)** | **Admin** |

### 🛍️ Productos (`/api/v1/products`)
| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| GET | `/` | Listar productos (paginado) | Público |
| GET | `/:id` | Obtener producto por ID | Público |
| GET | `/featured` | Productos destacados | Público |
| GET | `/search?q=` | Buscar productos | Público |
| POST | `/` | Crear producto | Admin |
| PUT | `/:id` | Actualizar producto | Admin |
| DELETE | `/:id` | **Eliminar producto (hard delete)** | Admin |
| PATCH | `/:id/stock` | Actualizar stock | Admin |

### 📦 Órdenes (`/api/v1/orders`)
| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| POST | `/` | Crear orden | Público |
| GET | `/number/:number` | Buscar por número | Público |
| GET | `/me` | Mis órdenes | Usuario |
| GET | `/:id` | Obtener orden por ID | Usuario |
| GET | `/` | Todas las órdenes | Admin |
| PATCH | `/:id/status` | Actualizar estado | Admin |
| PATCH | `/:id/payment` | Actualizar pago | Admin |

### 📍 Ubicaciones (`/api/v1/locations`)
| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| GET | `/` | Ubicaciones activas | Público |
| GET | `/all` | Todas las ubicaciones | Público |
| GET | `/:id` | Obtener por ID | Público |
| POST | `/` | Crear ubicación | Admin |
| PUT | `/:id` | Actualizar ubicación | Admin |
| DELETE | `/:id` | **Eliminar ubicación (hard delete)** | Admin |

### 🖼️ Galería (`/api/v1/gallery`)
| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| GET | `/active` | Imágenes activas | Público |
| GET | `/type/:type` | Por tipo | Público |
| GET | `/:id` | Obtener por ID | Público |
| GET | `/` | Todas las imágenes | Admin |
| POST | `/` | Crear imagen | Admin |
| POST | `/upload` | Subir imagen | Admin |
| PUT | `/:id` | Actualizar imagen | Admin |
| DELETE | `/:id` | **Eliminar imagen (hard delete)** | Admin |

### ⭐ Reseñas (`/api/v1/reviews`)
| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| POST | `/` | Crear reseña | Público |
| GET | `/` | Todas las reseñas | Admin |
| GET | `/:id` | Obtener por ID | Admin |
| GET | `/products/:id/reviews` | Reseñas de producto | Público |
| PUT | `/:id` | Actualizar reseña | Admin |
| DELETE | `/:id` | **Eliminar reseña (hard delete)** | Admin |

### 🎟️ Descuentos (`/api/v1/discounts`)
| Método | Endpoint | Descripción | Auth |
|--------|----------|-------------|------|
| POST | `/validate` | Validar código | Público |
| GET | `/` | Todos los códigos | Admin |
| POST | `/` | Crear código | Admin |
| GET | `/:id` | Obtener por ID | Admin |
| PUT | `/:id` | Actualizar código | Admin |
| DELETE | `/:id` | **Eliminar código (hard delete)** | Admin |

---

## 📝 Tareas Pendientes (CambiosAHacer.md)

### 2. LOCACIONES ⚠️ PENDIENTE
**Problema:** Los endpoints de locaciones siguen pidiendo token aunque deberían ser públicos.

**Endpoints que deben ser públicos:**
- `GET /api/v1/locations` ✅ (Ya está público)
- `GET /api/v1/locations/all` ✅ (Ya está público)
- `GET /api/v1/locations/:id` ✅ (Ya está público)

**Acción:** Verificar en Postman si realmente piden token o si el problema ya está resuelto.

### 3. ORDENES ⚠️ PENDIENTE
**Problema:** Las órdenes se están enviando con un correo predeterminado aunque el usuario esté logueado.

**Archivo a revisar:** `internal/handlers/order_handler.go` o `internal/services/order_service.go`

**Acción:** Cuando un usuario esté autenticado, usar su email del JWT en lugar de un email predeterminado.

---

## 🔑 Modelos de Datos Principales

### User
```go
type User struct {
    ID        uuid.UUID `json:"id" firestore:"id"`
    Email     string    `json:"email" firestore:"email"`
    Password  string    `json:"-" firestore:"password"`
    Name      string    `json:"name" firestore:"name"`
    Phone     string    `json:"phone" firestore:"phone"`
    Role      UserRole  `json:"role" firestore:"role"` // ADMIN | CUSTOMER
    IsActive  bool      `json:"is_active" firestore:"is_active"`
    CreatedAt time.Time `json:"created_at" firestore:"created_at"`
    UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`
}
```

### Product
```go
type Product struct {
    ID          uuid.UUID `json:"id" firestore:"id"`
    Name        string    `json:"name" firestore:"name"`
    Description string    `json:"description" firestore:"description"`
    Price       float64   `json:"price" firestore:"price"`
    Weight      int       `json:"weight" firestore:"weight"`
    Stock       int       `json:"stock" firestore:"stock"`
    Category    string    `json:"category" firestore:"category"`
    Images      []string  `json:"images" firestore:"images"`
    IsActive    bool      `json:"is_active" firestore:"is_active"`
    IsFeatured  bool      `json:"is_featured" firestore:"is_featured"`
    CreatedAt   time.Time `json:"created_at" firestore:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" firestore:"updated_at"`
}
```

### Location
```go
type Location struct {
    ID         uuid.UUID `json:"id" firestore:"id"`
    Name       string    `json:"name" firestore:"name"`
    Address    string    `json:"address" firestore:"address"`
    City       string    `json:"city" firestore:"city"`
    Department string    `json:"department" firestore:"department"`
    Phone      string    `json:"phone" firestore:"phone"`
    Latitude   float64   `json:"latitude" firestore:"latitude"`
    Longitude  float64   `json:"longitude" firestore:"longitude"`
    MapIframe  string    `json:"map_iframe" firestore:"map_iframe"` // NUEVO
    Schedule   *Schedule `json:"schedule" firestore:"schedule"`
    IsActive   bool      `json:"is_active" firestore:"is_active"`
    CreatedAt  time.Time `json:"created_at" firestore:"created_at"`
    UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at"`
}
```

---

## 🧪 Cómo Probar el Proyecto

### 1. Iniciar el servidor:
```bash
go run cmd/api/main.go
```

El servidor iniciará en `http://localhost:8080`

### 2. Endpoints de prueba:
```bash
# Health check
curl http://localhost:8080/health

# Listar productos
curl http://localhost:8080/api/v1/products

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@cheoscafe.com","password":"Admin123456"}'
```

### 3. Usuarios de prueba en la base de datos:
- **Admin:** `admin@cheoscafe.com` / `Admin123456`
- **Admin 2:** `admin2024@cheoscafe.com` / `AdminPass123`
- **Customer:** `customer@cheoscafe.com` / (contraseña desconocida)

---

## ⚠️ Decisiones de Diseño Importantes

### 1. Hard Delete en todos los endpoints
**DECISIÓN RECIENTE:** Todos los DELETE ahora eliminan físicamente de Firebase.

**Motivo:** Solicitud del cliente para que los registros no queden en la base de datos.

**Impacto:** Los registros eliminados NO se pueden recuperar.

### 2. Autenticación con cookies + headers
El sistema acepta JWT tanto en:
- Cookie `access_token`
- Header `Authorization: Bearer <token>`

### 3. Roles de usuario
Solo hay 2 roles:
- `CUSTOMER` - Usuario normal
- `ADMIN` - Acceso total

### 4. Firestore como única fuente de verdad
No hay SQL, todo está en Firestore (NoSQL).

### 5. Validación de emails únicos
El sistema valida que no haya emails duplicados al registrar o actualizar usuarios.

---

## 🐛 Problemas Conocidos

### 1. Redis no conecta
**Síntoma:** Warning al iniciar: "Failed to connect to Redis"

**Estado:** NO CRÍTICO - El sistema continúa sin cache

**Solución:** Instalar y ejecutar Redis, o ignorar el warning.

### 2. Búsqueda de productos limitada
**Problema:** Firestore no tiene búsqueda full-text.

**Solución temporal:** Búsqueda básica implementada.

**Solución futura:** Integrar Algolia o Elasticsearch.

### 3. Índices de Firestore
Algunos queries complejos requieren índices compuestos en Firestore.

**Solución:** Firebase muestra el link del índice necesario en los errores.

---

## 🔄 Comandos Git Útiles

```bash
# Ver estado actual
git status

# Ver últimos commits
git log --oneline -10

# Crear commit (NO uses commit --amend a menos que sea tu último commit)
git add .
git commit -m "Mensaje de commit"

# Push a main
git push origin main

# Pull últimos cambios
git pull origin main
```

---

## 📌 Notas Importantes para el Nuevo Claude

1. **NUNCA uses Soft Delete** - Todos los DELETE deben ser físicos (Hard Delete)

2. **Revisa CambiosAHacer.md** antes de empezar - Ahí están las tareas pendientes

3. **Usa el patrón Repository-Service-Handler** - Mantén la arquitectura en 3 capas

4. **Los endpoints públicos NO requieren auth** - Verifica que `AuthMiddleware` no esté aplicado

5. **Admin-only routes** usan dos middlewares:
   ```go
   adminUsers := users.Group("")
   adminUsers.Use(middleware.AuthMiddleware(cfg))
   adminUsers.Use(middleware.RequireAdmin())
   ```

6. **Firebase es la única base de datos** - No hay SQL

7. **JWT expira en 24 horas** - El refresh token en 7 días

8. **Usa UUIDs** en lugar de auto-increment IDs

9. **IMPORTANTE:** El servidor está en Windows, usa comandos compatibles:
   - `taskkill //F //PID <pid>` para matar procesos
   - `netstat -ano | findstr :8080` para encontrar procesos en puerto
   - `go run cmd/api/main.go` para iniciar servidor

10. **El proyecto YA está funcional** - Solo quedan tareas menores del CambiosAHacer.md

---

## 🎯 Próximos Pasos Sugeridos

1. **Verificar el problema de locaciones** (punto 2 de CambiosAHacer.md)
2. **Arreglar el email en órdenes** (punto 3 de CambiosAHacer.md)
3. **Probar todos los endpoints DELETE** para confirmar Hard Delete
4. **Documentar en Postman** los 3 nuevos endpoints de usuarios
5. **Crear pruebas unitarias** (opcional, no requerido actualmente)

---

## 📞 Información de Contacto del Proyecto

- **Cliente:** Cheos Café
- **Tipo:** E-commerce backend
- **Estado:** Desarrollo activo
- **Fecha última actualización:** 2025-11-19

---

## 🔐 Seguridad

### Implementado:
- ✅ Hashing de contraseñas con bcrypt
- ✅ JWT para autenticación
- ✅ Validación de roles (ADMIN/CUSTOMER)
- ✅ Validación de inputs con `validator`
- ✅ Validación de email único
- ✅ Cookies HttpOnly para tokens

### Por implementar:
- ⚠️ Rate limiting
- ⚠️ HTTPS en producción
- ⚠️ Sanitización de inputs HTML
- ⚠️ Logs de auditoría

---

## 🚀 Deployment

**NOTA:** Actualmente en desarrollo local. No hay deployment en producción.

**Variables de entorno necesarias para producción:**
```env
PORT=8080
JWT_SECRET=<secret-muy-seguro>
FIREBASE_CREDENTIALS_PATH=/path/to/credentials.json
REDIS_HOST=<redis-url>
GIN_MODE=release
```

---

## 📚 Recursos Útiles

- **Gin Framework:** https://gin-gonic.com/docs/
- **Firebase Admin Go SDK:** https://firebase.google.com/docs/admin/setup
- **JWT Go:** https://github.com/golang-jwt/jwt
- **UUID:** https://github.com/google/uuid

---

## ✅ Checklist de Migración

Antes de continuar trabajando, verifica:

- [ ] Go está instalado (1.21+)
- [ ] Firebase credentials están configuradas
- [ ] `.env` existe con las variables correctas
- [ ] `go mod download` para instalar dependencias
- [ ] `go run cmd/api/main.go` para iniciar servidor
- [ ] Server responde en `http://localhost:8080/health`
- [ ] Leer `CambiosAHacer.md` para ver tareas pendientes
- [ ] Tener Postman o similar para probar endpoints

---

**¡Bienvenido al proyecto! Todo está listo para continuar el desarrollo. Revisa primero el archivo `CambiosAHacer.md` para ver las tareas pendientes.** 🚀
