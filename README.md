# Cheos Café - Backend API

Backend API REST para el sistema de e-commerce de Cheos Café, desarrollado en Go con Firebase Firestore.

---

## Descripción del Proyecto

Este repositorio contiene el backend completo para la plataforma de e-commerce de Cheos Café, una empresa colombiana con 8 tiendas físicas en Antioquia que vende café molido de especialidad.

El backend proporciona una API REST que permite la gestión completa de un e-commerce de café, incluyendo:

- Autenticación de usuarios con JWT (roles ADMIN/CUSTOMER)
- Gestión de catálogo de productos
- Sistema de órdenes con checkout para usuarios autenticados e invitados
- Códigos de descuento con validación y límites de uso
- Sistema de reseñas con moderación
- Gestión de ubicaciones de tiendas físicas
- Rate limiting y CORS configurado

El sistema está diseñado con una arquitectura de tres capas (Repository, Service, Handler) y utiliza Firebase Firestore como base de datos NoSQL.

**Estado actual:** 42 de 45 endpoints implementados y probados.

---

## Tecnologías

### Stack Principal

<p align="center">
  <img src="internal/decoration/goland.png" alt="Go" height="120" style="margin: 0 30px;"/>
  <img src="internal/decoration/gogin.png" alt="Gin" height="120" style="margin: 0 30px;"/>
  <img src="internal/decoration/firebase.png" alt="Firebase" height="120" style="margin: 0 30px;"/>
</p>

### Dependencias Principales

```
github.com/gin-gonic/gin              # Framework HTTP
github.com/golang-jwt/jwt/v5          # Autenticación JWT
golang.org/x/crypto/bcrypt            # Hash de contraseñas
github.com/google/uuid                # Generación de UUIDs
github.com/go-playground/validator    # Validación de requests
github.com/sirupsen/logrus            # Sistema de logs
cloud.google.com/go/firestore         # SDK de Firestore
github.com/redis/go-redis/v9          # Cliente Redis
github.com/gin-contrib/cors           # Middleware CORS
```

---

## Guía de Instalación

### Requisitos Previos

- Go 1.21 o superior
- Git
- Cuenta de Firebase con Firestore habilitado

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

1. Ir a [Firebase Console](https://console.firebase.google.com/)
2. Seleccionar tu proyecto o crear uno nuevo
3. Habilitar Firestore Database
4. Ir a **Project Settings** → **Service Accounts**
5. Click en **Generate New Private Key**
6. Guardar el archivo JSON como `serviceAccountKey.json` en la raíz del proyecto
7. Copiar el Project ID

### 4. Configurar Variables de Entorno

Copiar el archivo de ejemplo:

```bash
cp .env.example .env
```

Editar `.env` con las siguientes variables obligatorias:

```env
# General
GO_ENV=development
PORT=8080

# Firebase (OBLIGATORIO)
FIREBASE_PROJECT_ID=tu-proyecto-id
FIREBASE_CREDENTIALS_PATH=./serviceAccountKey.json

# JWT (OBLIGATORIO)
JWT_SECRET=tu-secreto-jwt-seguro
JWT_REFRESH_SECRET=tu-secreto-refresh-seguro
JWT_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=168h

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=1m
```

**Generar JWT Secrets:**

En macOS/Linux:
```bash
openssl rand -base64 32
```

En Windows PowerShell:
```powershell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})
```

Ejecutar dos veces para generar `JWT_SECRET` y `JWT_REFRESH_SECRET`.

### 5. Iniciar el Backend

**Modo desarrollo:**
```bash
go run cmd/api/main.go
```

**Compilar y ejecutar:**
```bash
go build -o backend.exe cmd/api/main.go
./backend.exe
```

### 6. Verificar Funcionamiento

```bash
curl http://localhost:8080/health
```

Respuesta esperada:
```json
{
  "status": "ok",
  "message": "Cheos Café Backend API is running",
  "version": "v1"
}
```

---

## Manual de Usuario

### Tipos de Usuarios

- **CUSTOMER:** Usuario regular. Puede ver productos, crear órdenes y dejar reseñas.
- **ADMIN:** Administrador. Acceso completo para gestionar productos, órdenes, reseñas, códigos de descuento y ubicaciones.

### Autenticación

Endpoints protegidos requieren un token JWT en el header:

```
Authorization: Bearer <access_token>
```

Los access tokens expiran en 15 minutos. Los refresh tokens expiran en 7 días. Usar `/api/v1/auth/refresh` para renovar el token.

### Flujo de Compra (Cliente)

1. **Ver productos:** `GET /api/v1/products`
2. **Buscar productos:** `GET /api/v1/products/search?q=premium`
3. **Ver detalle:** `GET /api/v1/products/{id}`
4. **Crear orden:** `POST /api/v1/orders` (sin autenticación o con ella)
5. **Rastrear orden:** `GET /api/v1/orders/number/{order_number}`

### Flujo de Gestión (Administrador)

1. **Login:** `POST /api/v1/auth/login`
2. **Crear producto:** `POST /api/v1/products`
3. **Actualizar stock:** `PATCH /api/v1/products/{id}/stock`
4. **Ver órdenes:** `GET /api/v1/orders`
5. **Actualizar estado de orden:** `PATCH /api/v1/orders/{id}/status`
6. **Moderar reseñas:** `PUT /api/v1/reviews/{id}`
7. **Crear códigos de descuento:** `POST /api/v1/discounts`

### Estados de Orden

- **PENDING:** Orden creada, esperando confirmación
- **CONFIRMED:** Pago confirmado
- **PROCESSING:** Orden en preparación
- **SHIPPED:** Orden enviada
- **DELIVERED:** Orden entregada
- **CANCELLED:** Orden cancelada

### Métodos de Pago

- **CONTRA_ENTREGA:** Pago al recibir (activo)
- **TARJETA:** Tarjeta de crédito/débito (preparado)
- **PSE:** Transferencia bancaria (preparado)

---

## Endpoints Disponibles

### Resumen por Servicio

| Servicio | Total | Públicos | Autenticados | Admin |
|----------|-------|----------|--------------|-------|
| Health Checks | 3 | 3 | 0 | 0 |
| Authentication | 5 | 3 | 2 | 0 |
| Products | 8 | 4 | 0 | 4 |
| Orders | 8 | 2 | 2 | 4 |
| Discounts | 6 | 1 | 0 | 5 |
| Reviews | 6 | 2 | 0 | 4 |
| Locations | 6 | 2 | 0 | 4 |
| **TOTAL** | **42** | **17** | **4** | **21** |

### Endpoints Principales

**Health Checks**
```
GET  /health
GET  /health/firebase
GET  /health/redis
```

**Autenticación**
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
GET  /api/v1/users/me
PUT  /api/v1/users/me
```

**Productos**
```
GET    /api/v1/products
GET    /api/v1/products/featured
GET    /api/v1/products/search
GET    /api/v1/products/:id
POST   /api/v1/products              (Admin)
PUT    /api/v1/products/:id          (Admin)
PATCH  /api/v1/products/:id/stock    (Admin)
DELETE /api/v1/products/:id          (Admin)
```

**Órdenes**
```
POST  /api/v1/orders
GET   /api/v1/orders/number/:number
GET   /api/v1/orders/me              (User)
GET   /api/v1/orders/:id             (User)
GET   /api/v1/orders                 (Admin)
PATCH /api/v1/orders/:id/status      (Admin)
PATCH /api/v1/orders/:id/payment     (Admin)
```

**Códigos de Descuento**
```
POST   /api/v1/discounts/validate
GET    /api/v1/discounts             (Admin)
POST   /api/v1/discounts             (Admin)
GET    /api/v1/discounts/:id         (Admin)
PUT    /api/v1/discounts/:id         (Admin)
DELETE /api/v1/discounts/:id         (Admin)
```

**Reseñas**
```
POST   /api/v1/reviews
GET    /api/v1/products/:id/reviews
GET    /api/v1/reviews               (Admin)
GET    /api/v1/reviews/:id           (Admin)
PUT    /api/v1/reviews/:id           (Admin)
DELETE /api/v1/reviews/:id           (Admin)
```

**Ubicaciones**
```
GET    /api/v1/locations
GET    /api/v1/locations/:id
GET    /api/v1/locations/all         (Admin)
POST   /api/v1/locations             (Admin)
PUT    /api/v1/locations/:id         (Admin)
DELETE /api/v1/locations/:id         (Admin)
```

**Nota:** Documentación completa en `Endpoints_Test.md` y colección Postman en `Cheos_Cafe_API.postman_collection.json`.

---

## Estructura del Proyecto

```
GoBackend_Cheos/
├── cmd/
│   └── api/
│       └── main.go                    # Punto de entrada
│
├── internal/
│   ├── config/                        # Configuración
│   ├── database/                      # Conexiones (Firebase, Redis)
│   ├── handlers/                      # Controladores HTTP
│   ├── middleware/                    # Middlewares (Auth, CORS, Rate Limit)
│   ├── models/                        # Modelos de datos
│   ├── repository/                    # Capa de acceso a datos
│   ├── services/                      # Lógica de negocio
│   └── utils/                         # Utilidades (JWT, Password, Validation)
│
├── .env.example                       # Template variables de entorno
├── .gitignore                         # Reglas Git
├── Cheos_Cafe_API.postman_collection.json
├── Endpoints_Test.md
├── go.mod
├── go.sum
└── README.md
```

---

## Comandos Útiles

```bash
# Desarrollo
go run cmd/api/main.go              # Ejecutar servidor
go mod download                     # Descargar dependencias
go mod tidy                         # Limpiar dependencias

# Build
go build -o backend.exe cmd/api/main.go     # Windows
go build -o backend cmd/api/main.go         # Linux/Mac

# Code Quality
go fmt ./...                        # Formatear código
go vet ./...                        # Analizar código
```

- Fin... Gracias por leerlo todo <3