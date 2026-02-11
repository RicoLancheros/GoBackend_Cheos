# Migración de Proyecto - Cheos Café Backend + Frontend

## Contexto General del Proyecto

E-commerce SPA completo de venta de café molido de especialidad para **Cheos Café**, empresa colombiana con 8 tiendas físicas en Antioquia.

### Información del Proyecto:
- **Nombre:** Cheos Café E-commerce
- **Backend:** Go 1.21+ con Gin Framework (`GoBackend_Cheos/`)
- **Frontend:** React + Vite + Tailwind CSS (`ReactFront_Cheos/`)
- **Base de Datos:** Firebase Firestore (NoSQL) - única fuente de verdad
- **Autenticación:** JWT con cookies + headers (access: 15min, refresh: 7 días)
- **Patrón de Arquitectura:** Repository-Service-Handler (3 capas)
- **Imágenes:** Cloudinary para upload y hosting (cloud name: detib7vvw)
- **Puerto:** 8080
- **Versión API:** v1
- **Roles:** ADMIN y CUSTOMER
- **Redis:** Opcional, no configurado actualmente

---

## Deployment Actual

- **Backend:** Render (Docker) - `https://gobackend-cheos.onrender.com`
- **Frontend:** Netlify - `https://cheoscafesena.netlify.app`
- **Firebase:** Proyecto `golandbackend-cheos` (activo y conectado)
- Las credenciales de Firebase se pasan como variable de entorno `FIREBASE_CREDENTIALS_JSON` en Render (JSON completo como string)

### Variables de entorno en Render (backend)
```
GO_ENV=production
PORT=8080
FIREBASE_PROJECT_ID=golandbackend-cheos
FIREBASE_CREDENTIALS_JSON=<JSON completo del archivo de credenciales>
JWT_SECRET=<secret seguro>
JWT_REFRESH_SECRET=<secret seguro>
JWT_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=168h
CORS_ALLOWED_ORIGINS=https://cheoscafesena.netlify.app
CLOUDINARY_CLOUD_NAME=detib7vvw
CLOUDINARY_API_KEY=549224315686314
CLOUDINARY_API_SECRET=<secret>
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=15m
```

### Variable de entorno en Netlify (frontend)
```
VITE_API_URL=https://gobackend-cheos.onrender.com/api/v1
```

---

## Estado Actual del Desarrollo

### Completado

#### 1. Gestión de Usuarios (Admin)
- `GET /api/v1/users` - Listar todos los usuarios
- `PUT /api/v1/users/:id` - Actualizar usuario por ID (nombre, email, password, rol, is_active)
- `DELETE /api/v1/users/:id` - Eliminar usuario (hard delete)
- Validación de email único al actualizar
- Hashing de contraseña con bcrypt

#### 2. Campo map_iframe para Google Maps en Ubicaciones
- Campo `MapIframe string` en el modelo Location
- Permite almacenar iframes de Google Maps

#### 3. Hard Delete en TODOS los endpoints DELETE
- Todos los DELETE eliminan físicamente de Firebase
- Aplica a: productos, usuarios, ubicaciones, galería, descuentos, reseñas

#### 4. Deploy del Backend en Render
- `firebase.go` soporta credenciales desde variable de entorno (`FIREBASE_CREDENTIALS_JSON`) además del archivo local
- Se agregó `FirebaseCredentialsJSON` al config.go
- Se creó `.dockerignore` para optimizar el build
- Dockerfile multi-stage ya existía

#### 5. Fix: URLs hardcodeadas en el frontend
- 4 componentes tenían `http://localhost:8080/api/v1` hardcodeado en vez de usar `import.meta.env.VITE_API_URL`
- Archivos corregidos: `EditProductModal.jsx`, `AddProductModal.jsx`, `GalleryManagementModal.jsx`, `GalleryImageSelector.jsx`

#### 6. Fix: Gestión de usuarios no funcionaba
- El frontend usaba método `PATCH` pero el backend espera `PUT` para actualizar usuarios
- Se corrigió el método HTTP en `UserManagementModal.jsx`
- Se agregó funcionalidad de eliminar usuarios (con confirmación)
- Se agregó feedback visual (colores para estados, alertas de éxito/error)

#### 7. Fix: Galería no cargaba imágenes
- El `useEffect` no tenía `token` como dependencia
- Se agregó el token como dependencia y guard clause
- Se eliminaron campos inútiles del formulario de subida (etiquetas, imagen activa)
- Se agregó funcionalidad de editar imágenes existentes (título, descripción, tipo)

#### 8. Feature: Carrusel editable por admin
- Se creó `CarouselEditorModal.jsx` para que el admin seleccione hasta 6 imágenes
- Se filtran solo imágenes de tipo GENERAL y CAROUSEL para el carrusel
- Se creó endpoint `GET/PUT /api/v1/config/carousel` en el backend (colección `site_config` en Firebase)
- El `HeroCarousel.jsx` carga imágenes desde la API, no desde localStorage
- Se agregó botón "Editar carrusel" en el menú de admin del Navbar
- Archivos backend nuevos: `site_config.go` (modelo), `site_config_repository.go`, `site_config_service.go`, `site_config_handler.go`

#### 9. Feature: GalleryImageSelector filtra por tipo
- Nuevo prop `allowedTypes` (default: `['GENERAL', 'PRODUCT']`)
- Para productos solo muestra imágenes de tipo GENERAL y PRODUCT
- Para carrusel solo muestra imágenes de tipo GENERAL y CAROUSEL

#### 10. Fix: Rate limiter compartido (bug crítico)
- El rate limiter global y el de login compartían el mismo mapa `visitors`
- Se separaron en mapas independientes: `globalVisitors` y `loginVisitors`
- Login ahora tiene su propio límite: 5 intentos / 15 minutos
- Se evitó la creación de múltiples goroutines de cleanup

---

## Estructura del Proyecto (Backend)

```
GoBackend_Cheos/
├── cmd/api/main.go                        # Punto de entrada + rutas
├── internal/
│   ├── config/config.go                   # Configuración desde .env
│   ├── database/
│   │   ├── firebase.go                    # Conexión Firebase (archivo o env var)
│   │   └── redis.go                       # Conexión Redis (opcional)
│   ├── handlers/                          # 8 handlers
│   │   ├── auth_handler.go                # Login, registro, gestión usuarios
│   │   ├── product_handler.go             # CRUD productos
│   │   ├── order_handler.go               # Gestión de órdenes
│   │   ├── discount_handler.go            # Códigos de descuento
│   │   ├── review_handler.go              # Reseñas de productos
│   │   ├── location_handler.go            # CRUD ubicaciones
│   │   ├── gallery_handler.go             # Galería de imágenes
│   │   └── site_config_handler.go         # Configuración del sitio (carrusel)
│   ├── middleware/
│   │   ├── auth.go                        # AuthMiddleware (JWT)
│   │   ├── cors.go                        # CORS
│   │   ├── rate_limit.go                  # Rate limiting (global + login separados)
│   │   └── admin.go                       # RequireAdmin
│   ├── models/                            # 8 modelos + DTOs
│   │   ├── user.go
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── location.go
│   │   ├── gallery.go
│   │   ├── review.go
│   │   ├── discount.go
│   │   └── site_config.go                 # Modelo de configuración del sitio
│   ├── repository/                        # 8 repositorios
│   │   ├── user_repository.go
│   │   ├── product_repository.go
│   │   ├── order_repository.go
│   │   ├── location_repository.go
│   │   ├── gallery_repository.go
│   │   ├── review_repository.go
│   │   ├── discount_repository.go
│   │   └── site_config_repository.go      # Colección site_config en Firebase
│   ├── services/                          # 8 servicios + upload_service
│   │   ├── auth_service.go
│   │   ├── product_service.go
│   │   ├── order_service.go
│   │   ├── location_service.go
│   │   ├── gallery_service.go
│   │   ├── review_service.go
│   │   ├── discount_service.go
│   │   ├── site_config_service.go
│   │   └── upload_service.go              # Cloudinary upload
│   └── utils/
│       ├── jwt.go                         # Generación y validación JWT
│       ├── password.go                    # Hashing bcrypt
│       ├── response.go                    # Respuestas HTTP estandarizadas
│       ├── validator.go                   # Validación de structs
│       └── order_number.go               # Generación de números de orden
├── .env                                   # Variables de entorno (NO commitear)
├── .env.example                           # Template con documentación
├── .dockerignore                          # Exclusiones para Docker build
├── firebase-credentials.json              # Credenciales Firebase (NO commitear)
├── Dockerfile                             # Multi-stage build para Render
├── Cheos_Cafe_API.postman_collection.json
├── go.mod / go.sum
├── README.md
├── MigracionClaude.md                     # Este archivo
└── ContextoSesion.md                      # Contexto detallado de la última sesión
```

---

## API Endpoints (46 totales)

### Health Checks (3)
```
GET  /health
GET  /health/firebase
GET  /health/redis
```

### Auth (5)
```
POST /api/v1/auth/register          (público)
POST /api/v1/auth/login             (público, rate limited: 5 intentos/15min)
POST /api/v1/auth/refresh           (público)
POST /api/v1/auth/logout            (público)
GET  /api/v1/ping                   (público)
```

### Users (5)
```
GET  /api/v1/users/me               (usuario)
PUT  /api/v1/users/me               (usuario) - solo nombre y teléfono
GET  /api/v1/users                  (admin)
PUT  /api/v1/users/:id              (admin) - nombre, email, password, rol, is_active
DELETE /api/v1/users/:id            (admin) - hard delete
```

### Products (8)
```
GET    /api/v1/products             (público)
GET    /api/v1/products/featured    (público)
GET    /api/v1/products/search?q=   (público)
GET    /api/v1/products/:id         (público)
POST   /api/v1/products             (admin)
PUT    /api/v1/products/:id         (admin)
DELETE /api/v1/products/:id         (admin)
PATCH  /api/v1/products/:id/stock   (admin)
```

### Orders (7)
```
POST  /api/v1/orders                (público - guest checkout)
GET   /api/v1/orders/number/:number (público)
GET   /api/v1/orders/me             (usuario)
GET   /api/v1/orders/:id            (usuario)
GET   /api/v1/orders                (admin)
PATCH /api/v1/orders/:id/status     (admin)
PATCH /api/v1/orders/:id/payment    (admin)
```

### Discounts (6)
```
POST   /api/v1/discounts/validate   (público)
GET    /api/v1/discounts            (admin)
POST   /api/v1/discounts            (admin)
GET    /api/v1/discounts/:id        (admin)
PUT    /api/v1/discounts/:id        (admin)
DELETE /api/v1/discounts/:id        (admin)
```

### Reviews (6)
```
POST   /api/v1/reviews              (público)
GET    /api/v1/products/:id/reviews (público)
GET    /api/v1/reviews              (admin)
GET    /api/v1/reviews/:id          (admin)
PUT    /api/v1/reviews/:id          (admin)
DELETE /api/v1/reviews/:id          (admin)
```

### Locations (6)
```
GET    /api/v1/locations            (público)
GET    /api/v1/locations/all        (público)
GET    /api/v1/locations/:id        (público)
POST   /api/v1/locations            (admin)
PUT    /api/v1/locations/:id        (admin)
DELETE /api/v1/locations/:id        (admin)
```

### Gallery (8)
```
GET    /api/v1/gallery/active       (público)
GET    /api/v1/gallery/type/:type   (público)
GET    /api/v1/gallery/:id          (público)
GET    /api/v1/gallery              (admin)
POST   /api/v1/gallery              (admin)
POST   /api/v1/gallery/upload       (admin)
PUT    /api/v1/gallery/:id          (admin)
DELETE /api/v1/gallery/:id          (admin)
```

### Site Config (2)
```
GET  /api/v1/config/carousel        (público)
PUT  /api/v1/config/carousel        (admin) - body: {"images": ["url1", "url2", ...]}
```

---

## Datos Técnicos Importantes

### Credenciales de prueba
- **Admin:** `admin@cheoscafe.com` / `Admin123!`

### Rate limiting
- Global: 100 requests / 15 minutos por IP
- Login: 5 intentos / 15 minutos por IP (mapa independiente)

### Decisiones de diseño vigentes
- **Hard Delete en TODO:** Todos los DELETE eliminan físicamente de Firebase
- **UUIDs:** Se usan UUIDs en vez de auto-increment IDs
- **JWT:** Access token 15min, refresh token 168h (7 días)
- **Cloudinary:** Para subida de imágenes (cloud name: detib7vvw)
- **Sin Redis:** El sistema funciona sin cache, Redis es opcional
- **Firebase credentials:** Soporta archivo local o variable de entorno (para Render)
- **Autenticación:** JWT en Cookie `access_token` o Header `Authorization: Bearer <token>`

### Colecciones de Firebase
- `users` - Usuarios
- `products` - Productos
- `orders` - Órdenes
- `discounts` - Códigos de descuento
- `reviews` - Reseñas
- `locations` - Ubicaciones de tiendas
- `gallery` - Imágenes de galería
- `site_config` - Configuración del sitio (documento `carousel`)

---

## Seguridad Implementada

- Hashing de contraseñas con bcrypt
- JWT para autenticación (access + refresh tokens)
- Validación de roles (ADMIN/CUSTOMER)
- Validación de inputs con `validator`
- Validación de email único
- Cookies HttpOnly para tokens
- Rate limiting independiente para login (5 intentos / 15 minutos)
- Rate limiting global (100 req / 15 minutos)
- CORS configurado para producción

---

## Pendiente

1. **Email en órdenes:** Cuando un usuario autenticado crea una orden, se usa un email predeterminado en vez del email del JWT. Revisar `order_handler.go` / `order_service.go`.
2. **Pasarela de pagos con Wompi:** Pendiente de integrar.
3. **Funcionalidades adicionales del frontend:** Por definir.

---

## Notas para el Nuevo Claude

- Backend desplegado en Render: `https://gobackend-cheos.onrender.com`
- Frontend desplegado en Netlify: `https://cheoscafesena.netlify.app`
- Firebase activo y conectado (proyecto: `golandbackend-cheos`)
- El rate limiter del login es independiente del global (ya corregido)
- Leer `ContextoSesion.md` para contexto detallado de la última sesión
- Leer `README.md` del backend para documentación completa
- La colección Postman está en `Cheos_Cafe_API.postman_collection.json`
- **NUNCA usar Soft Delete** - Todos los DELETE deben ser físicos
- **Usar el patrón Repository-Service-Handler** - Mantener la arquitectura en 3 capas
- **Usar UUIDs** en lugar de auto-increment IDs
- **Firebase es la única base de datos** - No hay SQL
- El sistema operativo del entorno de desarrollo actual es Linux

### Patrón para crear nuevos módulos (ejemplo: site_config)
1. Crear modelo en `internal/models/`
2. Crear repositorio en `internal/repository/`
3. Crear servicio en `internal/services/`
4. Crear handler en `internal/handlers/`
5. Registrar en `cmd/api/main.go` (inicialización + rutas)

### Admin-only routes usan dos middlewares:
```go
adminGroup := group.Group("")
adminGroup.Use(middleware.AuthMiddleware(cfg))
adminGroup.Use(middleware.RequireAdmin())
```

---

Fecha última actualización: 2026-02-11
