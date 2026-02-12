# Carrito de Compras Persistente en Firebase Firestore

**Branch:** `feature/persistent-cart`
**Description:** Implementar carrito de compras persistente en Firestore para usuarios autenticados, manteniendo la experiencia de invitado para usuarios no logueados y fusionando ambos carritos al hacer login.

## Goal

Actualmente el carrito de compras solo vive en memoria (React `useState`) y se pierde al refrescar la pagina. Este feature agrega persistencia en Firebase Firestore para usuarios autenticados, con sincronizacion en background (optimistic updates), fusion del carrito de invitado al hacer login, y limpieza automatica del carrito al completar checkout.

---

## Estructura de datos en Firestore

**Coleccion:** `carts`
**Documento ID:** `{user_id}` (un documento por usuario)

```json
{
  "user_id": "uuid-string",
  "items": [
    {
      "product_id": "uuid-string",
      "product_name": "Cafe Colombiano 250g",
      "product_price": 25000,
      "product_image": "https://...",
      "quantity": 2
    }
  ],
  "updated_at": "2026-02-12T..."
}
```

Decisiones de diseno:
- Un documento por usuario (no una subcoleccion) porque el carrito es pequeno y se lee/escribe entero.
- Se guardan `product_name`, `product_price` y `product_image` denormalizados en cada item para evitar consultas extra al cargar el carrito. Los precios se revalidan al momento del checkout (ya existe esta logica en `order_service.go`).
- El `product_id` se usa como clave para identificar items de forma unica dentro del array.

---

## Implementation Steps

### Step 1: Backend - Modelo del carrito (Go)
**Files:**
- `GoBackend_Cheos/internal/models/cart.go` (CREAR)

**What:**
Crear el modelo `Cart` y `CartItem` siguiendo el patron existente de `models/order.go`. Incluir DTOs para las peticiones de la API.

**Detalles del modelo:**
```go
// Cart representa el carrito de un usuario en Firestore
type Cart struct {
    UserID    uuid.UUID  `json:"user_id" firestore:"user_id"`
    Items     []CartItem `json:"items" firestore:"items"`
    UpdatedAt time.Time  `json:"updated_at" firestore:"updated_at"`
}

type CartItem struct {
    ProductID    uuid.UUID `json:"product_id" firestore:"product_id"`
    ProductName  string    `json:"product_name" firestore:"product_name"`
    ProductPrice float64   `json:"product_price" firestore:"product_price"`
    ProductImage string    `json:"product_image" firestore:"product_image"`
    Quantity     int       `json:"quantity" firestore:"quantity"`
}

// DTOs
type AddToCartRequest struct {
    ProductID uuid.UUID `json:"product_id" validate:"required"`
    Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type UpdateCartItemRequest struct {
    Quantity int `json:"quantity" validate:"required,gt=0"`
}

type SyncCartRequest struct {
    Items []AddToCartRequest `json:"items" validate:"required"`
}
```

**Testing:** Compilar el proyecto (`go build ./...`) sin errores.

---

### Step 2: Backend - Repositorio del carrito (Firebase)
**Files:**
- `GoBackend_Cheos/internal/repository/cart_repository.go` (CREAR)

**What:**
Crear `CartRepository` siguiendo el patron de `site_config_repository.go` (documento unico por clave) y `product_repository.go` (operaciones CRUD con Firebase). Usar el `user_id` como Document ID para acceso directo O(1).

**Metodos a implementar:**
- `NewCartRepository(firebase *database.FirebaseClient) *CartRepository`
- `GetByUserID(ctx, userID uuid.UUID) (*models.Cart, error)` -- Obtener el carrito de un usuario. Si no existe, retornar carrito vacio (no error). Usar patron de `site_config_repository.go` con verificacion `codes.NotFound`.
- `Save(ctx, cart *models.Cart) error` -- Guardar/sobreescribir el carrito completo. Usar `Set()` como en `site_config_repository.go`.
- `Delete(ctx, userID uuid.UUID) error` -- Eliminar el carrito (para despues del checkout). Usar `Delete()` como en `gallery_repository.go`.

**Patron clave de acceso al documento:**
```go
r.firebase.Collection("carts").Doc(userID.String())
```

**Testing:** Compilar el proyecto sin errores.

---

### Step 3: Backend - Servicio del carrito (logica de negocio)
**Files:**
- `GoBackend_Cheos/internal/services/cart_service.go` (CREAR)

**What:**
Crear `CartService` siguiendo el patron de `order_service.go` (que inyecta `productRepo` para validar productos). El servicio encapsula toda la logica de negocio del carrito.

**Constructor:**
```go
type CartService struct {
    cartRepo    *repository.CartRepository
    productRepo *repository.ProductRepository
}
func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService
```

**Metodos a implementar:**

1. `GetCart(ctx, userID uuid.UUID) (*models.Cart, error)`
   - Delega a `cartRepo.GetByUserID()`.

2. `AddItem(ctx, userID uuid.UUID, req *models.AddToCartRequest) (*models.Cart, error)`
   - Validar que el producto existe y esta activo (`productRepo.GetByID()`, verificar `IsActive`).
   - Obtener carrito actual (o crear uno vacio si no existe).
   - Si el producto ya esta en el carrito, sumar la cantidad.
   - Si es nuevo, agregar al array con datos denormalizados del producto (`Name`, `Price`, `Images[0]`).
   - Guardar y retornar el carrito actualizado.

3. `UpdateItemQuantity(ctx, userID uuid.UUID, productID uuid.UUID, req *models.UpdateCartItemRequest) (*models.Cart, error)`
   - Obtener carrito, buscar el item por `product_id`.
   - Si no existe, retornar error.
   - Actualizar la cantidad.
   - Guardar y retornar el carrito actualizado.

4. `RemoveItem(ctx, userID uuid.UUID, productID uuid.UUID) (*models.Cart, error)`
   - Obtener carrito, filtrar el item con ese `product_id`.
   - Guardar y retornar el carrito actualizado.

5. `ClearCart(ctx, userID uuid.UUID) error`
   - Delega a `cartRepo.Delete()`.

6. `SyncCart(ctx, userID uuid.UUID, req *models.SyncCartRequest) (*models.Cart, error)`
   - Para la fusion al hacer login: recibe los items del carrito local (invitado).
   - Obtiene el carrito guardado del usuario.
   - Fusiona: para cada item del request, si ya existe en el carrito guardado, sumar cantidades; si no existe, agregarlo (validando que el producto existe y esta activo).
   - Guardar y retornar el carrito fusionado.

**Testing:** Compilar el proyecto sin errores.

---

### Step 4: Backend - Handler del carrito (endpoints HTTP)
**Files:**
- `GoBackend_Cheos/internal/handlers/cart_handler.go` (CREAR)

**What:**
Crear `CartHandler` siguiendo el patron de `gallery_handler.go` (con validacion via `utils.ValidateStruct` y respuestas via `utils.SuccessResponse`/`utils.ErrorResponse`). Todos los handlers extraen `user_id` del contexto Gin (puesto por el middleware de auth).

**Constructor:**
```go
type CartHandler struct {
    cartService *services.CartService
}
func NewCartHandler(cartService *services.CartService) *CartHandler
```

**Endpoints a implementar:**

1. `GetCart(c *gin.Context)` -- `GET /api/v1/cart`
   - Extraer `user_id` del contexto (patron de `order_handler.go` linea 83-94, usando cast a `uuid.UUID`).
   - Llamar `cartService.GetCart()`.
   - Responder con el carrito.

2. `AddItem(c *gin.Context)` -- `POST /api/v1/cart/items`
   - Extraer `user_id`, parsear body (`AddToCartRequest`), validar, llamar `cartService.AddItem()`.
   - Responder 200 con el carrito actualizado.

3. `UpdateItemQuantity(c *gin.Context)` -- `PUT /api/v1/cart/items/:productId`
   - Extraer `user_id`, parsear `productId` del path, parsear body (`UpdateCartItemRequest`), validar, llamar `cartService.UpdateItemQuantity()`.
   - Responder 200 con el carrito actualizado.

4. `RemoveItem(c *gin.Context)` -- `DELETE /api/v1/cart/items/:productId`
   - Extraer `user_id`, parsear `productId` del path, llamar `cartService.RemoveItem()`.
   - Responder 200 con el carrito actualizado.

5. `ClearCart(c *gin.Context)` -- `DELETE /api/v1/cart`
   - Extraer `user_id`, llamar `cartService.ClearCart()`.
   - Responder 200 con mensaje de exito.

6. `SyncCart(c *gin.Context)` -- `POST /api/v1/cart/sync`
   - Extraer `user_id`, parsear body (`SyncCartRequest`), validar, llamar `cartService.SyncCart()`.
   - Responder 200 con el carrito fusionado.

**Nota sobre extraccion de user_id:** Usar el patron correcto de `GetUserOrders` (cast directo a `uuid.UUID`) ya que el middleware almacena `claims.UserID` que es tipo `uuid.UUID`:
```go
userIDInterface, exists := c.Get("user_id")
if !exists {
    utils.ErrorResponse(c, http.StatusUnauthorized, "No autenticado", nil)
    return
}
userID, ok := userIDInterface.(uuid.UUID)
if !ok {
    utils.ErrorResponse(c, http.StatusBadRequest, "ID de usuario invalido", nil)
    return
}
```

**Testing:** Compilar el proyecto sin errores.

---

### Step 5: Backend - Registro de rutas e inyeccion de dependencias
**Files:**
- `GoBackend_Cheos/cmd/api/main.go` (MODIFICAR)

**What:**
Registrar el nuevo modulo de carrito en el flujo de inyeccion de dependencias y en las rutas, siguiendo el patron exacto de los modulos existentes.

**Cambios especificos:**

1. **Inicializacion del repositorio** (despues de linea 72, `siteConfigRepo`):
   ```go
   cartRepo := repository.NewCartRepository(firebaseClient)
   ```

2. **Inicializacion del servicio** (despues de linea 82, `siteConfigService`):
   ```go
   cartService := services.NewCartService(cartRepo, productRepo)
   ```

3. **Inicializacion del handler** (despues de linea 99, `siteConfigHandler`):
   ```go
   cartHandler := handlers.NewCartHandler(cartService)
   ```

4. **Agregar `cartHandler` a la firma de `setupRoutes`** y a la llamada en linea 112.

5. **Registrar rutas** (dentro de `setupRoutes`, despues del bloque `siteConfig`):
   ```go
   // Cart routes (authenticated users only)
   cart := v1.Group("/cart")
   cart.Use(middleware.AuthMiddleware(cfg))
   {
       cart.GET("", cartHandler.GetCart)
       cart.POST("/items", cartHandler.AddItem)
       cart.PUT("/items/:productId", cartHandler.UpdateItemQuantity)
       cart.DELETE("/items/:productId", cartHandler.RemoveItem)
       cart.DELETE("", cartHandler.ClearCart)
       cart.POST("/sync", cartHandler.SyncCart)
   }
   ```

**Testing:** Compilar el proyecto (`go build ./cmd/api`). Verificar que el servidor arranca y las rutas se registran en los logs.

---

### Step 6: Backend - Vaciar carrito automaticamente al crear orden
**Files:**
- `GoBackend_Cheos/internal/services/order_service.go` (MODIFICAR)

**What:**
Modificar `OrderService` para que al crear una orden exitosamente, si el usuario esta autenticado (`userID != nil`), se vacie su carrito automaticamente.

**Cambios especificos:**

1. **Agregar `cartRepo` al struct y constructor:**
   ```go
   type OrderService struct {
       orderRepo   *repository.OrderRepository
       productRepo *repository.ProductRepository
       cartRepo    *repository.CartRepository  // AGREGAR
   }

   func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, cartRepo *repository.CartRepository) *OrderService {
       return &OrderService{
           orderRepo:   orderRepo,
           productRepo: productRepo,
           cartRepo:    cartRepo,  // AGREGAR
       }
   }
   ```

2. **En `CreateOrder`, despues de guardar los items exitosamente (despues de linea 113):**
   ```go
   // Vaciar carrito del usuario despues de crear la orden
   if userID != nil {
       _ = s.cartRepo.Delete(ctx, *userID) // Ignorar error, no es critico
   }
   ```

3. **Actualizar la llamada en `main.go`** (linea 77):
   ```go
   orderService := services.NewOrderService(orderRepo, productRepo, cartRepo)
   ```

**Testing:** Compilar el proyecto. Verificar que crear una orden sigue funcionando correctamente.

---

### Step 7: Frontend - Servicio API para el carrito
**Files:**
- `ReactFront_Cheos/src/routes/cart.js` (CREAR)

**What:**
Crear un modulo de funciones API para comunicarse con los endpoints del carrito, siguiendo el patron de `routes/orders.js` y `routes/products.js`.

**Funciones a implementar:**
```javascript
const API = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

export async function getCart(token) {
  const res = await fetch(`${API}/cart`, {
    headers: { 'Authorization': `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('Error al obtener carrito')
  return res.json()
}

export async function addCartItem(token, productId, quantity) {
  const res = await fetch(`${API}/cart/items`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ product_id: productId, quantity })
  })
  if (!res.ok) throw new Error('Error al agregar al carrito')
  return res.json()
}

export async function updateCartItemQuantity(token, productId, quantity) {
  const res = await fetch(`${API}/cart/items/${productId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ quantity })
  })
  if (!res.ok) throw new Error('Error al actualizar cantidad')
  return res.json()
}

export async function removeCartItem(token, productId) {
  const res = await fetch(`${API}/cart/items/${productId}`, {
    method: 'DELETE',
    headers: { 'Authorization': `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('Error al eliminar del carrito')
  return res.json()
}

export async function clearCart(token) {
  const res = await fetch(`${API}/cart`, {
    method: 'DELETE',
    headers: { 'Authorization': `Bearer ${token}` }
  })
  if (!res.ok) throw new Error('Error al vaciar carrito')
  return res.json()
}

export async function syncCart(token, items) {
  const res = await fetch(`${API}/cart/sync`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ items })
  })
  if (!res.ok) throw new Error('Error al sincronizar carrito')
  return res.json()
}
```

**Testing:** Verificar que el archivo no tiene errores de sintaxis con `npx eslint src/routes/cart.js` (si hay eslint configurado) o que Vite no muestra errores.

---

### Step 8: Frontend - Refactorizar CartContext con sincronizacion API
**Files:**
- `ReactFront_Cheos/src/context/CartContext.jsx` (MODIFICAR COMPLETAMENTE)

**What:**
Reescribir `CartContext.jsx` para que soporte dos modos de operacion:
1. **Modo invitado** (sin login): funciona igual que ahora, en memoria con `useState`.
2. **Modo autenticado** (con login): cada operacion actualiza el state local inmediatamente (optimistic update) y luego sincroniza con la API en background. Si la API falla, se muestra un warning en consola pero NO se revierte el state (para no interrumpir la experiencia).

**Cambios clave:**

1. **Importar `useUser` de `UserContext`** para detectar si el usuario esta logueado y obtener el token.

2. **Importar funciones API** de `routes/cart.js`.

3. **Refactorizar `addToCart(product)`:**
   - Actualizar el state local inmediatamente (codigo actual).
   - Si `isLoggedIn` (existe `token` y `user`), llamar a `addCartItem(token, product.id, 1)` en background (sin `await` bloqueante, usando `.catch(console.error)`).

4. **Refactorizar `removeFromCart(id)`:**
   - Actualizar state local inmediatamente.
   - Si `isLoggedIn`, llamar a `removeCartItem(token, id)` en background.

5. **Refactorizar `updateQuantity(id, qty)`:**
   - Actualizar state local inmediatamente.
   - Si `isLoggedIn`, llamar a `updateCartItemQuantity(token, id, qty)` en background.

6. **Refactorizar `clearCart()`:**
   - Actualizar state local inmediatamente.
   - Si `isLoggedIn`, llamar a `clearCart(token)` en background.

7. **Agregar `loadCartFromAPI()`:**
   - Funcion interna que llama a `getCart(token)` y setea el state con los items recibidos.
   - Mapear la respuesta del API al formato que usa el state local: `{ id: item.product_id, name: item.product_name, price: item.product_price, image: item.product_image, quantity: item.quantity }`.

8. **Agregar `mergeAndSyncCart()`:**
   - Funcion que se ejecuta al hacer login.
   - Si el carrito local (invitado) tiene items, llamar a `syncCart(token, localItems)` para fusionarlos con los guardados.
   - Si el carrito local esta vacio, simplemente cargar el carrito guardado con `loadCartFromAPI()`.
   - En ambos casos, al final el state queda con el carrito fusionado que retorna la API.

9. **Agregar `useEffect` que reacciona al cambio de `user`/`token`:**
   ```javascript
   useEffect(() => {
     if (user && token) {
       // Usuario acaba de loguearse: fusionar carrito
       mergeAndSyncCart()
     } else if (!user && prevUserRef.current) {
       // Usuario acaba de cerrar sesion: limpiar state local
       setCart([])
     }
     prevUserRef.current = user
   }, [user, token])
   ```
   Usar un `useRef` para trackear el estado previo del usuario y detectar transiciones login/logout.

**Nota importante sobre el provider hierarchy:**
En `App.jsx`, `CartProvider` esta dentro de `UserProvider`, por lo tanto `CartContext` puede usar `useUser()` sin problemas. No hay que cambiar la jerarquia.

**Testing:** Verificar manualmente en el navegador:
- Como invitado: agregar/quitar productos funciona igual que antes.
- Loguearse: el carrito del invitado se fusiona con el guardado.
- Como usuario logueado: agregar/quitar productos se refleja instantaneamente en la UI.
- Refrescar la pagina estando logueado: el carrito persiste.
- Cerrar sesion: el carrito se vacia localmente.

---

### Step 9: Frontend - Actualizar CartDrawer para pasar token al checkout
**Files:**
- `ReactFront_Cheos/src/components/CartDrawer.jsx` (MODIFICAR)

**What:**
Modificar el componente `CartDrawer` para que al hacer checkout envie el token de autenticacion si el usuario esta logueado. Esto permite que `order_service.go` identifique al usuario y vacie su carrito automaticamente.

**Cambios especificos:**

1. **Importar `useUser`:**
   ```javascript
   import { useUser } from '../context/UserContext'
   ```

2. **Obtener `token` y `user` del contexto:**
   ```javascript
   const { token, user } = useUser()
   ```

3. **Modificar la funcion `checkout()`:**
   - Si el usuario esta logueado, agregar datos del usuario al payload (`customer_name: user.name`, `customer_email: user.email`).
   - Pasar el `token` a `createOrder(payload, token)` (la funcion en `routes/orders.js` ya acepta el parametro `token`, ver linea 2-6).

4. **Ya no es necesario llamar `clearCart()` manualmente tras checkout si el usuario esta logueado**, porque el backend lo hace automaticamente. Pero se sigue llamando para limpiar el state local (el `clearCart` de `CartContext` ahora tambien limpia el state local).

**Testing:** Verificar que el checkout funciona tanto como invitado como logueado. Verificar que tras checkout el carrito queda vacio.

---

## Resumen de archivos

### Archivos NUEVOS (4):
| Archivo | Descripcion |
|---------|-------------|
| `GoBackend_Cheos/internal/models/cart.go` | Modelo Cart, CartItem y DTOs |
| `GoBackend_Cheos/internal/repository/cart_repository.go` | Repositorio Firestore para carrito |
| `GoBackend_Cheos/internal/services/cart_service.go` | Logica de negocio del carrito |
| `GoBackend_Cheos/internal/handlers/cart_handler.go` | Endpoints HTTP del carrito |
| `ReactFront_Cheos/src/routes/cart.js` | Funciones API del frontend |

### Archivos MODIFICADOS (3):
| Archivo | Cambio |
|---------|--------|
| `GoBackend_Cheos/cmd/api/main.go` | Inyeccion de dependencias + rutas del carrito |
| `GoBackend_Cheos/internal/services/order_service.go` | Vaciar carrito al crear orden |
| `ReactFront_Cheos/src/context/CartContext.jsx` | Sincronizacion con API + fusion al login |
| `ReactFront_Cheos/src/components/CartDrawer.jsx` | Pasar token al checkout |

---

## Endpoints API resultantes

| Metodo | Ruta | Auth | Descripcion |
|--------|------|------|-------------|
| `GET` | `/api/v1/cart` | Si | Obtener carrito del usuario |
| `POST` | `/api/v1/cart/items` | Si | Agregar item al carrito |
| `PUT` | `/api/v1/cart/items/:productId` | Si | Actualizar cantidad de un item |
| `DELETE` | `/api/v1/cart/items/:productId` | Si | Eliminar item del carrito |
| `DELETE` | `/api/v1/cart` | Si | Vaciar carrito completo |
| `POST` | `/api/v1/cart/sync` | Si | Fusionar carrito de invitado con guardado |

---

## Notas tecnicas

1. **Seguridad:** Todos los endpoints del carrito requieren autenticacion. El `user_id` se extrae del token JWT (via middleware), nunca del body o query params. Un usuario solo puede acceder a su propio carrito.

2. **Optimistic updates:** El frontend actualiza la UI inmediatamente sin esperar la respuesta del servidor. Las llamadas a la API van en background con `.catch(console.error)`. Esto da una experiencia instantanea.

3. **Consistencia eventual:** Puede haber un breve periodo donde el state local y Firestore no coinciden (si la API falla silenciosamente). Esto es aceptable porque al recargar la pagina se carga el estado de Firestore.

4. **Fusion al login:** El endpoint `POST /cart/sync` resuelve el caso donde un usuario agrego productos como invitado y luego se loguea. El backend fusiona sumando cantidades de productos repetidos.

5. **Denormalizacion de datos del producto:** Se guardan nombre, precio e imagen del producto dentro de cada item del carrito. Esto evita N+1 queries al cargar el carrito. Los precios se revalidan al hacer checkout (logica existente en `order_service.go`).

6. **Incompatibilidad detectada en `order_handler.go`:** El handler `CreateOrder` hace `userIDStr.(string)` pero el middleware almacena `claims.UserID` como `uuid.UUID`. El handler del carrito usara el patron correcto de `GetUserOrders` (cast a `uuid.UUID`). Se recomienda corregir `CreateOrder` en un commit separado si se desea, pero no es parte del scope de este PR.
