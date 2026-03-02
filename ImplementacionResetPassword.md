# Plan de Implementacion: Reset de Contraseña por Email

## Resumen

Sistema de restablecimiento de contraseña por correo electronico para Cheos Cafe. El usuario ingresa su email, recibe un correo con un enlace unico, y al hacer clic puede crear una nueva contraseña.

---

## Datos de Configuracion

- **Email remitente:** cheoscafe4@gmail.com
- **App Password Gmail:** lxrp zhki yrgc xthr
- **Frontend URL:** https://cheoscafe.netlify.app
- **Libreria Go para emails:** `net/smtp` (nativo de Go, no requiere dependencias extra)

---

## Flujo Completo

```
1. Usuario hace clic en "Olvide mi contraseña"
2. Frontend muestra formulario para ingresar email
3. Frontend envia POST /api/v1/auth/forgot-password con { "email": "..." }
4. Backend:
   a. Busca el usuario por email
   b. Genera un token JWT de reset (expira en 15 minutos)
   c. Guarda el token en Firestore (coleccion password_resets)
   d. Envia email con enlace: https://cheoscafe.netlify.app/reset-password?token=TOKEN
   e. Responde "Si el email existe, se envio un correo" (siempre, por seguridad)
5. Usuario abre su correo y hace clic en el enlace
6. Frontend lee el token de la URL y muestra formulario de nueva contraseña
7. Frontend envia POST /api/v1/auth/reset-password con { "token": "...", "new_password": "..." }
8. Backend:
   a. Valida el token JWT
   b. Verifica que el token existe en Firestore y no fue usado
   c. Hashea la nueva contraseña
   d. Actualiza la contraseña del usuario en Firestore
   e. Elimina el token de password_resets (hard delete)
   f. Responde "Contraseña actualizada exitosamente"
```

---

## Endpoints Nuevos (2)

| Metodo | Ruta | Auth | Descripcion |
|--------|------|------|-------------|
| POST | /api/v1/auth/forgot-password | Publico | Solicitar reset de contraseña |
| POST | /api/v1/auth/reset-password | Publico | Confirmar nueva contraseña |

---

## Coleccion Firestore: password_resets

Cada documento representa un token de reset activo.

```json
{
  "id": "uuid",
  "user_id": "uuid del usuario",
  "email": "correo@ejemplo.com",
  "token": "jwt-token-completo",
  "used": false,
  "created_at": "2026-02-18T20:00:00Z",
  "expires_at": "2026-02-18T20:15:00Z"
}
```

- Se usa el `id` (UUID) como document ID
- `expires_at` = created_at + 15 minutos
- `used` = true cuando se usa (luego se elimina)
- Si un usuario pide reset varias veces, se eliminan los tokens anteriores y se crea uno nuevo

---

## Variables de Entorno Nuevas

Agregar al `.env`:
```
# EMAIL (Gmail SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=cheoscafe4@gmail.com
SMTP_PASSWORD=lxrp zhki yrgc xthr
```

Agregar al `.env` en Render (produccion) las mismas 4 variables.

---

## Archivos a Crear

### 1. internal/models/password_reset.go

Modelo del token de reset:

```go
type PasswordReset struct {
    ID        uuid.UUID `json:"id" firestore:"id"`
    UserID    uuid.UUID `json:"user_id" firestore:"user_id"`
    Email     string    `json:"email" firestore:"email"`
    Token     string    `json:"token" firestore:"token"`
    Used      bool      `json:"used" firestore:"used"`
    CreatedAt time.Time `json:"created_at" firestore:"created_at"`
    ExpiresAt time.Time `json:"expires_at" firestore:"expires_at"`
}

type ForgotPasswordRequest struct {
    Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
    Token       string `json:"token" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=6"`
}
```

### 2. internal/repository/password_reset_repository.go

Metodos:
- `Create(ctx, resetDoc)` - Guardar token de reset
- `GetByToken(ctx, token)` - Buscar por token JWT
- `DeleteByUserID(ctx, userID)` - Eliminar tokens previos del usuario
- `Delete(ctx, id)` - Eliminar token especifico (hard delete)

### 3. internal/services/email_service.go

Servicio de envio de emails usando SMTP de Gmail:
- `NewEmailService(host, port, email, password, frontendURL)` - Constructor
- `SendPasswordResetEmail(toEmail, toName, resetToken)` - Envia el correo con el enlace

El correo debe:
- Tener asunto: "Cheos Cafe - Restablecer contraseña"
- Incluir el nombre del usuario
- Tener un boton/enlace a: `{FRONTEND_URL}/reset-password?token={TOKEN}`
- Indicar que expira en 15 minutos
- Ser HTML con estilo basico (colores cafe, logo)

### 4. internal/services/password_reset_service.go

Logica de negocio:
- `ForgotPassword(ctx, email)`:
  1. Buscar usuario por email (si no existe, retornar nil sin error por seguridad)
  2. Eliminar tokens previos del usuario
  3. Generar token JWT con claims: user_id, email, tipo "reset" (expira 15 min)
  4. Guardar en Firestore
  5. Enviar email en goroutine (no bloquear respuesta)

- `ResetPassword(ctx, token, newPassword)`:
  1. Validar token JWT
  2. Buscar token en Firestore
  3. Verificar que no este usado y no haya expirado
  4. Buscar usuario por ID
  5. Hashear nueva contraseña
  6. Actualizar contraseña del usuario
  7. Eliminar token de Firestore

---

## Archivos a Modificar

### 5. internal/config/config.go

Agregar al struct Config:
```go
// SMTP Email
SMTPHost     string
SMTPPort     string
SMTPEmail    string
SMTPPassword string
```

En LoadConfig():
```go
SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
SMTPPort:     getEnv("SMTP_PORT", "587"),
SMTPEmail:    getEnv("SMTP_EMAIL", ""),
SMTPPassword: getEnv("SMTP_PASSWORD", ""),
```

### 6. cmd/api/main.go

Agregar:
- Inicializar `emailService` con config SMTP
- Inicializar `passwordResetRepo`
- Inicializar `passwordResetService`
- Registrar 2 rutas nuevas en el grupo `/auth`:
  ```go
  auth.POST("/forgot-password", authHandler.ForgotPassword)
  auth.POST("/reset-password", authHandler.ResetPassword)
  ```

### 7. internal/handlers/auth_handler.go

Agregar 2 handlers:
- `ForgotPassword(c *gin.Context)` - Recibe email, llama al servicio
- `ResetPassword(c *gin.Context)` - Recibe token + new_password, llama al servicio

Ambos retornan mensajes genericos por seguridad (no revelar si el email existe o no).

---

## Consideraciones de Seguridad

1. **No revelar si el email existe:** El endpoint forgot-password SIEMPRE responde "Si el correo esta registrado, recibiras un email" aunque el email no exista. Esto evita enumeracion de usuarios.

2. **Token de un solo uso:** Cada token solo se puede usar una vez. Despues se elimina de Firestore.

3. **Expiracion corta:** 15 minutos. Suficiente para que el usuario revise su correo pero no tanto como para que un atacante tenga tiempo.

4. **Rate limiting:** El endpoint forgot-password esta protegido por el rate limiter global (100 req/15min). Si se necesita, se puede agregar un rate limiter especifico mas estricto.

5. **Tokens previos se invalidan:** Al pedir un nuevo reset, se eliminan todos los tokens anteriores de ese usuario.

6. **JWT firmado:** El token de reset se firma con JWT_SECRET, no se puede falsificar.

---

## Template del Email (HTML)

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f5f0eb; padding: 20px;">
  <div style="max-width: 500px; margin: 0 auto; background: white; border-radius: 12px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
    <h1 style="color: #6F4E37; text-align: center;">Cheos Cafe</h1>
    <p>Hola <strong>{NOMBRE}</strong>,</p>
    <p>Recibimos una solicitud para restablecer la contraseña de tu cuenta.</p>
    <p>Haz clic en el siguiente boton para crear una nueva contraseña:</p>
    <div style="text-align: center; margin: 25px 0;">
      <a href="{RESET_URL}" style="background-color: #6F4E37; color: white; padding: 12px 30px; text-decoration: none; border-radius: 8px; font-size: 16px;">
        Restablecer contraseña
      </a>
    </div>
    <p style="color: #888; font-size: 13px;">Este enlace expira en 15 minutos.</p>
    <p style="color: #888; font-size: 13px;">Si no solicitaste este cambio, puedes ignorar este correo. Tu contraseña no sera modificada.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
    <p style="color: #aaa; font-size: 11px; text-align: center;">Cheos Cafe - Cafe de especialidad colombiano</p>
  </div>
</body>
</html>
```

---

## Orden de Implementacion

| Step | Accion | Archivo |
|------|--------|---------|
| 1 | Agregar variables SMTP a config | config.go + .env |
| 2 | Crear modelo PasswordReset | models/password_reset.go |
| 3 | Crear repositorio | repository/password_reset_repository.go |
| 4 | Crear servicio de email (SMTP Gmail) | services/email_service.go |
| 5 | Crear servicio de password reset | services/password_reset_service.go |
| 6 | Agregar handlers ForgotPassword y ResetPassword | handlers/auth_handler.go |
| 7 | Inyeccion de dependencias + rutas en main.go | cmd/api/main.go |
| 8 | Actualizar coleccion Postman | Cheos_Cafe_API.postman_collection.json |
| 9 | Agregar variables SMTP en Render (produccion) | Panel de Render |

---

## Para el Equipo de Frontend

El frontend debe crear una ruta `/reset-password` que:

1. Lea el query param `token` de la URL
2. Muestre un formulario con "Nueva contraseña" y "Confirmar contraseña"
3. Al enviar, haga POST a `/api/v1/auth/reset-password` con:
```json
{
  "token": "el-token-de-la-url",
  "new_password": "la-nueva-contraseña"
}
```
4. Si responde success, mostrar "Contraseña actualizada" y redirigir al login
5. Si responde error, mostrar el mensaje de error

---

Fecha: 2026-02-27
