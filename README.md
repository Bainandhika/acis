# 💰 ACIS - Manajemen Keuangan Keluarga

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-v4-38B2AC?logo=tailwind-css)](https://tailwindcss.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15_%2F_Supabase-316192?logo=postgresql)](https://www.postgresql.org/)
[![Security](https://img.shields.io/badge/Security-OWASP_Top_10_Compliant-green)](https://owasp.org/)

**ACIS** is a secure, lightweight, and user-friendly household financial management platform. Designed for families to track cash flow, manage envelope-based budgets (virtual wallets with short IDs), collaborate on expense proposals/approvals, and record transactions seamlessly via both Web and Telegram Bot.

> 🛡️ **Security First:** Built with strict adherence to **OWASP Top 10** standards, featuring Supabase Auth (OAuth/JWKS verification), PostgreSQL **Row-Level Security (RLS)** with dual connection pools, transactional outbox notifications, rate limiting, and context-aware distributed query tracing (`X-Transaction-ID`).

---

## 🚀 Tech Stack

### Backend (`apps/backend`)
- **Language:** Golang 1.22+
- **Framework:** Gin (HTTP Web Router)
- **Database Access:** Native SQL with `sqlx` (PostgreSQL / Supabase)
- **Authentication:** Supabase Auth JWT verification via JWKS (`lestrrat-go/jwx/v2`)
- **Database Security:** PostgreSQL Row-Level Security (RLS) with Dual Connection Pool (`userDB` subject to RLS + `adminDB` for background workers)
- **Caching & Rate Limiting:** Redis + In-memory Token Bucket rate limiter
- **Async Workers:** Transactional Outbox Pattern with worker pool for Telegram low-balance alerts and notifications
- **Observability:** `slog` structured logging with end-to-end `X-Transaction-ID` trace propagation

### Telegram Bot (`apps/bot`)
- **Language:** Golang 1.22+
- **Architecture:** Standalone polling bot service communicating securely with backend internal endpoints (`X-Bot-Secret`)
- **Capabilities:** Account & family linking, real-time balance queries, and instant transaction recording

### Frontend (`apps/frontend`)
- **Framework:** Vue 3 (Composition API, `<script setup>`)
- **Build Tool:** Vite
- **State Management:** Pinia
- **Routing:** Vue Router
- **Styling:** Tailwind CSS v4 + PostCSS
- **Authentication:** `@supabase/supabase-js` (Google OAuth / Passwordless)

### Database & Migrations (`apps/migrations`)
- **Database:** PostgreSQL 15 / Supabase
- **Migrations:** Versioned SQL migrations managed via custom Go migration runner (`apps/backend/cmd/migrate`)

---

## 🏗️ System Architecture & Key Features

```
                   ┌─────────────────────────────────────────┐
                   │               User Client               │
                   │  (Vue 3 SPA / Tailwind CSS / Supabase)   │
                   └────────────────────┬────────────────────┘
                                        │ Bearer JWT (Supabase)
                                        ▼
┌─────────────────┐  Internal Secret    ┌─────────────────────────────────────────┐
│  Telegram Bot   │────────────────────▶│             Backend API (Gin)           │
│  (Polling svc)  │   (X-Bot-Secret)    │  • JWKS JWT Validation                  │
└─────────────────┘                     │  • Context & Trace ID Propagation       │
                                        │  • Rate Limiting & Security Headers     │
                                        └───────────┬─────────────────┬───────────┘
                                                    │                 │
                               (SET LOCAL ROLE)     │                 │ (Service Role)
                                    userDB          │                 │ adminDB
                                        ▼           ▼                 ▼
                                  ┌─────────────────────────────────────────┐
                                  │      PostgreSQL (Row-Level Security)    │
                                  │  • Multi-tenant Family Isolation        │
                                  │  • Wallets, Transactions, Proposals     │
                                  │  • Outbox Notification Queue            │
                                  └────────────────────┬────────────────────┘
                                                       │
                                                       ▼
                                        ┌─────────────────────────────┐
                                        │    Outbox Worker Pool       │
                                        │ (Dispatches Telegram alerts)│
                                        └─────────────────────────────┘
```

### 1. 🔐 Multi-Tenant Security & Row-Level Security (RLS)
- Every authenticated request sets PostgreSQL local session variables (`request.jwt.claims` and `SET LOCAL ROLE authenticated`) in transaction scope.
- Postgres RLS policies ensure users can strictly access only data belonging to their own family.
- Backend uses a **Dual Connection Pool**:
  - `userDB` (`DATABASE_APP_DSN`): runs under the `acis_app` / `authenticated` role where RLS is strictly enforced.
  - `adminDB` (`DATABASE_ADMIN_DSN`): service-role connection for background workers, outbox pollers, and migrations.

### 2. 👛 Envelope Budgeting & Virtual Wallets
- Create virtual wallets with friendly short identifiers (e.g. `SMTH01-1`, `SMTH01-2`) and minimum balance alert thresholds.
- Direct transactions (Income & Expense) automatically update wallet balances and trigger low-balance notifications when limits are breached.

### 3. 📝 Expense Proposals & Approvals
- Family members can submit expense proposals with target amounts and descriptions.
- Family administrators review, approve (which materializes into a real transaction), or reject proposals.

### 4. 🤖 Telegram Bot Integration & Linking
- **Account Linking:** Generate a 6-character linking code in the web interface and execute `/link <KODE>` in Telegram to connect your Telegram account.
- **Family Group Linking:** Link family invite codes to Telegram groups.
- **Bot Commands:**
  - `/start` - Display onboarding guide and command overview.
  - `/link <KODE>` or `/hubungkan <KODE>` - Link Telegram account or family chat.
  - `/saldo` or `/balance` - Check live balances and short IDs for all family wallets.
  - `/transaksi <WALLET_ID> <pemasukan/pengeluaran> <nominal> <keterangan>` - Record transactions instantly via chat.

---

## 📁 Monorepo Structure

```
acis/
├── apps/
│   ├── backend/             # Golang REST API & Background Worker
│   │   ├── cmd/
│   │   │   ├── api/         # Main server entrypoint
│   │   │   └── migrate/     # SQL migration runner
│   │   ├── config/          # YAML + Environment variable configuration
│   │   ├── domain/          # Domain business logic (auth, bot, family, transaction)
│   │   ├── infrastructure/  # DB pools, JWKS middleware, Outbox worker, Telegram client
│   │   └── shared/          # Logger (slog), Redis cache, Token bucket rate limiter
│   ├── bot/                 # Standalone Telegram Polling Bot
│   │   ├── backendclient/   # Internal HTTP client for Backend API
│   │   ├── bot/             # Message dispatcher & command handlers
│   │   └── telegram/        # Telegram Bot API client
│   ├── frontend/            # Vue 3 SPA
│   │   ├── src/
│   │   │   ├── components/  # Reusable UI components
│   │   │   ├── views/       # Dashboard, Transactions, Family, Proposals, Auth
│   │   │   ├── stores/      # Pinia stores (auth, etc.)
│   │   │   └── services/    # API and Supabase clients
│   └── migrations/          # 001 to 018+ versioned SQL migration files
├── docker-compose.yml       # Local PostgreSQL, Redis, PgAdmin, and Bot container
└── README.md
```

---

## 🛠️ Local Development Setup

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

---

### 1. Start Local Infrastructure (PostgreSQL & Redis)
```bash
git clone https://github.com/Bainandhika/acis.git
cd acis
docker-compose up -d acis_db redis pgadmin
```

---

### 2. Configure Environment Files

#### Backend (`apps/backend/.env`)
Copy `apps/backend/.env.example` to `apps/backend/.env`:
```env
DATABASE_APP_DSN=postgresql://postgres:acis_secret_password@localhost:5432/acis_db?sslmode=disable
DATABASE_ADMIN_DSN=postgresql://postgres:acis_secret_password@localhost:5432/acis_db?sslmode=disable
SUPABASE_JWKS_URL=https://<your_project_id>.supabase.co/auth/v1/.well-known/jwks.json

REDIS_URL=redis://localhost:6379/0
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_from_botfather
TELEGRAM_BOT_USERNAME=your_acis_bot_username
BOT_INTERNAL_SECRET=changeme_internal_bot_secret_key
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

#### Bot Service (`apps/bot/.env`)
Copy `apps/bot/.env.example` to `apps/bot/.env`:
```env
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_from_botfather
BACKEND_BASE_URL=http://localhost:8080
BOT_INTERNAL_SECRET=changeme_internal_bot_secret_key
```

#### Frontend (`apps/frontend/.env`)
Copy `apps/frontend/.env.example` to `apps/frontend/.env`:
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_SUPABASE_URL=https://<your_project_id>.supabase.co
VITE_SUPABASE_ANON_KEY=your_supabase_anon_key
```

---

### 3. Run Database Migrations
```bash
cd apps/backend
go run ./cmd/migrate
```

---

### 4. Run Services

#### Start Backend API
```bash
cd apps/backend
go run ./cmd/api/main.go
# API running on http://localhost:8080
```

#### Start Telegram Bot Service
```bash
cd apps/bot
go run ./cmd/main.go
```

#### Start Frontend Web App
```bash
cd apps/frontend
npm install
npm run dev
# Frontend running on http://localhost:5173
```

---

## 📡 API Overview (`/api/v1`)

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | Service health, DB & Redis latency | Public |
| `POST` | `/auth/provision` | Provision user profile upon first Supabase login | Supabase JWT |
| `GET` | `/auth/me` | Retrieve authenticated user profile | Supabase JWT |
| `POST` | `/telegram/link-code` | Generate 6-char link code for Telegram bot | Supabase JWT |
| `POST` | `/family` | Create new family group | Supabase JWT |
| `POST` | `/family/join` | Join family via 6-char invite code | Supabase JWT |
| `GET` | `/family/me` | Get current family details and members | Supabase JWT |
| `PATCH` | `/family` | Update family profile (name) | Admin Role |
| `GET` | `/family/wallets` | List family wallets and balances | Family Member |
| `POST` | `/family/wallets` | Create new virtual wallet | Admin Role |
| `PATCH` | `/family/wallets/:id` | Update wallet details | Admin Role |
| `DELETE` | `/family/wallets/:id` | Delete virtual wallet | Admin Role |
| `DELETE` | `/family/members/:id` | Remove member from family | Admin Role |
| `GET` | `/transaction` | Get family transactions by month/year | Family Member |
| `POST` | `/transaction` | Create direct income/expense transaction | Admin Role |
| `PATCH` | `/transaction/:id` | Update existing transaction | Admin Role |
| `DELETE` | `/transaction/:id` | Delete transaction | Admin Role |
| `GET` | `/transaction/proposals` | List pending/reviewed expense proposals | Family Member |
| `POST` | `/transaction/proposals` | Submit new expense proposal | Family Member |
| `POST` | `/transaction/proposals/:id/approve` | Approve proposal (creates transaction) | Admin Role |
| `POST` | `/transaction/proposals/:id/reject` | Reject proposal | Admin Role |
| `POST` | `/bot/link` | Internal Telegram link endpoint | `X-Bot-Secret` |
| `GET` | `/bot/balance` | Internal wallet balances query | `X-Bot-Secret` |
| `POST` | `/bot/transaction` | Internal bot direct transaction entry | `X-Bot-Secret` |

---

## 📄 License
This project is private and maintained for family financial tracking.

