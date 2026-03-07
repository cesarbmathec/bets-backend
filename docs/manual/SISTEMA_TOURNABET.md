# Documentación Técnica del Sistema TournaBet

## Tabla de Contenidos

1. [Introducción y Visión General](#1-introducción-y-visión-general)
2. [Arquitectura del Sistema](#2-arquitectura-del-sistema)
3. [Modelos de Datos](#3-modelos-de-datos)
4. [Endpoints de la API](#4-endpoints-de-la-api)
5. [Frontend - Estructura y Componentes](#5-frontend---estructura-y-componentes)
6. [Flujo de Trabajo y Casos de Uso](#6-flujo-de-trabajo-y-casos-de-uso)
7. [Configuración y Ejecución](#7-configuración-y-ejecución)
8. [Características Implementadas](#8-características-implementadas)
9. [Estado Actual del Proyecto](#9-estado-actual-del-proyecto)

---

## 1. Introducción y Visión General

### 1.1 ¿Qué es TournaBet?

**TournaBet** es una plataforma completa para gestionar torneos de pronósticos deportivos (quinielas), donde los participantes realizan selecciones (pronósticos) y acumulan puntos según los resultados de los eventos deportivos. El sistema soporta múltiples tipos de deportes y modalidades de apuesta.

### 1.2 Tipos de Torneos Soportados

El sistema actualmente soporta tres tipos principales de torneos:

- **Fútbol**: Apuestas tradicionales (ganador, línea de dinero, over/under)
- **Béisbol**: Runline, over/under
- **Caballos (Polla Hípica)**: Macho/Hembra, Alta/Baja

### 1.3 Tipos de Selección Disponibles

| Tipo | Código | Descripción |
|------|--------|-------------|
| Macho | `macho` | Apuesta al caballo/equipo favorito |
| Hembra | `hembra` | Apuesta al caballo/equipo no favorito |
| Alta | `alta` | Apuesta a que el score total será mayor a la línea |
| Baja | `baja` | Apuesta a que el score total será menor a la línea |
| Runline | `runline` | Línea de carrera (diferencial de puntos) |
| Super Alta | `sralta` | Sobre alta con línea especial (SuperLine) |
| Super Baja | `srbaja` | Sobre baja con línea especial (SuperLine) |

### 1.4 Roles de Usuario

- **Admin**: Gestiona torneos, categorías, sesiones, eventos globales y liquidaciones
- **Usuario**: Se registra, participa en torneos, hace pronósticos y consulta el leaderboard

### 1.5 Usuario Administrador Inicial

Al ejecutar la aplicación por primera vez, se crea automáticamente:
- **Email**: admin@betsystem.com
- **Password**: Admin123!

---

## 2. Arquitectura del Sistema

### 2.1 Stack Tecnológico

#### Backend (Go)
- **Lenguaje**: Go 1.21+
- **Framework Web**: Gin Gonic
- **ORM**: GORM (MySQL/PostgreSQL)
- **Autenticación**: JWT (JSON Web Tokens)
- **Documentación**: Swagger/OpenAPI

#### Frontend (React)
- **Framework**: React 18+ con Vite
- **Lenguaje**: TypeScript
- **Estado Global**: Redux Toolkit
- **HTTP Client**: Axios
- **Estilos**: Tailwind CSS v4
- **Enrutamiento**: React Router DOM v6

### 2.2 Estructura de Directorios

```
tournabet_system/
├── bets-backend/                 # Backend en Go
│   ├── config/                   # Configuración de base de datos
│   ├── controllers/              # Controladores de la API
│   ├── docs/                     # Documentación Swagger
│   ├── docs/manual/              # Manual LaTeX
│   ├── dtos/                     # Data Transfer Objects
│   ├── middlewares/              # Middlewares (auth, security)
│   ├── migrations/               # Migraciones de base de datos
│   ├── models/                   # Modelos de GORM
│   ├── routes/                   # Definición de rutas
│   ├── tests/                    # Pruebas unitarias
│   ├── utils/                    # Utilidades
│   └── main.go                   # Punto de entrada
│
└── bets-frontend/                # Frontend en React
    ├── src/
    │   ├── api/                  # Clientes API (Axios)
    │   ├── components/          # Componentes reutilizables
    │   │   └── layout/          # Layouts (Admin, User)
    │   ├── context/             # Contextos de React
    │   ├── docs/                # Documentación frontend
    │   ├── hooks/               # Hooks personalizados
    │   ├── pages/               # Páginas
    │   │   ├── admin/          # Dashboard Admin
    │   │   └── user/           # Dashboard Usuario
    │   ├── store/              # Redux store y slices
    │   ├── types/              # Tipos TypeScript
    │   └── main.tsx            # Punto de entrada
    └── package.json
```

---

## 3. Modelos de Datos

### 3.1 Modelo Relacional (Descripción Conceptual)

```
┌─────────────────┐       ┌─────────────────────┐       ┌─────────────────┐
│   Tournament    │       │     Session         │       │     Category    │
│   (Torneo)      │──────<│  (Jornada/Sesión)   │       │   (Categoría)   │
└────────┬────────┘       └─────────────────────┘       └─────────────────┘
         │                                                    ▲
         │ 1:N                                               │
         ▼                                                    │
┌─────────────────────┐       ┌─────────────────────┐         │
│  TournamentEvent   │       │    PickableSelection│         │
│  (Tabla Intermedia)│──────<│  (Opciones de Apuesta)│        │
└────────┬────────────┘       └─────────────────────┘         │
         │                                                    │
         │ N:1                                                │
         ▼                                                    │
┌─────────────────────┐       ┌─────────────────────┐         │
│      Event         │       │   EventCompetitor   │         │
│   (Evento Global)  │───N:N─<│  (Competidor por    │         │
└─────────────────────┘       │     Evento)         │─────────┘
         │                     └─────────────────────┘
         │                            │
         │                            ▼
         │                     ┌─────────────────┐
         │                     │   Competitor    │
         │                     │ (Catálogo)     │
         └────────────────────>└─────────────────┘
                                    
┌─────────────────────┐       ┌─────────────────────┐
│ TournamentParticipant│      │     UserPick       │
│ (Participante)      │──1:N─<│  (Selección Usuario)│
└─────────────────────┘       └─────────────────────┘

┌─────────────────────┐       ┌─────────────────────┐
│       User          │──1:1─<│      Wallet         │
│      (Usuario)      │       │    (Billetera)     │
└─────────────────────┘       └─────────────────────┘
```

### 3.2 Modelos Principales (Backend Go)

#### Event (Evento Global)
Ubicación: `bets-backend/models/event.go`

```go
// Event representa un evento/deporte/partido global que puede ser asignado a diferentes torneos.
// Los eventos son independientes de los torneos y se relacionan a través de TournamentEvent.
type Event struct {
    BaseModel
    Name      string    // Nombre del evento (ej: "España vs Francia")
    Slug      string    // URL amigable
    Order     int       // Orden (ej: Carrera #1, Partido #1)
    StartTime time.Time // Hora de inicio
    Venue     string    // Lugar (estadio, hipódromo)
    Line      float64   // Línea Over/Under (ej: 2.5 goles)
    Status    string    // scheduled, live, completed, cancelled
    ResultNote string   // Nota de resultado
    TotalScore float64  // Score total actual
    
    // Relaciones
    Competitors        []EventCompetitor
    PickableSelections []PickableSelection
}
```

#### TournamentEvent (Tabla Intermedia)
```go
// TournamentEvent define la relación muchos a muchos entre Tournament y Event.
// Permite que un evento sea usado en múltiples torneos.
type TournamentEvent struct {
    BaseModel
    TournamentID uint       // ID del torneo
    EventID      uint       // ID del evento global
    SessionID    *uint      // Sesión específica en este torneo
    Order        int        // Orden del evento en la sesión
    
    // Relaciones
    Tournament Tournament
    Event      Event
    Session    Session
}
```

#### Tournament (Torneo)
Ubicación: `bets-backend/models/tournament.go`

```go
type Tournament struct {
    BaseModel
    Name             string            // Nombre del torneo
    Slug             string            // URL amigable
    Description      string            // Descripción
    Category         string            // Categoría (Fútbol, Caballos, etc.)
    Status           string            // open, closed, finished
    StartDate        time.Time         // Fecha de inicio
    EndDate          time.Time         // Fecha de fin
    MaxParticipants  int               // Límite (0 = ilimitado)
    EntryFee         float64           // Costo de inscripción
    PrizePool        float64           // Premio acumulado
    AdminFeePercent  float64           // % comisión casa
    
    Settings         TournamentSettings // Configuración JSON
    
    // Relaciones
    Events       []TournamentEvent
    Participants []TournamentParticipant
}
```

#### TournamentSettings (Configuración del Torneo)
```go
type TournamentSettings struct {
    PrizeDistribution        []float64           // [0.7, 0.2, 0.1] para 1ro, 2do, 3ro
    SelectionsPerSession    int                 // Selecciones por sesión
    HorseRacingPoints       []int               // [10, 5, 3] para 1ro, 2do, 3ro en caballos
    RequiredSelectionTypes  []string            // ["macho", "hembra", "alta", "baja"]
    PointsBySelectionType   map[string]int      // {"macho": 3, "hembra": 5, ...}
    ExtraPointsForPerfectSession int            // Puntos por sesión perfecta
    TotalSessions           int                 // Total de sesiones
    FreeSelection           bool                //true = elección libre
    SportCategory           string              // "futbol", "beisbol", "caballos"
}
```

#### Session (Sesión/Jornada)
Ubicación: `bets-backend/models/session.go`

```go
type Session struct {
    BaseModel
    TournamentID  uint       // ID del torneo
    SessionNumber int        // Número de sesión (1, 2, 3...)
    StartTime    time.Time  // Inicio del período de picks
    EndTime      time.Time  // Fin del período de picks
    SuperLine    float64    // Línea especial para playoffs
    Description  string     // Descripción
    Status       string     // open, closed, settled
}
```

#### EventCompetitor (Competidor por Evento)
```go
type EventCompetitor struct {
    BaseModel
    EventID uint
    CompetitorID *uint     // Referencia al catálogo global
    
    Name           string  // Nombre del competidor
    AssignedNumber int     // Número asignado
    Odds           int     // Cuota (ej: -300, +400)
    Runline        float64 // Runline (ej: -1.5)
    SuperRunline   float64 // Super Runline
    IsFavorite     bool    // true = Macho (favorito)
    FinalScore     int     // Score final (deportes de equipo)
    Position       int     // Posición (carreras: 1=ganador)
    IsScratched    bool    // Si el caballo se retiró
    ScoredFirst    bool    // Marcó primero
    ScoredFirstHalf bool   // Marcó en primer tiempo
    ScoredSecondHalf bool  // Marcó en segundo tiempo
    ScoredFirstInning bool // Marcó en primer inning
}
```

#### PickableSelection (Opción de Apuesta)
```go
type PickableSelection struct {
    BaseModel
    EventID          uint
    Description      string  // "España gana", "Más de 2.5"
    SelectionType    string  // macho, hembra, alta, baja, runline, sralta, srbaja
    Line             float64 // Línea para alta/baja
    Odds             float64
    PointsForWin     int     // Puntos por acierto
    PointsForPush    int     // Puntos por empate
    CompetitorID     *uint   // Competidor asociado
    Status           string  // pending, won, lost
    
    Competitor       Competitor
}
```

#### UserPick (Selección del Usuario)
```go
type UserPick struct {
    BaseModel
    ParticipantID   uint    // ID del participante
    SelectionID     uint    // ID de la selección
    SessionID       uint    // ID de la sesión
    Status          string  // pending, won, lost, push
    AwardedPoints   int     // Puntos otorgados
    
    Participant     TournamentParticipant
    Selection       PickableSelection
    Session         Session
}
```

#### TournamentParticipant (Participante)
```go
type TournamentParticipant struct {
    BaseModel
    UserID       uint
    TournamentID uint
    TotalPoints  int
    
    User         User
    Tournament   Tournament
}
```

#### User (Usuario)
```go
type User struct {
    BaseModel
    Username     string
    Email        string
    Password     string (hash)
    Role         string  // admin, user
    Status       string  // active, inactive
    Wallet       Wallet
}
```

#### Wallet (Billetera)
```go
type Wallet struct {
    BaseModel
    UserID          uint
    Balance         float64  // Saldo disponible
    FrozenBalance   float64  // Saldo congelado
    BonusBalance    float64  // Bono
    TokenBalance    int     // Tokens
    Currency        string   // USD, etc.
}
```

---

## 4. Endpoints de la API

### 4.1 Estructura de URLs

```
/api/v1/
├── /auth/                    # Autenticación (público)
│   ├── POST /login           # Iniciar sesión
│   └── POST /register        # Registrarse
│
├── /categories              # Categorías (público)
│   ├── GET /                 # Listar categorías
│   └── GET /:id              # Obtener categoría por ID
│
├── /tournaments             # Torneos (público)
│   ├── GET /                 # Listar torneos
│   ├── GET /id/:id           # Obtener torneo por ID
│   ├── GET /s/:slug          # Obtener torneo por slug
│   ├── GET /id/:id/events    # Obtener eventos del torneo
│   ├── GET /id/:id/sessions  # Obtener sesiones del torneo
│   ├── GET /id/:id/leaderboard # Obtener ranking
│   └── GET /my-tournaments   # Mis torneos (auth)
│
├── /events                   # Eventos globales (público)
│   ├── GET /                 # Listar eventos globales
│   ├── GET /:id              # Obtener evento por ID
│   └── GET /:id/selections  # Obtener selecciones del evento
│
├── /session-events           # Eventos de sesión
│   ├── GET /:id              # Obtener sesión por ID
│   └── GET /:id/events      # Obtener eventos de una sesión
│
├── /competitors              # Competidores (público)
│   ├── GET /                 # Listar competidores
│   └── GET /categories       # Obtener categorías de competidores
│
├── /me                       # Perfil usuario (auth)
│   └── GET /                 # Obtener perfil
│
├── /wallet                   # Billetera (auth)
│   ├── GET /balance          # Consultar saldo
│   ├── POST /deposit         # Depositar
│   └── GET /history          # Historial de transacciones
│
├── /tournaments/:id/         # Torneos usuario (auth)
│   ├── POST /join            # Inscribirse al torneo
│   └── POST /sessions/picks  # Enviar pronósticos
│
├── /my-sessions/:session_id/ # Picks usuario (auth)
│   └── GET /picks            # Ver mis picks de sesión
│
└── /admin/                   # Rutas de ADMIN
    ├── /users/
    │   ├── GET /             # Listar usuarios
    │   ├── GET /:id         # Obtener usuario
    │   ├── PATCH /:id/role   # Cambiar rol
    │   └── PATCH /:id/status # Cambiar estado
    │
    ├── /tournaments/
    │   ├── POST /            # Crear torneo
    │   └── PATCH /:id/status # Actualizar estado
    │
    ├── /sessions/
    │   ├── POST /            # Crear sesión
    │   └── PATCH /:id/status # Actualizar estado
    │
    ├── /events/
    │   ├── POST /            # Crear evento global
    │   ├── PUT /:id          # Actualizar evento
    │   ├── DELETE /:id      # Eliminar evento
    │   ├── POST /:event_id/selections     # Crear selección
    │   ├── POST /:event_id/competitors    # Asignar competidores
    │   ├── POST /:event_id/settle        # Liquidar evento
    │   └── GET /available                # Eventos disponibles
    │
    ├── /tournament-events/
    │   ├── POST /            # Asignar evento a torneo
    │   └── DELETE /:id      # Remover evento
    │
    ├── /categories/
    │   ├── POST /            # Crear categoría
    │   ├── PUT /:id         # Actualizar categoría
    │   ├── DELETE /:id      # Eliminar categoría
    │   └── PATCH /:id/status # Toggle estado
    │
    └── /competitors/
        ├── POST /            # Crear competidor
        ├── PUT /:id          # Actualizar competidor
        └── DELETE /:id      # Eliminar competidor
```

### 4.2 Autenticación

Todos los endpoints protegidos requieren el header:
```
Authorization: Bearer <token_jwt>
```

### 4.3 Ejemplos de Solicitudes

#### Registro de Usuario
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "usuario@email.com",
  "password": "Password123!",
  "name": "Juan Pérez",
  "username": "juanperez"
}
```

#### Inicio de Sesión
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@betsystem.com",
  "password": "Admin123!"
}
```

#### Crear Torneo (Admin)
```bash
POST /api/v1/admin/tournaments
Authorization: Bearer <token_admin>
Content-Type: application/json

{
  "name": "Copa Mundial 2026",
  "category": "Fútbol",
  "description": "Pronósticos del mundial",
  "start_date": "2026-06-01T00:00:00Z",
  "end_date": "2026-07-15T23:59:59Z",
  "entry_fee": 500,
  "max_participants": 100,
  "settings": {
    "selections_per_session": 4,
    "total_sessions": 8,
    "required_selection_types": ["macho", "hembra", "alta", "baja"],
    "points_by_selection_type": {
      "macho": 3,
      "hembra": 5,
      "alta": 3,
      "baja": 3,
      "runline": 4,
      "sralta": 5,
      "srbaja": 5
    },
    "prize_distribution": [0.7, 0.2, 0.1],
    "extra_points_for_perfect_session": 2,
    "sport_category": "futbol"
  }
}
```

#### Crear Evento Global (Admin)
```bash
POST /api/v1/admin/events
Authorization: Bearer <token_admin>
Content-Type: application/json

{
  "name": "España vs Francia",
  "venue": "Santiago Bernabéu",
  "line": 2.5,
  "start_time": "2026-06-15T20:00:00Z",
  "status": "scheduled"
}
```

#### Asignar Evento a Torneo (Admin)
```bash
POST /api/v1/admin/tournament-events
Authorization: Bearer <token_admin>
Content-Type: application/json

{
  "event_id": 1,
  "tournament_id": 1,
  "session_id": 1,
  "order": 1
}
```

#### Establecer Competidores de un Evento (Admin)
```bash
POST /api/v1/admin/events/1/competitors
Authorization: Bearer <token_admin>
Content-Type: application/json

{
  "competitors": [
    {
      "name": "Selección Española",
      "assigned_number": 1,
      "odds": -300,
      "rl": -1.5,
      "srl": -2.5,
      "is_favorite": true
    },
    {
      "name": "Selección Francesa",
      "assigned_number": 2,
      "odds": 250,
      "rl": 1.5,
      "srl": 2.5,
      "is_favorite": false
    }
  ]
}
```

#### Crear Selecciones (Admin)
```bash
POST /api/v1/admin/events/1/selections
Authorization: Bearer <token_admin>
Content-Type: application/json

{
  "event_id": 1,
  "description": "España gana",
  "selection_type": "macho",
  "competitor_id": 1,
  "points_for_win": 3
}
```

#### Inscribirse al Torneo (Usuario)
```bash
POST /api/v1/tournaments/1/join
Authorization: Bearer <token_usuario>
```

#### Enviar Pronósticos (Usuario)
```bash
POST /api/v1/tournaments/1/sessions/picks
Authorization: Bearer <token_usuario>
Content-Type: application/json

{
  "session_id": 1,
  "selection_ids": [1, 2, 3, 4]
}
```

#### Liquidar Evento (Admin)
```bash
POST /api/v1/admin/events/1/settle
Authorization: Bearer <token_admin>
Content-Type: application/json

{
  "results": [
    {"competitor_id": 1, "final_score": 2},
    {"competitor_id": 2, "final_score": 1}
  ]
}
```

---

## 5. Frontend - Estructura y Componentes

### 5.1 Tecnologías del Frontend

- **React 18+** con Vite
- **TypeScript** para tipado estático
- **Redux Toolkit** para gestión de estado
- **Axios** para HTTP requests
- **React Router DOM v6** para enrutamiento
- **Tailwind CSS v4** para estilos
- **React Hook Form** para formularios
- **date-fns** para fechas

### 5.2 Estructura de Archivos

```
bets-frontend/src/
├── api/
│   ├── client.ts           # Configuración de Axios
│   ├── categories.ts       # Endpoints de categorías
│   ├── competitors.ts      # Endpoints de competidores
│   ├── picks.ts            # Endpoints de picks
│   └── tournaments.ts     # Endpoints de torneos
│
├── components/
│   ├── ConfirmModal.tsx    # Modal de confirmación
│   └── layout/
│       ├── AdminLayout.tsx # Layout del dashboard admin
│       ├── UserLayout.tsx  # Layout del dashboard usuario
│       └── ProtectedRoute.tsx # Ruta protegida
│
├── hooks/
│   └── useTheme.ts         # Hook para modo oscuro/claro
│
├── pages/
│   ├── LoginPage.tsx       # Página de login
│   ├── RegisterPage.tsx    # Página de registro
│   ├── DashboardPage.tsx   # Redirect según rol
│   ├── admin/
│   │   ├── AdminDashboardHome.tsx  # Home admin
│   │   ├── TournamentsPage.tsx      # Gestión de torneos
│   │   ├── TournamentDetailPage.tsx # Detalle de torneo
│   │   ├── CategoriesPage.tsx       # Gestión de categorías
│   │   └── CompetitorsPage.tsx      # Gestión de competidores
│   └── user/
│       ├── UserDashboardHome.tsx     # Home usuario
│       └── UserPicksPage.tsx        # Picks del usuario
│
├── store/
│   ├── index.ts            # Configuración del store
│   ├── hooks.ts           # Hooks tipados
│   └── slices/
│       └── authSlice.ts   # Estado de autenticación
│
├── types/
│   ├── api.ts             # Tipos de la API
│   ├── auth.ts            # Tipos de autenticación
│   └── index.ts           # Exportaciones
│
├── App.tsx                # Componente principal
├── main.tsx               # Punto de entrada
└── index.css              # Estilos globales
```

### 5.3 Autenticación (Frontend)

El frontend usa Redux Toolkit para manejar el estado de autenticación:

```typescript
// store/slices/authSlice.ts
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}
```

El token JWT se almacena en localStorage y se incluye en todas las peticiones Axios automáticamente.

### 5.4 Páginas Principales

#### Admin Dashboard
- **AdminDashboardHome**: Panel principal del admin con estadísticas
- **TournamentsPage**: Crear/editar/eliminar tournaments
- **TournamentDetailPage**: Detalles de un tournament con sesiones y eventos
- **CategoriesPage**: Gestionar categorías (Fútbol, Caballos, etc.)
- **CompetitorsPage**: Gestionar catálogo de competidores

#### User Dashboard
- **UserDashboardHome**: Torneos disponibles e inscritos
- **UserPicksPage**: Hacer pronósticos y ver historial

### 5.5 Tema Oscuro/Claro

El sistema implementa un theme switcher que permite cambiar entre modo oscuro y claro:

- **Hook personalizado**: `useTheme.ts`
- **Almacenamiento**: localStorage
- **Clases CSS**: Tailwind con `dark:` prefix

---

## 6. Flujo de Trabajo y Casos de Uso

### 6.1 Flujo Completo de un Torneo

```
1. ADMIN: Crear Categoría
   └─> Ej: "Fútbol", "Caballos", "Beisbol"

2. ADMIN: Crear Tournament
   ├─> Configurar sesiones (8 jornadas)
   ├─> Configurar selecciones por sesión (4)
   ├─> Configurar tipos requeridos (macho, hembra, alta, baja)
   ├─> Configurar puntos por tipo
   └─> Configurar distribución de premios (70%, 20%, 10%)

3. ADMIN: Crear Sesiones
   └─> Cada sesión tiene: start_time, end_time, super_line

4. ADMIN: Crear Eventos Globales
   └─> Ej: "España vs Francia", "Carrera 5"

5. ADMIN: Asignar Eventos a Tournament
   └─> Vincular evento global a una sesión específica

6. ADMIN: Establecer Competidores
   └─> Por cada evento, definir equipos/caballos con odds

7. ADMIN: Crear Selecciones
   ├─> Macho (apuesta al favorito)
   ├─> Hembra (apuesta al no favorito)
   ├─> Alta (over)
   ├─> Baja (under)
   └─> Runline, SuperLine (para playoffs)

8. USUARIO: Registrarse e Iniciar sesión

9. USUARIO: Recargar Billetera
   └─> Realizar depósito

10. USUARIO: Inscribirse al Tournament
    └─> Pagar fee de inscripción

11. USUARIO: Hacer Pronósticos
    └─> Seleccionar opciones para cada sesión
    └─> Sistema valida: cantidad, tipos requeridos, timing

12. ADMIN: Liquidar Eventos
    └─> Ingresar resultados (score, posiciones)
    └─> Sistema calcula: picks winners, puntos, leaderboard

13. USUARIO/ADMIN: Consultar Leaderboard
    └─> Ver ranking en tiempo real

14. ADMIN: Finalizar Tournament
    └─> Distribuir premios según posición
```

### 6.2 Ejemplo: Tournament de Fútbol

**Configuración del Tournament:**
- Nombre: "La Liga 2026"
- Sesiones: 2
- Selecciones por sesión: 4
- Tipos requeridos: ["macho", "hembra", "alta", "baja"]
- Puntos: macho=3, hembra=5, alta=3, baja=3

**Sesión 1:**
- Evento 1: Barcelona vs Levante
  - Línea: 3 goles
  - Macho: Barcelona (-300)
  - Hembra: Levante (+400)
- Evento 2: Atlético vs Real Madrid
  - Línea: 3 goles

**Picks del usuario:**
- Macho: Barcelona
- Hembra: Real Madrid
- Alta: Barcelona vs Levante
- Baja: Atlético vs Real Madrid

**Resultado:**
- Barcelona 3 - Levante 1 → Gana Macho, Alta
- Atlético 1 - Real Madrid 0 → Gana Macho, Baja

**Puntos:** 3 + 0 + 3 + 3 = 9 puntos

### 6.3 Ejemplo: Tournament de Caballos (Polla Hípica)

**Configuración:**
- Sesiones: 1
- Selecciones por sesión: 6 (6 carreras)
- Tipos: selección de caballo por carrera
- Puntos: 1ro=5, 2do=3, 3ro=1

**Sesión 1 - 6 carreras:**
- Evento 1: Carrera 1 (8 caballos)
- Evento 2: Carrera 2
- ...

**Picks del usuario:**
- Carrera 1: Caballo #3
- Carrera 2: Caballo #5
- ...

**Resultado:**
- Carrera 1: #3 (1ro), #7 (2do), #1 (3ro)
- Puntos: 5 puntos

---

## 7. Configuración y Ejecución

### 7.1 Requisitos Previos

- **Backend**: Go 1.21+, MySQL o PostgreSQL
- **Frontend**: Node.js 18+, npm o yarn

### 7.2 Configuración del Backend

El backend usa variables de entorno (no incluidas en el repositorio por seguridad):

```bash
# Variables de entorno requeridas (crear archivo .env)
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=tournabet

JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

PORT=8080
GIN_MODE=debug  # o release
```

### 7.3 Ejecución del Backend

```bash
cd bets-backend
go run main.go
```

El servidor escuchará en `http://localhost:8080`

### 7.4 Ejecución del Frontend

```bash
cd bets-frontend
npm install
npm run dev
```

El frontend escuchará en `http://localhost:5173` (por defecto)

### 7.5 Documentación API

Swagger está disponible en: `http://localhost:8080/swagger/index.html`

---

## 8. Características Implementadas

### 8.1 Gestión de Torneos
- [x] Crear/editar/eliminar tournaments
- [x] Configurar sesiones y selecciones
- [x] Tipos de selecciones requeridos
- [x] Sistema de puntos por tipo
- [x] Límite de participantes
- [x] Estados: open, closed, finished

### 8.2 Eventos Globales
- [x] Crear eventos independientes de tournaments
- [x] Reutilizar eventos en múltiples tournaments
- [x] Tabla intermedia TournamentEvent
- [x] Asignar eventos a sesiones específicas

### 8.3 Competidores
- [x] Catálogo global de competidores
- [x] Odds, RL, SRL por evento
- [x] Determinación automática Macho/Hembra

### 8.4 Selecciones
- [x] Crear selecciones por evento
- [x] Tipos: macho, hembra, alta, baja, runline, sralta, srbaja
- [x] Puntos configurables por selección

### 8.5 Sistema de Puntuación
- [x] Puntos por tipo de selección
- [x] Puntos extra por sesión perfecta
- [x] Puntos para caballos por posición
- [x] Leaderboard en tiempo real

### 8.6 Liquidación
- [x] Ingresar resultados de eventos
- [x] Calcular automáticamente picks winners
- [x] Actualizar puntos de usuarios
- [x] Soporte para carreras de caballos (posiciones)

### 8.7 Frontend
- [x] Dashboard separado para Admin y Usuario
- [x] Autenticación JWT
- [x] Gestión de tournaments (admin)
- [x] Gestión de categorías (admin)
- [x] Gestión de competidores (admin)
- [x] Vista de detalles de tournament
- [x] Hacer picks (usuario)
- [x] Ver leaderboard
- [x] Modo oscuro/claro

---

## 9. Estado Actual del Proyecto

### 9.1 En Desarrollo

El proyecto está en modo de desarrollo activo. La base de datos puede ser eliminada y recreada según sea necesario para modificar los modelos.

### 9.2 Consideraciones Técnicas

1. **Eventos Globales**: El modelo actual de Event no tiene TournamentID directo. La relación es a través de TournamentEvent (tabla intermedia).

2. **Sesiones**: Cada sesión pertenece a un Tournament específico, pero los eventos pueden compartirse entre tournaments.

3. **Puntos**: El sistema permite configurar puntos por tipo de selección en el nivel del Tournament, pero también puede sobreescribirse en cada Selection.

4. **Liquidación**: El sistema calcula automáticamente los winners basado en:
   - Resultado (score/posición)
   - Tipo de selección
   - Línea del evento
   - SuperLine de la sesión

### 9.3 Próximos Pasos Sugeridos

1. **Mejoras en Frontend**:
   - Mejora de UI/UX
   - Gráficos de leaderboard
   - Notificaciones en tiempo real
   - PWA support

2. **Mejoras en Backend**:
   - WebSockets para updates en vivo
   - Cacheo de respuestas
   - Tests unitarios más completos
   - API de pagos

3. **Características**:
   - Torneos privados con invitación
   - Torneos multi-deporte
   - Modo práctica (sin dinero real)
   - Historial de desempeño de usuarios

---

## 10. Glosario de Términos

| Término | Descripción |
|---------|-------------|
| **Tournament** | Competición completa con múltiples sesiones |
| **Session** | Jornada/sesión individual dentro de un tournament |
| **Event** | Evento deportivo individual (partido, carrera) |
| **Selection** | Opción de apuesta (macho, alta, etc.) |
| **Pick** | Pronóstico hecho por un usuario |
| **Leaderboard** | Tabla de posiciones/ranking |
| **Settle** | Liquidar un evento con resultados |
| **Line** | Línea Over/Under para apuestas alta/baja |
| **SuperLine** | Línea especial para playoffs |
| **Odds** | Cuota decimal o americana |
| **Runline** | Línea de carrera (beisbol) |
| **Macho** | Favorito en una apuesta |
| **Hembra** | No favorito en una apuesta |
| **Alta** | Apuesta a que el total supera la línea |
| **Baja** | Apuesta a que el total está bajo la línea |

---

*Documentación generada automáticamente para TournaBet*
*Versión: 1.0*
*Última actualización: 2026*
