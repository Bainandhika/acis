# 💰 ACIS (Aplikasi Catatan Keuangan Istri/Suami)

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-316192?logo=postgresql)](https://www.postgresql.org/)
[![OWASP](https://img.shields.io/badge/Security-OWASP_Top_10_Compliant-green)](https://owasp.org/)

**ACIS** is a secure, lightweight, and user-friendly household financial management web application. Designed for families to track cash flow, manage envelope-based budgets (virtual wallets), and handle expense approvals seamlessly via Web and Telegram Bot.

> 🛡️ **Security First:** Built with strict adherence to **OWASP Top 10** standards, featuring passwordless authentication, context-aware distributed tracing, and secure session management.

## 🚀 Tech Stack

### Backend (Monorepo: `apps/backend`)
- **Language:** Golang 1.22+
- **Framework:** Gin (HTTP Router)
- **Database:** Native SQL via `sqlx` (PostgreSQL)
- **Architecture:** Clean Architecture (Handler -> Service -> Repository) with Manual Dependency Injection
- **Logging:** `zerolog` + `lumberjack` with custom `sqlx` wrapper for **Context-Aware Query Tracing** (Trace ID propagation from HTTP to DB layer)

### Frontend (Monorepo: `apps/frontend`)
- **Framework:** Vue 3 (Composition API, `<script setup>`)
- **Build Tool:** Vite
- **State Management:** Pinia
- **Styling:** Tailwind CSS + DaisyUI (Mobile-first responsive)

### Infrastructure & DevOps
- **Database:** PostgreSQL (Dockerized for local dev, Supabase/Neon for prod)
- **Migration:** Custom Go CLI tool for versioned SQL migrations
- **Deployment:** Vercel (Frontend), Render/Railway (Backend)

## ️ Architecture & Key Features
This project uses a **Modular Monorepo** architecture. 

### 🔍 Context-Aware Logging (Portfolio Highlight)
Unlike standard ORMs, ACIS implements a **custom `sqlx` wrapper** (`internal/database/db.go`). This wrapper intercepts all database queries and automatically extracts the `X-Transaction-ID` from the Go `context.Context`. 
- **Result:** Every HTTP request gets a unique Trace ID. When that request hits the database, the SQL query, arguments, and execution time are logged with the *exact same Trace ID*. 
- **Benefit:** Makes debugging and auditing in production extremely easy (Distributed Tracing without heavy tools like Jaeger).

##  Security & OWASP Compliance
- **A01 Broken Access Control:** Strict RBAC middleware (Admin vs Member).
- **A03 Injection:** 100% Parameterized queries via `sqlx` (`$1`, `$2`). No raw string concatenation.
- **A07 Auth Failures:** JWT stored in `HttpOnly`, `Secure`, `SameSite=Strict` cookies. Short-lived tokens with refresh mechanism.
- **A09 Security Logging:** Structured JSON logging with daily rotation. Sensitive data (passwords, OTPs) is never logged.

## 🛠️ Local Development Setup (Windows)

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

### 1. Clone & Start Database
```bash
git clone https://github.com/Bainandhika/acis.git
cd acis
docker-compose up -d
```

### 2. Run Database Migrations
```bash
git clone https://github.com/Bainandhika/acis.git
cd acis
docker-compose up -d
```

### 3. Start Backend API
```bash
# Copy .env.example to .env and configure
go run cmd/api/main.go
# API runs on http://localhost:8080
```

### 3. Start Frontend
```bash
cd apps/frontend
npm install
npm run dev
# Frontend runs on http://localhost:5173
```

