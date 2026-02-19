# Contexto: Servicio de Dashboard (Metricas y Estadisticas)

## Enfoque: Event-Driven (Metricas Pre-calculadas)

Cada vez que se crea o actualiza una orden, se actualizan documentos de metricas en Firestore. El dashboard solo lee estos documentos pre-calculados, lo que garantiza respuestas instantaneas sin necesidad de recorrer todas las ordenes.

---

## Datos disponibles para cruzar

### Orden (orders)
- total, subtotal, discount
- status: PENDING, CONFIRMED, PROCESSING, SHIPPED, DELIVERED, CANCELLED
- payment_method: MERCADO_PAGO, CONTRA_ENTREGA
- payment_status: PENDING, APPROVED, REJECTED, REFUNDED
- customer_name, customer_email, customer_phone
- user_id (null si es invitado)
- shipping_address.city, shipping_address.department
- discount_code_id (si uso cupon)
- utm_source, utm_medium, utm_campaign
- created_at

### Items de orden (order_items)
- product_id, product_name
- quantity, price, subtotal

### Usuario (users)
- city, municipality, neighborhood (opcionales)
- gender: MALE, FEMALE, OTHER (opcional)
- birth_date (opcional, formato "YYYY-MM-DD")
- created_at (fecha de registro)

---

## Coleccion Firestore: dashboard_metrics

Todos los documentos de metricas se almacenan en una sola coleccion llamada `dashboard_metrics`. Cada documento tiene un ID descriptivo que indica que metrica contiene.

### 1. Ventas Mensuales

**Documento:** `sales_monthly_YYYY-MM` (ej: `sales_monthly_2026-02`)

```json
{
  "year": 2026,
  "month": 2,
  "total_revenue": 4500000,
  "total_orders": 120,
  "completed_orders": 95,
  "cancelled_orders": 8,
  "pending_orders": 17,
  "average_ticket": 37500,
  "total_discount_given": 350000,
  "orders_with_discount": 25,
  "payment_methods": {
    "MERCADO_PAGO": { "count": 70, "total": 2800000 },
    "CONTRA_ENTREGA": { "count": 50, "total": 1700000 }
  },
  "daily_breakdown": {
    "2026-02-01": { "revenue": 180000, "orders": 5 },
    "2026-02-02": { "revenue": 220000, "orders": 7 }
  },
  "updated_at": "2026-02-18T20:00:00Z"
}
```

**Se actualiza cuando:**
- Se crea una nueva orden (incrementa total_revenue, total_orders, daily_breakdown del dia)
- Se cambia el estado de una orden (mueve contadores entre completed/cancelled/pending)
- Se cambia el estado de pago

---

### 2. Ventas Anuales

**Documento:** `sales_yearly_YYYY` (ej: `sales_yearly_2026`)

```json
{
  "year": 2026,
  "total_revenue": 48000000,
  "total_orders": 1350,
  "completed_orders": 1100,
  "cancelled_orders": 85,
  "average_ticket": 35555,
  "monthly_breakdown": {
    "01": { "revenue": 3200000, "orders": 87 },
    "02": { "revenue": 4500000, "orders": 120 }
  },
  "updated_at": "2026-02-18T20:00:00Z"
}
```

**Se actualiza cuando:**
- Se crea una nueva orden
- Se cambia el estado de una orden

---

### 3. Estadisticas de Compradores Mensuales

**Documento:** `buyers_monthly_YYYY-MM` (ej: `buyers_monthly_2026-02`)

```json
{
  "year": 2026,
  "month": 2,
  "total_buyers": 85,
  "registered_buyers": 60,
  "guest_buyers": 25,
  "new_registered_this_month": 18,
  "returning_buyers": 42,
  "gender_breakdown": {
    "MALE": 35,
    "FEMALE": 40,
    "OTHER": 3,
    "UNKNOWN": 7
  },
  "age_breakdown": {
    "18-24": 12,
    "25-34": 30,
    "35-44": 22,
    "45-54": 13,
    "55+": 5,
    "unknown": 3
  },
  "city_breakdown": {
    "Medellin": 45,
    "Envigado": 15,
    "Itagui": 10,
    "other": 15
  },
  "buyer_ids": ["uuid1", "uuid2"],
  "updated_at": "2026-02-18T20:00:00Z"
}
```

**Nota sobre `buyer_ids`:** Se guarda un array de user_ids (o emails para invitados) que ya compraron en el mes. Esto permite saber si un comprador es nuevo o recurrente sin recorrer todas las ordenes. Si el array crece mucho (>500), se puede mover a un subdocumento.

**Se actualiza cuando:**
- Se crea una nueva orden (se agrega el buyer si no existe, se incrementan contadores)
- Se consulta el perfil del comprador para obtener gender/birth_date/city

---

### 4. Estadisticas de Compradores Anuales

**Documento:** `buyers_yearly_YYYY` (ej: `buyers_yearly_2026`)

```json
{
  "year": 2026,
  "total_unique_buyers": 450,
  "registered_buyers": 320,
  "guest_buyers": 130,
  "gender_breakdown": {
    "MALE": 180,
    "FEMALE": 210,
    "OTHER": 15,
    "UNKNOWN": 45
  },
  "age_breakdown": {
    "18-24": 65,
    "25-34": 150,
    "35-44": 120,
    "45-54": 70,
    "55+": 25,
    "unknown": 20
  },
  "buyer_ids": ["uuid1", "uuid2"],
  "updated_at": "2026-02-18T20:00:00Z"
}
```

**Se actualiza cuando:**
- Se crea una nueva orden (misma logica que el mensual pero a nivel anual)

---

### 5. Top Productos

**Documento:** `top_products_monthly_YYYY-MM` (ej: `top_products_monthly_2026-02`)

```json
{
  "year": 2026,
  "month": 2,
  "most_sold": [
    {
      "product_id": "uuid",
      "product_name": "Cheo's Cafe 500g",
      "total_quantity": 85,
      "total_revenue": 2975000,
      "order_count": 60
    }
  ],
  "least_sold": [
    {
      "product_id": "uuid",
      "product_name": "Taza Cheos Edicion Limitada",
      "total_quantity": 2,
      "total_revenue": 50000,
      "order_count": 2
    }
  ],
  "all_products": {
    "product-uuid-1": { "name": "Cheo's Cafe 500g", "quantity": 85, "revenue": 2975000, "orders": 60 },
    "product-uuid-2": { "name": "Cafe Molido 250g", "quantity": 45, "revenue": 1125000, "orders": 38 }
  },
  "updated_at": "2026-02-18T20:00:00Z"
}
```

**Nota:** `all_products` es un mapa con todos los productos vendidos en el periodo. Los arrays `most_sold` y `least_sold` se recalculan en el servicio al momento de actualizar (se ordenan los items de `all_products` y se toman los top/bottom 10).

**Se actualiza cuando:**
- Se crea una nueva orden (se recorren los items y se incrementan cantidades por producto)

**Documento:** `top_products_yearly_YYYY` - misma estructura pero acumulado anual.

---

## Flujo de actualizacion (Event-Driven)

```
Usuario crea orden
       |
       v
OrderService.CreateOrder()
       |
       v
(1) Se guarda la orden en 'orders'
(2) Se guardan los items en 'order_items'
(3) Se llama a DashboardService.OnOrderCreated(order, items)
       |
       v
DashboardService actualiza en paralelo:
  - sales_monthly_YYYY-MM
  - sales_yearly_YYYY
  - buyers_monthly_YYYY-MM
  - buyers_yearly_YYYY
  - top_products_monthly_YYYY-MM
  - top_products_yearly_YYYY
```

```
Admin cambia estado de orden (ej: CANCELLED)
       |
       v
OrderService.UpdateStatus()
       |
       v
(1) Se actualiza la orden
(2) Se llama a DashboardService.OnOrderStatusChanged(order, oldStatus, newStatus)
       |
       v
DashboardService actualiza:
  - sales_monthly_YYYY-MM (mueve contadores de estado)
  - sales_yearly_YYYY (mueve contadores de estado)
```

---

## Endpoints del Dashboard (solo lectura)

Todos requieren autenticacion de ADMIN.

| Metodo | Ruta | Descripcion |
|--------|------|-------------|
| GET | /api/v1/dashboard/sales/monthly?year=2026&month=2 | Ventas del mes |
| GET | /api/v1/dashboard/sales/yearly?year=2026 | Ventas del año |
| GET | /api/v1/dashboard/buyers/monthly?year=2026&month=2 | Stats compradores del mes |
| GET | /api/v1/dashboard/buyers/yearly?year=2026 | Stats compradores del año |
| GET | /api/v1/dashboard/products/monthly?year=2026&month=2 | Top productos del mes |
| GET | /api/v1/dashboard/products/yearly?year=2026 | Top productos del año |
| GET | /api/v1/dashboard/summary | Resumen general (lee el mes y año actual) |

### Endpoint Summary

El endpoint `/summary` retorna un objeto consolidado con los datos mas relevantes para la vista principal del dashboard:

```json
{
  "current_month": {
    "revenue": 4500000,
    "orders": 120,
    "average_ticket": 37500,
    "new_buyers": 18
  },
  "current_year": {
    "revenue": 48000000,
    "orders": 1350
  },
  "top_products": [
    { "name": "Cheo's Cafe 500g", "quantity": 85, "revenue": 2975000 }
  ],
  "gender_breakdown": { "MALE": 35, "FEMALE": 40, "OTHER": 3, "UNKNOWN": 7 },
  "age_breakdown": { "18-24": 12, "25-34": 30, "35-44": 22 }
}
```

---

## Archivos a crear

| Archivo | Descripcion |
|---------|-------------|
| internal/models/dashboard.go | Structs de metricas (SalesMetrics, BuyerMetrics, ProductMetrics, etc.) |
| internal/repository/dashboard_repository.go | CRUD de documentos en dashboard_metrics |
| internal/services/dashboard_service.go | Logica de actualizacion de metricas + lectura |
| internal/handlers/dashboard_handler.go | Endpoints GET del dashboard |

## Archivos a modificar

| Archivo | Cambio |
|---------|--------|
| internal/services/order_service.go | Llamar a DashboardService.OnOrderCreated() y OnOrderStatusChanged() |
| cmd/api/main.go | Inyectar dashboard repo/service/handler + registrar rutas |

---

## Calculo de rangos de edad

A partir del campo `birth_date` del usuario (formato "YYYY-MM-DD"), se calcula la edad actual y se clasifica en rangos:

- 18-24
- 25-34
- 35-44
- 45-54
- 55+
- unknown (si birth_date es null)

---

## Endpoint de recalculo manual (Admin)

Para corregir inconsistencias o regenerar metricas de meses pasados:

| Metodo | Ruta | Descripcion |
|--------|------|-------------|
| POST | /api/v1/dashboard/recalculate?year=2026&month=2 | Recalcula metricas del mes desde cero |

Este endpoint recorre todas las ordenes del periodo, recalcula todas las metricas y sobreescribe los documentos. Es un proceso pesado, solo para uso ocasional del admin.

---

## Notas importantes

- Las metricas se actualizan de forma **sincrona** dentro de la misma transaccion de la orden. Si falla la actualizacion de metricas, se loguea el error pero NO se revierte la orden (la orden tiene prioridad).
- El campo `buyer_ids` puede crecer. Si supera 500 entries en un mes, considerar moverlo a un subdocumento o eliminarlo y usar solo contadores.
- Los montos siempre se manejan en COP (pesos colombianos) sin decimales significativos.
- `most_sold` y `least_sold` guardan los top 10 de cada uno.
- Este servicio es SOLO backend. El frontend consumira estos endpoints con una libreria de graficos (Chart.js, Recharts, etc.) pero eso lo maneja otro equipo.
