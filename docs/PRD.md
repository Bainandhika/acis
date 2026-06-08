# ACIS - Product Requirements Document (PRD)
**Version:** 1.1.0 (Updated with Architecture Details)
**Last Updated:** 2026-06-07
**Author:** Bagas

## 1. Product Overview
ACIS (Aplikasi Catatan Keuangan Istri/Suami) adalah aplikasi manajemen keuangan rumah tangga berbasis web yang memungkinkan keluarga melacak cash flow, mengelola anggaran berbasis "amplop" (dompet virtual), dan mengajukan pengeluaran. Aplikasi ini mengutamakan keamanan data (OWASP Top 10 compliant), kemudahan akses via web, serta integrasi Telegram Bot.

## 2. Target Users & Roles
- **Admin (Kepala Keluarga):** Full access. Mengelola dompet, alokasi dana, approve/reject pengajuan, dan konfigurasi sistem.
- **Member (Anggota Keluarga):** View-only dashboard. Dapat mengajukan pengeluaran (proposal) yang menunggu approval Admin.

## 3. Core Features (MVP)
### 3.1. Authentication & Authorization (Passwordless)
- **Google OAuth (1-Click):** Menggunakan Google Identity Services.
- **Email OTP (Passwordless):** 6-digit code, expire in 5 mins, dikirim via Resend/SendGrid.
- **Session Management:** JWT disimpan di `HttpOnly`, `Secure`, `SameSite=Strict` cookies (Mencegah XSS).
- **Role-Based Access Control (RBAC):** Middleware enforcement untuk Admin vs Member routes.

### 3.2. Family & Wallet Management (Envelope System)
- **Family Grouping:** User bisa membuat atau bergabung dengan "Keluarga" menggunakan Invite Code.
- **Virtual Wallets:** Admin dapat membuat "Dompet" (misal: Makan, Nabung, Hiburan).
- **Budget Constraints:** Setiap dompet memiliki `initial_balance` dan `minimum_limit`. Total alokasi tidak boleh melebihi pendapatan.

### 3.3. Transaction & Proposal System
- **Admin:** Input transaksi masuk/keluar langsung (langsung memotong saldo dompet via DB Transaction).
- **Member:** Mengajukan pengeluaran (Status: `PENDING`).
- **Approval Flow:** Admin Approve/Reject proposal. Jika di-approve, saldo dompet terpotong otomatis dalam satu transaksi DB (Atomic).

### 3.4. Telegram Bot Integration (Phase 2)
- Input transaksi via chat (format: `/catat [dompet] [nominal] [keterangan]`).
- Cek saldo via chat (`/saldo`).
- *Cron job* reminder jika saldo dompet <= `minimum_limit`.

## 4. Technical Architecture & Decisions
### 4.1. Monorepo Structure
Project menggunakan struktur Monorepo sederhana tanpa tool berat (seperti Nx/Turborepo) untuk memudahkan deployment terpisah (Vercel untuk Frontend, Render untuk Backend).
```text
acis/
├── apps/backend/      # Golang API (Gin + sqlx)
├── apps/frontend/     # Vue 3 SPA (Vite + Pinia)
├── migrations/        # Shared SQL Migration files
└── docker-compose.yml # Local PostgreSQL setup