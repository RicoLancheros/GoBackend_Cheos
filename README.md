# Cheos Cafe - Backend API

Backend API REST para el sistema de e-commerce de Cheos Cafe, desarrollado en Go con Firebase Firestore.

---

## Descripcion del Proyecto

Repositorio del backend completo para la plataforma de e-commerce de Cheos Cafe, empresa colombiana con 8 tiendas fisicas en Antioquia dedicada a la venta de cafe molido de especialidad.

El backend proporciona una API REST con las siguientes funcionalidades:

- Autenticacion de usuarios con JWT (roles ADMIN/CUSTOMER)
- Registro con campos opcionales (ciudad, municipio, barrio, genero, fecha de nacimiento)
- Recuperacion de contrasena por correo electronico
- Gestion de catalogo de productos con busqueda y filtros
- Sistema de ordenes con checkout para usuarios autenticados e invitados
- Carrito de compras persistente por usuario
- Codigos de descuento con validacion y limites de uso
- Integracion con pasarela de pagos Wompi
- Sistema de resenas con moderacion
- Gestion de ubicaciones de tiendas fisicas
- Galeria de imagenes con subida a Cloudinary
- Configuracion del sitio (carrusel y seccion About Us)
- Dashboard de metricas con sistema event-driven
- Rate limiting independiente para login y global
- CORS configurado para produccion

Arquitectura de tres capas (Repository, Service, Handler) con Firebase Firestore como base de datos NoSQL.

**Estado actual:** 65+ endpoints implementados. Backend desplegado en Render, frontend en Netlify.

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
git clone https://github.com/RicoLancheros/GoBackend_Cheos.git
cd GoBackend_Cheos
```

### 2. Instalar Dependencias

```bash
go mod download
go mod tidy
```

### 3. Configurar Firebase

1. Ir a Firebase Console (https://console.firebase.google.com/)
2. Seleccionar el proyecto o crear uno nuevo
3. Habilitar Firestore Database
4. Ir a Project Settings > Service Accounts
5. Click en Generate New Private Key
6. Guardar el archivo JSON como `firebase-credentials.json` en la raiz del proyecto

### 4. Configurar Variables de Entorno

Copiar el archivo de ejemplo:

```bash
cp .env.example .env
```

Variables requeridas:

```env
# General
GO_ENV=development
PORT=8080

# Firebase
FIREBASE_PROJECT_ID=tu-proyecto-id
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json

# JWT
JWT_SECRET=tu-secreto-jwt-seguro
JWT_REFRESH_SECRET=tu-secreto-refresh-seguro
JWT_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=168h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# Cloudinary
CLOUDINARY_CLOUD_NAME=tu-cloud-name
CLOUDINARY_API_KEY=tu-api-key
CLOUDINARY_API_SECRET=tu-api-secret

# Email (SMTP para desarrollo local)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=tu-email@gmail.com
SMTP_PASSWORD=tu-app-password

# Email (Resend para produccion)
RESEND_API_KEY=re_xxxxx
RESEND_FROM=onboarding@resend.dev

# Frontend URL (para enlaces en emails)
FRONTEND_URL=http://localhost:5173

# Wompi (pasarela de pagos)
WOMPI_PRIVATE_KEY=prv_test_xxxxx
WOMPI_PUBLIC_KEY=pub_test_xxxxx
WOMPI_EVENTS_SECRET=test_events_xxxxx
WOMPI_INTEGRITY_KEY=test_integrity_xxxxx
WOMPI_API_BASE=https://sandbox.wompi.co/v1
```

### 5. Iniciar el Backend

```bash
go run cmd/api/main.go
```

### 6. Verificar Funcionamiento

```bash
curl http://localhost:8080/health
```

---

## Deployment

### Backend en Render

Variables de entorno requeridas en Render:

```
GO_ENV=production
PORT=8080
FIREBASE_PROJECT_ID=<project-id>
FIREBASE_CREDENTIALS_JSON=<JSON completo del archivo de credenciales>
JWT_SECRET=<secret>
JWT_REFRESH_SECRET=<secret>
CORS_ALLOWED_ORIGINS=https://cheoscafe.netlify.app
CLOUDINARY_CLOUD_NAME=<cloud-name>
CLOUDINARY_API_KEY=<api-key>
CLOUDINARY_API_SECRET=<api-secret>
RESEND_API_KEY=<api-key>
RESEND_FROM=onboarding@resend.dev
FRONTEND_URL=https://cheoscafe.netlify.app
WOMPI_PRIVATE_KEY=<key>
WOMPI_PUBLIC_KEY=<key>
WOMPI_EVENTS_SECRET=<secret>
WOMPI_INTEGRITY_KEY=<key>
WOMPI_API_BASE=https://sandbox.wompi.co/v1
```

Nota: En Render se usa `FIREBASE_CREDENTIALS_JSON` (JSON completo como string) en lugar del archivo local. El servicio de email usa Resend (HTTP) en produccion ya que Render bloquea puertos SMTP.

### Frontend en Netlify

```
VITE_API_URL=https://gobackend-cheos.onrender.com/api/v1
```

---

## Endpoints (65+ totales)

### Resumen por Servicio

| Servicio | Endpoints | Acceso |
|----------|-----------|--------|
| Health Checks | 3 | Publico |
| Autenticacion | 6 | Publico |
| Usuarios | 5 | User / Admin |
| Productos | 8 | Publico / Admin |
| Ordenes | 7 | Publico / User / Admin |
| Carrito | 6 | User |
| Descuentos | 6 | Publico / Admin |
| Resenas | 6 | Publico / Admin |
| Ubicaciones | 6 | Publico / Admin |
| Galeria | 8 | Publico / Admin |
| Config Sitio | 4 | Publico / Admin |
| Pagos Wompi | 3 | Publico |
| Dashboard | 8 | Admin |
| Ping | 1 | Publico |

### Autenticacion y Usuarios

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login                 (rate limited: 5 intentos / 15 min)
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
GET    /api/v1/users/me                   (User)
PUT    /api/v1/users/me                   (User)
GET    /api/v1/users                      (Admin)
PUT    /api/v1/users/:id                  (Admin)
DELETE /api/v1/users/:id                  (Admin)
```

### Productos

```
GET    /api/v1/products
GET    /api/v1/products/featured
GET    /api/v1/products/search?q=
GET    /api/v1/products/:id
POST   /api/v1/products                   (Admin)
PUT    /api/v1/products/:id               (Admin)
PATCH  /api/v1/products/:id/stock         (Admin)
DELETE /api/v1/products/:id               (Admin)
```

### Ordenes

```
POST   /api/v1/orders
GET    /api/v1/orders/number/:number
GET    /api/v1/orders/me                  (User)
GET    /api/v1/orders/:id                 (User)
GET    /api/v1/orders                     (Admin)
PATCH  /api/v1/orders/:id/status          (Admin)
PATCH  /api/v1/orders/:id/payment         (Admin)
```

### Carrito

```
GET    /api/v1/cart                       (User)
POST   /api/v1/cart/items                 (User)
PUT    /api/v1/cart/items/:productId      (User)
DELETE /api/v1/cart/items/:productId      (User)
DELETE /api/v1/cart                       (User)
POST   /api/v1/cart/sync                  (User)
```

### Descuentos, Resenas, Ubicaciones

```
POST   /api/v1/discounts/validate
GET    /api/v1/discounts                  (Admin)
POST   /api/v1/discounts                  (Admin)
GET    /api/v1/discounts/:id              (Admin)
PUT    /api/v1/discounts/:id              (Admin)
DELETE /api/v1/discounts/:id              (Admin)

POST   /api/v1/reviews
GET    /api/v1/products/:id/reviews
GET    /api/v1/reviews                    (Admin)
GET    /api/v1/reviews/:id                (Admin)
PUT    /api/v1/reviews/:id                (Admin)
DELETE /api/v1/reviews/:id                (Admin)

GET    /api/v1/locations
GET    /api/v1/locations/all
GET    /api/v1/locations/:id
POST   /api/v1/locations                  (Admin)
PUT    /api/v1/locations/:id              (Admin)
DELETE /api/v1/locations/:id              (Admin)
```

### Galeria y Configuracion del Sitio

```
GET    /api/v1/gallery/active
GET    /api/v1/gallery/type/:type
GET    /api/v1/gallery/:id
GET    /api/v1/gallery                    (Admin)
POST   /api/v1/gallery                    (Admin)
POST   /api/v1/gallery/upload             (Admin)
PUT    /api/v1/gallery/:id                (Admin)
DELETE /api/v1/gallery/:id                (Admin)

GET    /api/v1/config/carousel
GET    /api/v1/config/about
PUT    /api/v1/config/carousel            (Admin)
PUT    /api/v1/config/about               (Admin)
```

### Pagos Wompi

```
POST   /api/v1/payments/wompi/signature
GET    /api/v1/payments/wompi/transaction/:id
POST   /api/v1/payments/wompi/webhook
```

### Dashboard (Admin)

```
GET    /api/v1/dashboard/sales/monthly
GET    /api/v1/dashboard/sales/yearly
GET    /api/v1/dashboard/buyers/monthly
GET    /api/v1/dashboard/buyers/yearly
GET    /api/v1/dashboard/products/monthly
GET    /api/v1/dashboard/products/yearly
GET    /api/v1/dashboard/summary
POST   /api/v1/dashboard/recalculate
```

---

## Estructura del Proyecto

```
GoBackend_Cheos/
├── cmd/api/
│   └── main.go                          # Punto de entrada, DI y rutas
├── internal/
│   ├── config/config.go                 # Configuracion desde variables de entorno
│   ├── database/
│   │   ├── firebase.go                  # Conexion Firebase/Firestore
│   │   └── redis.go                     # Conexion Redis (opcional)
│   ├── handlers/                        # Controladores HTTP (11 handlers)
│   ├── middleware/                       # Auth, CORS, Rate Limiting, Admin
│   ├── models/                          # Modelos de datos y DTOs (11 modelos)
│   ├── repository/                      # Capa de acceso a Firestore (11 repos)
│   ├── services/                        # Logica de negocio (14 servicios)
│   └── utils/                           # JWT, password hash, validador, responses
├── .env.example
├── Dockerfile
├── Cheos_Cafe_API.postman_collection.json
├── go.mod / go.sum
└── README.md
```

---

## Comandos

```bash
go run cmd/api/main.go                   # Ejecutar servidor
go build -o backend cmd/api/main.go      # Compilar
go mod download                          # Descargar dependencias
go mod tidy                              # Limpiar dependencias
go fmt ./...                             # Formatear codigo
go vet ./...                             # Analizar codigo
```

Documentacion detallada de endpoints con ejemplos en la coleccion Postman incluida en el repositorio.
