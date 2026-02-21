# Sistema de Torneos y Apuestas - Backend

Backend API REST desarrollado en Go con el framework Gin para un sistema de apuestas y torneos deportivos.

## Características

- 🔐 **Autenticación JWT** - Login con email o username
- 👥 **Gestión de Usuarios** - Roles de usuario y administrador
- 🏆 **Sistema de Torneos** - Torneos con sesiones, eventos y pronósticos
- 💰 **Billetera Digital** - Depósitos, retiros y transacciones
- 💳 **Métodos de Pago** - Pago móvil, Zelle, Binance, PayPal, transferencia bancaria
- 📊 **Leaderboard** - Clasificación de participantes
- 📚 **Documentación API** - Swagger/OpenAPI

## Requisitos

- Go 1.21+
- PostgreSQL 14+
- Redis (opcional, para rate limiting)

## Instalación

1. **Clonar el repositorio:**
```bash
git clone <url-del-repositorio>
cd bets-backend
```

2. **Instalar dependencias:**
```bash
go mod download
go mod tidy
```

3. **Configurar variables de entorno:**

Crear archivo `.env`:
```env
# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=tu_password
DB_NAME=bets_system
DB_SSLMODE=disable

# JWT
JWT_SECRET=tu_secret_key_muy_segura

# Servidor
PORT=8080
GIN_MODE=debug
```

4. **Ejecutar migraciones:**
```bash
go run main.go
```

El servidor ejecutará automáticamente las migraciones al iniciar.

## Ejecución

### Desarrollo
```bash
go run main.go
```

### Producción
```bash
go build -o bin/server .
./bin/server
```

El servidor estará disponible en: `http://localhost:8080`

## Documentación API

Swagger UI disponible en: `http://localhost:8080/swagger/index.html`

## Estructura del Proyecto

```
bets-backend/
├── cmd/                    # Punto de entrada
├── config/                 # Configuración de base de datos
├── controllers/             # Controladores de la API
├── docs/                   # Documentación Swagger
├── dtos/                   # Objetos de transferencia de datos
├── middlewares/            # Middlewares de Gin
├── migrations/             # Migraciones de base de datos
├── models/                 # Modelos de GORM
├── routes/                 # Definición de rutas
├── tests/                  # Tests automatizados
├── utils/                 # Utilidades
├── main.go               # Archivo principal
└── .env                  # Variables de entorno
```

## Rutas API

### Autenticación
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Registro de usuario |
| POST | `/api/v1/auth/login` | Inicio de sesión |

### Torneos (Público)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/tournaments` | Listar torneos |
| GET | `/api/v1/tournaments/id/:id` | Ver torneo |
| GET | `/api/v1/tournaments/s/:slug` | Ver por slug |
| GET | `/api/v1/tournaments/id/:id/leaderboard` | Clasificación |
| GET | `/api/v1/tournaments/id/:id/events` | Eventos del torneo |
| GET | `/api/v1/tournaments/id/:id/sessions` | Sesiones del torneo |

### Eventos (Público)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/events/id/:id` | Ver evento |
| GET | `/api/v1/events/s/:slug` | Ver por slug |
| GET | `/api/v1/events/id/:id/selections` | Selecciones disponibles |

### Usuario (Autenticado)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/me` | Mi perfil |
| POST | `/api/v1/tournaments/:id/join` | Inscribirse a torneo |
| POST | `/api/v1/tournaments/:id/sessions/picks` | Enviar pronósticos |
| GET | `/api/v1/my-sessions/:session_id/picks` | Ver mis pronósticos |
| GET | `/api/v1/wallet/balance` | Consultar saldo |
| POST | `/api/v1/wallet/deposit` | Recargar saldo |
| GET | `/api/v1/wallet/history` | Historial de transacciones |
| GET | `/api/v1/payment-methods` | Métodos de pago |
| POST | `/api/v1/payment-methods` | Agregar método de pago |
| DELETE | `/api/v1/payment-methods/:id` | Eliminar método de pago |

### Administrador
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/admin/users` | Listar usuarios |
| POST | `/api/v1/admin/tournaments` | Crear torneo |
| POST | `/api/v1/admin/sessions` | Crear sesión |
| POST | `/api/v1/admin/events` | Crear evento |
| POST | `/api/v1/admin/events/selections` | Crear selección |
| POST | `/api/v1/admin/events/:id/settle` | Liquidar evento |

## Pruebas

Ejecutar todas las pruebas:
```bash
go test ./tests/... -v
```

Ejecutar pruebas específicas:
```bash
go test ./tests/... -v -run TestLogin
```

## Usuarios de Prueba

### Administrador
- **Email:** admin@betsystem.com
- **Password:** Admin123!

## Tecnologías Utilizadas

- **Framework:** Gin (Go)
- **ORM:** GORM
- **Base de datos:** PostgreSQL
- **Autenticación:** JWT
- **Documentación:** Swagger/OpenAPI
- **Testing:** Go testing + testify

## Contribución

1. Fork del repositorio
2. Crear rama (`git checkout -b feature/nueva-funcionalidad`)
3. Commit de cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

## Licencia

MIT License
