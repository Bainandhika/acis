# 🚀 ACIS Production Deployment & Operations Checklist

This document provides a step-by-step checklist to safely configure, deploy, and operate **ACIS** (Aplikasi Catatan Keuangan Istri/Suami) in production.

---

## 1. Secrets & Environment Variables Matrix

Before launching, configure the following secrets in your deployment dashboard (e.g. Render, Railway, Vercel):

| Variable | Description | Required | Example / Recommended Value |
| :--- | :--- | :---: | :--- |
| `GIN_MODE` | Web server environment | Yes | `release` |
| `PORT` | Listening HTTP port | Yes | `8080` (or injected by platform) |
| `DATABASE_URL` | PostgreSQL connection string | Yes | `postgres://user:pass@ep-host.region.neon.tech/acis_db?sslmode=require` |
| `REDIS_URL` | Redis connection string (standalone or Upstash) | Yes | `rediss://default:token@region.upstash.io:6379` |
| `JWT_SECRET` | 32+ byte cryptographically secure string | Yes | Generate with `openssl rand -hex 32` |
| `JWT_EXPIRY` | Short-lived access token validity | Yes | `15m` |
| `OTP_ENCRYPTION_KEY` | 32-byte key for AES-GCM OTP payload encryption | Yes | Generate with `openssl rand -base64 32` |
| `OTP_TTL` | OTP code expiry duration | Yes | `5m` |
| `TELEGRAM_BOT_TOKEN` | BotFather API Token | Yes | `123456789:ABCdefGHIjklMNOpqrsTUVwxyz` |
| `TELEGRAM_WEBHOOK_SECRET` | Secret token verified on webhook delivery | Yes | Generate with `openssl rand -hex 24` |
| `RESEND_API_KEY` | Resend API Key | Yes | `re_123456789abcdef` |
| `RESEND_FROM` | Verified sender email address | Yes | `ACIS <notifications@yourfamilydomain.com>` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed frontend origins | Yes | `https://acis.vercel.app,https://acis.yourdomain.com` |

---

## 2. Database Provisioning & Migrations

1. **PostgreSQL Setup (v15+)**:
   - Provision a PostgreSQL database on **Neon**, **Supabase**, **Render**, or **AWS RDS**.
   - Ensure the `uuid-ossp` or `pgcrypto` extension is permitted.
2. **Execute Migrations**:
   - Run the migration CLI tool before starting the API server:
     ```bash
     export DATABASE_URL="postgres://..."
     ./migrate
     ```
   - All migrations from `001_initial_schema.sql` to `008_add_monthly_income_and_telegram_unique.sql` will be applied atomically within database transactions.

---

## 3. Redis Setup & Outbox Worker Verification

1. **Redis Cache & Outbox Pub/Sub**:
   - ACIS uses Redis for single-use token rotation, AES-GCM encrypted OTP rate limiting, and real-time notification dispatch via `outbox:notify` pub/sub channel.
2. **Fallback Poller**:
   - If Redis is momentarily unavailable, the background outbox poller automatically queries Postgres `pending_notifications` table every 1 minute with `FOR UPDATE SKIP LOCKED` and exponential backoff.

---

## 4. Telegram Bot Webhook Registration

Once your backend is deployed and accessible over HTTPS, register the Telegram Webhook with secret token validation:

```bash
curl -X POST "https://api.telegram.org/bot<TELEGRAM_BOT_TOKEN>/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://<your-backend-api-domain>/api/v1/telegram/webhook",
    "secret_token": "<TELEGRAM_WEBHOOK_SECRET>",
    "allowed_updates": ["message"]
  }'
```

Verify webhook status:
```bash
curl -s "https://api.telegram.org/bot<TELEGRAM_BOT_TOKEN>/getWebhookInfo"
```

---

## 5. Resend Email Domain Verification

1. Log in to [Resend Dashboard](https://resend.com/domains).
2. Add and verify your custom sending domain (DKIM, SPF, and DMARC DNS records).
3. Set `RESEND_FROM` to an address on your verified domain (e.g. `auth@yourdomain.com`).

---

## 6. Frontend Deployment (Vercel / Cloudflare Pages)

1. Set the environment variable in Vercel / Cloudflare Pages:
   ```env
   VITE_API_URL=https://<your-backend-api-domain>/api/v1
   ```
2. Deploy the `apps/frontend` directory.
3. Verify that `CORS_ALLOWED_ORIGINS` on the backend includes the Vercel URL.

---

## 7. Backups & Disaster Recovery

1. **Automated Scheduled Backups**:
   - Add a daily cron job calling `./scripts/backup_db.sh`.
2. **Manual Ad-hoc Backup**:
   - Linux/macOS: `./scripts/backup_db.sh`
   - Windows PowerShell: `.\scripts\backup_db.ps1`
3. **Database Restore Test**:
   ```bash
   gunzip -c backups/acis_backup_YYYYMMDD_HHMMSS.sql.gz | psql "$DATABASE_URL"
   ```

---

## 8. Monitoring & Health Checking

- **Health Check Endpoint:** `GET /api/v1/health`
- Returns detailed status for uptime monitoring:
  ```json
  {
    "status": "ok",
    "version": "1.2.0",
    "environment": "release",
    "database": {
      "status": "healthy",
      "latency_ms": 2
    },
    "redis": {
      "status": "healthy",
      "latency_ms": 1
    }
  }
  ```
- Configure UptimeRobot, BetterStack, or Datadog to poll `/api/v1/health` every 1–5 minutes.
