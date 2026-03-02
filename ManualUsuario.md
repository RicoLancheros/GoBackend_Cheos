# Manual de Usuario - Cheos Cafe API

Documentacion de todos los servicios disponibles en la API REST de Cheos Cafe.

**Base URL desarrollo:** `http://localhost:8080/api/v1`
**Base URL produccion:** `https://gobackend-cheos.onrender.com/api/v1`

---

## 1. Autenticacion

### Registro

```
POST /auth/register
```

Body:
```json
{
  "name": "Juan Perez",
  "email": "juan@email.com",
  "password": "miPassword123",
  "phone": "3001234567"
}
```

Campos opcionales en registro: `city`, `municipality`, `neighborhood`, `gender`, `birth_date`.

### Login

```
POST /auth/login
```

Body:
```json
{
  "email": "juan@email.com",
  "password": "miPassword123"
}
```

Respuesta exitosa:
```json
{
  "token": "eyJhbGciOi...",
  "refresh_token": "eyJhbGciOi...",
  "user": { "id": "...", "name": "Juan Perez", "email": "juan@email.com", "role": "CUSTOMER" }
}
```

Rate limit: 5 intentos cada 15 minutos por IP.

### Refresh Token

```
POST /auth/refresh
```

Body:
```json
{
  "refresh_token": "eyJhbGciOi..."
}
```

### Logout

```
POST /auth/logout
```

Header: `Authorization: Bearer <token>`

### Recuperar Contrasena

```
POST /auth/forgot-password
```

Body:
```json
{
  "email": "juan@email.com"
}
```

Envia un correo con enlace de restablecimiento. En desarrollo usa SMTP (Gmail), en produccion usa Resend HTTP API.

### Restablecer Contrasena

```
POST /auth/reset-password
```

Body:
```json
{
  "token": "token-recibido-por-email",
  "new_password": "nuevaPassword123"
}
```

El token expira en 15 minutos.

---

## 2. Usuarios

Requiere autenticacion. Header: `Authorization: Bearer <token>`

### Obtener Perfil

```
GET /users/me
```

### Actualizar Perfil

```
PUT /users/me
```

Body (campos opcionales):
```json
{
  "name": "Juan Perez",
  "phone": "3001234567",
  "city": "Medellin",
  "municipality": "Envigado",
  "neighborhood": "El Poblado",
  "gender": "masculino",
  "birth_date": "1990-05-15"
}
```

### Listar Usuarios (Admin)

```
GET /users
```

### Actualizar Usuario por ID (Admin)

```
PUT /users/:id
```

### Eliminar Usuario (Admin)

```
DELETE /users/:id
```

---

## 3. Productos

Las rutas de lectura son publicas. Las de escritura requieren rol ADMIN.

### Listar Productos

```
GET /products
```

Parametros opcionales: `?category=molido&sort=price_asc&limit=10&offset=0`

### Productos Destacados

```
GET /products/featured
```

### Buscar Productos

```
GET /products/search?q=cafe
```

### Obtener Producto

```
GET /products/:id
```

### Crear Producto (Admin)

```
POST /products
```

Body:
```json
{
  "name": "Cafe Especial",
  "description": "Cafe molido de origen",
  "price": 25000,
  "weight": "250g",
  "category": "molido",
  "stock": 100,
  "image_url": "https://...",
  "is_featured": true
}
```

### Actualizar Producto (Admin)

```
PUT /products/:id
```

### Actualizar Stock (Admin)

```
PATCH /products/:id/stock
```

Body:
```json
{
  "stock": 50
}
```

### Eliminar Producto (Admin)

```
DELETE /products/:id
```

---

## 4. Ordenes

### Crear Orden

```
POST /orders
```

Permite compras de usuarios autenticados e invitados.

Body (usuario autenticado):
```json
{
  "items": [
    { "product_id": "abc123", "quantity": 2 }
  ],
  "discount_code": "DESC10",
  "delivery_method": "delivery",
  "shipping_address": "Calle 10 #20-30, Medellin"
}
```

Body (invitado):
```json
{
  "items": [
    { "product_id": "abc123", "quantity": 2 }
  ],
  "guest_name": "Maria Lopez",
  "guest_email": "maria@email.com",
  "guest_phone": "3009876543",
  "delivery_method": "pickup"
}
```

### Consultar Orden por Numero

```
GET /orders/number/:number
```

Publico. Permite a cualquier persona consultar el estado de su orden con el numero.

### Mis Ordenes (User)

```
GET /orders/me
```

### Detalle de Orden (User)

```
GET /orders/:id
```

### Listar Todas las Ordenes (Admin)

```
GET /orders
```

### Actualizar Estado de Orden (Admin)

```
PATCH /orders/:id/status
```

Body:
```json
{
  "status": "completed"
}
```

Valores: `pending`, `confirmed`, `preparing`, `shipped`, `delivered`, `completed`, `cancelled`.

### Actualizar Estado de Pago (Admin)

```
PATCH /orders/:id/payment
```

Body:
```json
{
  "payment_status": "paid"
}
```

---

## 5. Carrito

Requiere autenticacion. Persistente por usuario en Firestore.

### Obtener Carrito

```
GET /cart
```

### Agregar Item

```
POST /cart/items
```

Body:
```json
{
  "product_id": "abc123",
  "quantity": 1
}
```

### Actualizar Cantidad

```
PUT /cart/items/:productId
```

Body:
```json
{
  "quantity": 3
}
```

### Eliminar Item

```
DELETE /cart/items/:productId
```

### Vaciar Carrito

```
DELETE /cart
```

### Sincronizar Carrito

```
POST /cart/sync
```

Body:
```json
{
  "items": [
    { "product_id": "abc123", "quantity": 2 },
    { "product_id": "def456", "quantity": 1 }
  ]
}
```

Fusiona el carrito local (del navegador) con el carrito del servidor al iniciar sesion.

---

## 6. Codigos de Descuento

### Validar Codigo (Publico)

```
POST /discounts/validate
```

Body:
```json
{
  "code": "DESC10",
  "purchase_total": 50000
}
```

Respuesta:
```json
{
  "valid": true,
  "discount_type": "percentage",
  "discount_value": 10,
  "message": "Descuento aplicado"
}
```

### CRUD de Descuentos (Admin)

```
GET    /discounts              # Listar todos
POST   /discounts              # Crear
GET    /discounts/:id          # Obtener uno
PUT    /discounts/:id          # Actualizar
DELETE /discounts/:id          # Eliminar
```

Body para crear/actualizar:
```json
{
  "code": "DESC10",
  "discount_type": "percentage",
  "value": 10,
  "min_purchase": 30000,
  "max_uses": 100,
  "is_active": true,
  "expires_at": "2025-12-31T23:59:59Z"
}
```

Tipos: `percentage` (porcentaje) o `fixed` (valor fijo en pesos).

---

## 7. Resenas

### Crear Resena (Publico)

```
POST /reviews
```

Body:
```json
{
  "product_id": "abc123",
  "author_name": "Juan",
  "rating": 5,
  "comment": "Excelente cafe"
}
```

### Resenas de un Producto (Publico)

```
GET /products/:id/reviews
```

### CRUD de Resenas (Admin)

```
GET    /reviews                # Listar todas
GET    /reviews/:id            # Obtener una
PUT    /reviews/:id            # Actualizar (moderar)
DELETE /reviews/:id            # Eliminar
```

---

## 8. Ubicaciones de Tiendas

### Obtener Activas (Publico)

```
GET /locations
```

### Obtener Todas (Publico)

```
GET /locations/all
```

### Obtener Una (Publico)

```
GET /locations/:id
```

### CRUD de Ubicaciones (Admin)

```
POST   /locations
PUT    /locations/:id
DELETE /locations/:id
```

Body:
```json
{
  "name": "Cheos Cafe Envigado",
  "address": "Calle 37 Sur #27-23",
  "city": "Envigado",
  "latitude": 6.1716,
  "longitude": -75.5866,
  "phone": "3001234567",
  "hours": "Lun-Sab 7:00-19:00",
  "is_active": true
}
```

---

## 9. Galeria de Imagenes

### Imagenes Activas (Publico)

```
GET /gallery/active
```

### Imagenes por Tipo (Publico)

```
GET /gallery/type/:type
```

### Obtener Imagen (Publico)

```
GET /gallery/:id
```

### CRUD de Galeria (Admin)

```
GET    /gallery                # Listar todas
POST   /gallery                # Crear registro
POST   /gallery/upload         # Subir imagen a Cloudinary
PUT    /gallery/:id            # Actualizar
DELETE /gallery/:id            # Eliminar
```

El endpoint `/gallery/upload` acepta multipart/form-data con el campo `image`. Las imagenes se almacenan en Cloudinary.

---

## 10. Configuracion del Sitio

### Carrusel (Publico lectura, Admin escritura)

```
GET /config/carousel
PUT /config/carousel           (Admin)
```

### About Us (Publico lectura, Admin escritura)

```
GET /config/about
PUT /config/about              (Admin)
```

---

## 11. Pagos Wompi

Integracion con la pasarela de pagos colombiana Wompi.

### Generar Firma de Integridad

```
POST /payments/wompi/signature
```

Body:
```json
{
  "reference": "order-123",
  "amount_in_cents": 5000000,
  "currency": "COP"
}
```

Se llama desde el frontend antes de redirigir al widget de pago de Wompi. La firma garantiza la integridad de la transaccion.

### Consultar Transaccion

```
GET /payments/wompi/transaction/:id
```

Verifica el estado de una transaccion despues de que el usuario regresa de Wompi.

### Webhook

```
POST /payments/wompi/webhook
```

Wompi llama a este endpoint automaticamente cuando cambia el estado de una transaccion. No requiere autenticacion (Wompi envia directamente).

---

## 12. Dashboard de Metricas (Admin)

Requiere rol ADMIN. Todos los datos se calculan mediante un sistema event-driven que actualiza metricas al crear ordenes.

### Ventas

```
GET /dashboard/sales/monthly       # Ventas del mes actual
GET /dashboard/sales/yearly        # Ventas del ano actual
```

### Compradores

```
GET /dashboard/buyers/monthly      # Compradores del mes
GET /dashboard/buyers/yearly       # Compradores del ano
```

### Productos mas Vendidos

```
GET /dashboard/products/monthly    # Top productos del mes
GET /dashboard/products/yearly     # Top productos del ano
```

### Resumen General

```
GET /dashboard/summary
```

Retorna: total de ventas, ordenes, clientes y productos activos.

### Recalcular Metricas

```
POST /dashboard/recalculate
```

Body:
```json
{
  "year": 2025,
  "month": 6
}
```

Recalcula las metricas de un mes especifico a partir de las ordenes existentes.

---

## 13. Health Checks

```
GET /health                        # Estado general del servidor
GET /health/firebase               # Conexion a Firebase
GET /health/redis                  # Conexion a Redis
GET /api/v1/ping                   # Test basico (retorna "pong")
```

---

## Autenticacion

Todos los endpoints protegidos requieren el header:

```
Authorization: Bearer <token_jwt>
```

El token de acceso expira en 15 minutos. Usar el endpoint `/auth/refresh` con el refresh token para obtener uno nuevo. El refresh token expira en 7 dias.

## Roles

| Rol | Descripcion |
|-----|-------------|
| CUSTOMER | Usuario registrado. Acceso a perfil, carrito, ordenes propias |
| ADMIN | Acceso completo. CRUD de todos los recursos, dashboard, gestion de usuarios |

## Codigos de Estado

| Codigo | Significado |
|--------|-------------|
| 200 | Exito |
| 201 | Recurso creado |
| 400 | Error en la solicitud (datos invalidos) |
| 401 | No autenticado (token faltante o invalido) |
| 403 | No autorizado (rol insuficiente) |
| 404 | Recurso no encontrado |
| 429 | Rate limit excedido |
| 500 | Error interno del servidor |
