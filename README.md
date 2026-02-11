# Cheos Cafe - Backend API

Backend API REST para el sistema de e-commerce de Cheos Cafe, desarrollado en Go con Firebase Firestore.

---

## Descripcion del Proyecto

Este repositorio contiene el backend completo para la plataforma de e-commerce de Cheos Cafe, una empresa colombiana con 8 tiendas fisicas en Antioquia que vende cafe molido de especialidad.

El backend proporciona una API REST que permite la gestion completa de un e-commerce de cafe, incluyendo:

- Autenticacion de usuarios con JWT (roles ADMIN/CUSTOMER)
- Gestion de catalogo de productos
- Sistema de ordenes con checkout para usuarios autenticados e invitados
- Codigos de descuento con validacion y limites de uso
- Sistema de resenas con moderacion
- Gestion de ubicaciones de tiendas fisicas
- Galeria de imagenes con subida a Cloudinary
- Carrusel editable por admin (configuracion del sitio)
- Rate limiting independiente para login y global
- CORS configurado para produccion

El sistema esta disenado con una arquitectura de tres capas (Repository, Service, Handler) y utiliza Firebase Firestore como base de datos NoSQL.

**Estado actual:** 46 endpoints implementados y probados. Backend desplegado en Render, frontend en Netlify.

---

## Tecnologias

### Stack Principal

<p align="center">
  <img src="internal/decoration/goland.png" alt="Go" height="120" style="margin: 0 30px;"/>
  <img src="internal/decoration/gogin.png" alt="Gin" height="120" style="margin: 0 30px;"/>
  <img src="internal/decoration/firebase.png" alt="Firebase" height="120" style="margin: 0 30px;"/>
</p>

### Dependencias Principales

```
github.com/gin-gonic/gin              # Framework HTTP
github.com/golang-jwt/jwt/v5          # Autenticacion JWT
golang.org/x/crypto/bcrypt            # Hash de contrasenas
github.com/google/uuid                # Generacion de UUIDs
github.com/go-playground/validator    # Validacion de requests
github.com/sirupsen/logrus            # Sistema de logs
cloud.google.com/go/firestore         # SDK de Firestore
github.com/redis/go-redis/v9          # Cliente Redis (opcional)
github.com/cloudinary/cloudinary-go   # SDK de Cloudinary
```

---

## Guia de Instalacion

### Requisitos Previos

- Go 1.21 o superior
- Git
- Cuenta de Firebase con Firestore habilitado
- Cuenta de Cloudinary (para subida de imagenes)

### 1. Clonar el Repositorio

```bash
git clone <repository-url>
cd GoBackend_Cheos
```

### 2. Instalar Dependencias

```bash
go mod download
go mod tidy
```

### 3. Configurar Firebase

1. Ir a Firebase Console (https://console.firebase.google.com/)
2. Seleccionar tu proyecto o crear uno nuevo
3. Habilitar Firestore Database
4. Ir a Project Settings > Service Accounts
5. Click en Generate New Private Key
6. Guardar el archivo JSON como `firebase-credentials.json` en la raiz del proyecto
7. Copiar el Project ID

### 4. Configurar Variables de Entorno

Copiar el archivo de ejemplo:

```bash
cp .env.example .env
```

Editar `.env` con las siguientes variables:

```env
# General
GO_ENV=development
PORT=8080

# Firebase (OBLIGATORIO - opcion 1: archivo local)
FIREBASE_PROJECT_ID=tu-proyecto-id
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json

# Firebase (opcion 2: variable de entorno, para produccion/Render)
# FIREBASE_CREDENTIALS_JSON=<JSON completo del archivo de credenciales>

# JWT (OBLIGATORIO)
JWT_SECRET=tu-secreto-jwt-seguro
JWT_REFRESH_SECRET=tu-secreto-refresh-seguro
JWT_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=168h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=15m

# Cloudinary (para subida de imagenes)
CLOUDINARY_CLOUD_NAME=tu-cloud-name
CLOUDINARY_API_KEY=tu-api-key
CLOUDINARY_API_SECRET=tu-api-secret

# Redis (opcional)
# REDIS_HOST=localhost:6379
```

Generar JWT Secrets:

```bash
# Linux/Mac
openssl rand -base64 32

# Windows PowerShell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})
```

Ejecutar dos veces para generar `JWT_SECRET` y `JWT_REFRESH_SECRET`.

### 5. Iniciar el Backend

Modo desarrollo:
```bash
go run cmd/api/main.go
```

Compilar y ejecutar:
```bash
go build -o backend cmd/api/main.go
./backend
```

### 6. Verificar Funcionamiento

```bash
curl http://localhost:8080/health
```

Respuesta esperada:
```json
{
  "status": "ok",
  "message": "Cheos Cafe Backend API is running",
  "version": "v1"
}
```

---

## Deployment

### Backend en Render (Docker)

El proyecto incluye un Dockerfile multi-stage optimizado para Render.

Variables de entorno en Render:
```
GO_ENV=production
PORT=8080
FIREBASE_PROJECT_ID=<project-id>
FIREBASE_CREDENTIALS_JSON=<JSON completo del archivo de credenciales>
JWT_SECRET=<secret>
JWT_REFRESH_SECRET=<secret>
JWT_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=168h
CORS_ALLOWED_ORIGINS=https://cheoscafesena.netlify.app
CLOUDINARY_CLOUD_NAME=<cloud-name>
CLOUDINARY_API_KEY=<api-key>
CLOUDINARY_API_SECRET=<api-secret>
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=15m
```

Nota: En Render se usa `FIREBASE_CREDENTIALS_JSON` (JSON completo como string) en lugar del archivo local.

### Frontend en Netlify

Variable de entorno en Netlify:
```
VITE_API_URL=https://gobackend-cheos.onrender.com/api/v1
```

---

## Manual de Usuario

### Tipos de Usuarios

- **CUSTOMER:** Usuario regular. Puede ver productos, crear ordenes y dejar resenas.
- **ADMIN:** Administrador. Acceso completo para gestionar productos, ordenes, resenas, codigos de descuento, ubicaciones, galeria y configuracion del sitio.

### Autenticacion

Endpoints protegidos requieren un token JWT en el header:

```
Authorization: Bearer <access_token>
```

Los access tokens expiran en 15 minutos. Los refresh tokens expiran en 7 dias. Usar `/api/v1/auth/refresh` para renovar el token.

### Flujo de Compra (Cliente)

1. Ver productos: `GET /api/v1/products`
2. Buscar productos: `GET /api/v1/products/search?q=premium`
3. Ver detalle: `GET /api/v1/products/{id}`
4. Crear orden: `POST /api/v1/orders` (sin autenticacion o con ella)
5. Rastrear orden: `GET /api/v1/orders/number/{order_number}`

### Flujo de Gestion (Administrador)

1. Login: `POST /api/v1/auth/login`
2. Crear producto: `POST /api/v1/products`
3. Actualizar stock: `PATCH /api/v1/products/{id}/stock`
4. Ver ordenes: `GET /api/v1/orders`
5. Actualizar estado de orden: `PATCH /api/v1/orders/{id}/status`
6. Moderar resenas: `PUT /api/v1/reviews/{id}`
7. Crear codigos de descuento: `POST /api/v1/discounts`
8. Gestionar galeria: `POST /api/v1/gallery/upload`
9. Editar carrusel: `PUT /api/v1/config/carousel`

### Estados de Orden

- **PENDING:** Orden creada, esperando confirmacion
- **CONFIRMED:** Pago confirmado
- **PROCESSING:** Orden en preparacion
- **SHIPPED:** Orden enviada
- **DELIVERED:** Orden entregada
- **CANCELLED:** Orden cancelada

### Metodos de Pago

- **CONTRA_ENTREGA:** Pago al recibir (activo)
- **TARJETA:** Tarjeta de credito/debito (preparado)
- **PSE:** Transferencia bancaria (preparado)

---

## Endpoints Disponibles (46 totales)

### Resumen por Servicio

| Servicio | Total | Publicos | Autenticados | Admin |
|----------|-------|----------|--------------|-------|
| Health Checks | 3 | 3 | 0 | 0 |
| Auth | 5 | 5 | 0 | 0 |
| Users | 5 | 0 | 2 | 3 |
| Products | 8 | 4 | 0 | 4 |
| Orders | 7 | 2 | 2 | 3 |
| Discounts | 6 | 1 | 0 | 5 |
| Reviews | 6 | 2 | 0 | 4 |
| Locations | 6 | 3 | 0 | 3 |
| Gallery | 8 | 3 | 0 | 5 |
| Site Config | 2 | 1 | 0 | 1 |
| **TOTAL** | **56** | **24** | **4** | **28** |

### Health Checks

```
GET  /health
GET  /health/firebase
GET  /health/redis
```

### Autenticacion

```
POST /api/v1/auth/register
POST /api/v1/auth/login              (rate limited: 5 intentos / 15 min)
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/ping
```

### Usuarios

```
GET    /api/v1/users/me              (User)
PUT    /api/v1/users/me              (User)
GET    /api/v1/users                 (Admin)
PUT    /api/v1/users/:id             (Admin)
DELETE /api/v1/users/:id             (Admin)
```

### Productos

```
GET    /api/v1/products
GET    /api/v1/products/featured
GET    /api/v1/products/search?q=
GET    /api/v1/products/:id
POST   /api/v1/products              (Admin)
PUT    /api/v1/products/:id          (Admin)
PATCH  /api/v1/products/:id/stock    (Admin)
DELETE /api/v1/products/:id          (Admin)
```

### Ordenes

```
POST  /api/v1/orders
GET   /api/v1/orders/number/:number
GET   /api/v1/orders/me              (User)
GET   /api/v1/orders/:id             (User)
GET   /api/v1/orders                 (Admin)
PATCH /api/v1/orders/:id/status      (Admin)
PATCH /api/v1/orders/:id/payment     (Admin)
```

### Codigos de Descuento

```
POST   /api/v1/discounts/validate
GET    /api/v1/discounts             (Admin)
POST   /api/v1/discounts             (Admin)
GET    /api/v1/discounts/:id         (Admin)
PUT    /api/v1/discounts/:id         (Admin)
DELETE /api/v1/discounts/:id         (Admin)
```

### Resenas

```
POST   /api/v1/reviews
GET    /api/v1/products/:id/reviews
GET    /api/v1/reviews               (Admin)
GET    /api/v1/reviews/:id           (Admin)
PUT    /api/v1/reviews/:id           (Admin)
DELETE /api/v1/reviews/:id           (Admin)
```

### Ubicaciones

```
GET    /api/v1/locations
GET    /api/v1/locations/all
GET    /api/v1/locations/:id
POST   /api/v1/locations             (Admin)
PUT    /api/v1/locations/:id         (Admin)
DELETE /api/v1/locations/:id         (Admin)
```

### Galeria

```
GET    /api/v1/gallery/active
GET    /api/v1/gallery/type/:type
GET    /api/v1/gallery/:id
GET    /api/v1/gallery               (Admin)
POST   /api/v1/gallery               (Admin)
POST   /api/v1/gallery/upload        (Admin)
PUT    /api/v1/gallery/:id           (Admin)
DELETE /api/v1/gallery/:id           (Admin)
```

### Configuracion del Sitio

```
GET  /api/v1/config/carousel         (publico)
PUT  /api/v1/config/carousel         (Admin) - body: {"images": ["url1", "url2", ...]}
```

Nota: Documentacion detallada y coleccion Postman en `Cheos_Cafe_API.postman_collection.json`.

---

## Estructura del Proyecto

```
GoBackend_Cheos/
├── cmd/
│   └── api/
│       └── main.go                    # Punto de entrada + rutas
│
├── internal/
│   ├── config/                        # Configuracion
│   │   └── config.go
│   ├── database/                      # Conexiones (Firebase, Redis)
│   │   ├── firebase.go
│   │   └── redis.go
│   ├── handlers/                      # Controladores HTTP (8 handlers)
│   │   ├── auth_handler.go
│   │   ├── product_handler.go
│   │   ├── order_handler.go
│   │   ├── discount_handler.go
│   │   ├── review_handler.go
│   │   ├── location_handler.go
│   │   ├── gallery_handler.go
│   │   └── site_config_handler.go
│   ├── middleware/                     # Middlewares
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── rate_limit.go
│   │   └── admin.go
│   ├── models/                        # Modelos de datos (8 modelos + DTOs)
│   │   ├── user.go
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── location.go
│   │   ├── gallery.go
│   │   ├── review.go
│   │   ├── discount.go
│   │   └── site_config.go
│   ├── repository/                    # Capa de acceso a datos (8 repositorios)
│   │   ├── user_repository.go
│   │   ├── product_repository.go
│   │   ├── order_repository.go
│   │   ├── location_repository.go
│   │   ├── gallery_repository.go
│   │   ├── review_repository.go
│   │   ├── discount_repository.go
│   │   └── site_config_repository.go
│   ├── services/                      # Logica de negocio (8 servicios + upload)
│   │   ├── auth_service.go
│   │   ├── product_service.go
│   │   ├── order_service.go
│   │   ├── location_service.go
│   │   ├── gallery_service.go
│   │   ├── review_service.go
│   │   ├── discount_service.go
│   │   ├── site_config_service.go
│   │   └── upload_service.go
│   └── utils/                         # Utilidades
│       ├── jwt.go
│       ├── password.go
│       ├── response.go
│       ├── validator.go
│       └── order_number.go
│
├── .env.example                       # Template variables de entorno
├── .dockerignore                      # Exclusiones para Docker build
├── .gitignore                         # Reglas Git
├── Dockerfile                         # Multi-stage build para Render
├── Cheos_Cafe_API.postman_collection.json
├── go.mod
├── go.sum
├── ContextoSesion.md
├── MigracionClaude.md
└── README.md
```

---

## Comandos Utiles

```bash
# Desarrollo
go run cmd/api/main.go              # Ejecutar servidor
go mod download                     # Descargar dependencias
go mod tidy                         # Limpiar dependencias

# Build
go build -o backend cmd/api/main.go        # Linux/Mac
go build -o backend.exe cmd/api/main.go    # Windows

# Docker
docker build -t cheos-backend .
docker run -p 8080:8080 cheos-backend

# Code Quality
go fmt ./...                        # Formatear codigo
go vet ./...                        # Analizar codigo
```

- Fin... Gracias por leerlo todo <3
