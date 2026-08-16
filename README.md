# 💰 ACIS - Manajemen Keuangan Keluarga

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-316192?logo=postgresql)](https://www.postgresql.org/)
[![OWASP](https://img.shields.io/badge/Security-OWASP_Top_10_Compliant-green)](https://owasp.org/)

**ACIS** is a secure, lightweight, and user-friendly household financial management web application. Designed for families to track cash flow, manage envelope-based budgets (virtual wallets), and handle expense approvals seamlessly via Web and Telegram Bot.

> 🛡️ **Security First:** Built with strict adherence to **OWASP Top 10** standards, featuring Telegram bot passwordless OTP authentication, composite `(email, phone_number)` identity keys with strict anti-reuse uniqueness, context-aware distributed query tracing, and secure rotating session cookies.

## 🚀 Tech Stack

### Backend (Monorepo: `apps/backend`)
- **Language:** Golang 1.22+
- **Framework:** Gin (HTTP Router)
- **Database:** Native SQL via `sqlx` (PostgreSQL)
- **Architecture:** Clean Architecture (Handler -> Service -> Repository) with Manual Dependency Injection
- **Logging:** `slog` + `lumberjack` with custom `sqlx` wrapper for **Context-Aware Query Tracing** (Trace ID propagation from HTTP to DB layer)
- **Authentication:** Telegram Bot OTP Delivery, Composite Primary Key, Redis rate-limited AES-GCM encrypted OTPs

### Frontend (Monorepo: `apps/frontend`)
- **Framework:** Vue 3 (Composition API, `<script setup>`)
- **Build Tool:** Vite
- **State Management:** Pinia
- **Styling:** Tailwind CSS + DaisyUI (Mobile-first responsive)

### Infrastructure & DevOps
- **Database:** PostgreSQL (Dockerized for local dev, Supabase/Neon for prod)
- **Migration:** Custom Go CLI tool for versioned SQL migrations
- **Deployment:** Vercel (Frontend), Render/Railway (Backend)

## 🏗️ Architecture & Key Features
This project uses a **Modular Monorepo** architecture. 

### 🔐 Telegram Bot Authentication Flow
1. User provides **Email** and **Phone Number** on the ACIS login screen.
2. The combination `(email, phone_number)` serves as the primary key with uniqueness constraints on both columns to prevent cross-account reuse.
3. User receives a 6-digit OTP directly on Telegram (if linked) or via Telegram bot deep link (`t.me/<bot_username>?start=auth_<token>`).
4. Entering the OTP verifies the account, establishes HttpOnly rotating session cookies, and logs the user in.

## 🛠️ Local Development Setup (Windows)

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

### 1. Clone & Start Database & Redis
```bash
git clone https://github.com/Bainandhika/acis.git
cd acis
docker-compose up -d
```

### 2. Run Database Migrations
```bash
cd apps/backend
go run ./cmd/migrate
```

### 3. Start Backend API
```bash
# Copy .env.example to .env and configure Telegram bot token
go run ./cmd/api/main.go
# API runs on http://localhost:8080
```

### 4. Start Frontend
```bash
cd apps/frontend
npm install
npm run dev
# Frontend runs on http://localhost:5173
```
