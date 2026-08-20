# IMPLEMENTATION PLAN: Migrate ACIS from Telegram OTP Auth to Supabase Auth (Google OAuth)

> **EXECUTION PROTOCOL FOR THE AGENT (Gemini 3.7 Flash)**
> 1. Execute **exactly ONE task per session**. At session start, read ONLY the files listed in that task.
> 2. Before writing any SQL/Go/Vue code, read the referenced existing files and **imitate their style** (naming, error handling, comments).
> 3. **DO NOT** refactor unrelated code, rename existing symbols, add dependencies not listed, or "improve" formatting outside the task's files.
> 4. After every task: run the verification commands. If any check fails, fix ONLY within the task's files. Commit with the exact commit message given.
> 5. If a step references a symbol/column that does not exist, **STOP and report** instead of guessing.
> 6. All identifiers, comments, and commit messages MUST be in English.

## CONTEXT (Read once at the start of Phase 0)

- **Repo:** monorepo. `apps/backend` (Go 1.22, Gin, `sqlx`, Clean Architecture: handler → service → repository, manual DI), `apps/frontend` (Vue 3 + Vite + Pinia + Tailwind/DaisyUI), `apps/bot` (Telegram bot), `apps/migrations` (sequential `.sql`, custom runner at `apps/backend/cmd/migrate/main.go`).
- **Current auth (TO BE REMOVED):** Telegram OTP, `users.phone_number` as PK, AES-GCM OTP in Redis, rotating session cookies, custom JWT.
- **Target auth:** Supabase Auth (Google OAuth). `auth.users` becomes identity source of truth; `users` becomes a profile table keyed by the same UUID. RLS enforced via `auth.uid()`.
- **Telegram bot (RETAINED):** transaction notifications/approvals via existing internal API (`BOT_INTERNAL_SECRET`), plus a new account-linking flow.
- **Infrastructure:** Supabase (Postgres), Upstash (Redis), Koyeb (backend), Cloudflare Pages (frontend).
- **Key architectural rule:** The Go backend uses **two connection pools**:
  - `userPool` → login role `acis_app` (subject to RLS; fails closed). Every user-scoped query runs in an explicit transaction starting with `set_config('request.jwt.claims', …, true)` + `SET LOCAL ROLE authenticated`.
  - `adminPool` → `service_role` (bypasses RLS). Used ONLY by `/internal/*` endpoints (bot) and outbox consumption.

---

## PHASE 0 — Preparation

### TASK-001: Establish current schema mental model
- **Files to read:** all files in `apps/migrations/` (001–014), `apps/backend/cmd/migrate/main.go`, `apps/backend/acis-config.yaml.example`, `apps/backend/.env.example`.
- **Output (report, no code):** a list of all tables with their current columns and constraints, focusing on `users`, `families`, `family_members`, `wallets`, `transactions`, `proposals`, `pending_notifications`.
- **Acceptance:** reported schema matches migration files exactly.
- **Commit:** none.

### TASK-002: Create feature branch
- **Steps:** `git checkout -b feat/supabase-auth-migration`.
- **Commit:** none (branch only).

---

## PHASE 1 — Database (Supabase)

### TASK-101: Write migration 015 (schema refactor)
- **Create file:** `apps/migrations/015_supabase_auth_refactor.sql`
- **Follow the header/marker style of `012_wallet_short_id_and_no_email.sql`** (include `-- +goose Up` / `-- +goose Down`; the runner strips the Down part).
- **Content (use as-is, adjust only if TASK-001 report shows differences):**

```sql
-- Migration: 015_supabase_auth_refactor.sql
-- Description: Refactor users into a Supabase Auth profile table (id = auth.users.id).
-- WARNING: Destructive. Acceptable only because the project currently holds seed/test data.
-- +goose Up

-- 1. Remove seed/test users so the FK to auth.users can be created cleanly.
DELETE FROM users;

-- 2. Drop legacy identity constraints (idempotent guards).
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_id_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_number_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_telegram_chat_id_key;

-- 3. Drop legacy identity columns.
ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
ALTER TABLE users DROP COLUMN IF EXISTS email;
ALTER TABLE users DROP COLUMN IF EXISTS google_id;

-- 4. id becomes the PK and mirrors auth.users.id.
ALTER TABLE users ALTER COLUMN id SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE users ADD CONSTRAINT users_id_fkey
    FOREIGN KEY (id) REFERENCES auth.users(id) ON DELETE CASCADE;

-- 5. telegram_chat_id stays for bot linking: nullable, unique when set.
DROP INDEX IF EXISTS idx_users_telegram_chat_id;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_chat_id
    ON users (telegram_chat_id) WHERE telegram_chat_id IS NOT NULL;

-- 6. Backfill denormalized family_id on transactions (required by RLS policies).
UPDATE transactions t
SET family_id = w.family_id
FROM wallets w
WHERE t.wallet_id = w.id AND t.family_id IS NULL;

-- +goose Down
-- Rollback not supported.
```

- **DO NOT** run it yet.
- **Commit:** `feat(db): refactor users table into supabase auth profile table`

### TASK-102: Write migration 016 (roles, grants, RLS policies)
- **Create file:** `apps/migrations/016_enable_rls.sql`
- **Content requirements, in this exact order:**
  1. Create login role (password set manually later, never in repo):
     ```sql
     DO $$
     BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'acis_app') THEN
           CREATE ROLE acis_app NOINHERIT NOBYPASSRLS LOGIN;
        END IF;
     END $$;
     GRANT authenticated TO acis_app; -- enables SET LOCAL ROLE authenticated
     ```
  2. Grants (deterministic, do not rely on defaults):
     ```sql
     GRANT USAGE ON SCHEMA public TO acis_app, authenticated;
     GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO acis_app, authenticated;
     GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO acis_app, authenticated;
     ALTER DEFAULT PRIVILEGES IN SCHEMA public
        GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO acis_app, authenticated;
     ```
  3. RLS helper (SECURITY DEFINER to avoid recursive RLS on `family_members`):
     ```sql
     CREATE OR REPLACE FUNCTION public.is_family_member(family_uuid uuid)
     RETURNS boolean LANGUAGE sql STABLE SECURITY DEFINER
     SET search_path = pg_catalog, public AS $$
        SELECT EXISTS (
           SELECT 1 FROM public.family_members fm
           WHERE fm.family_id = family_uuid AND fm.user_id = auth.uid()
        );
     $$;
     GRANT EXECUTE ON FUNCTION public.is_family_member(uuid) TO authenticated, acis_app;
     ```
  4. Enable RLS and create policies per table using this rule table. **Imitate the wallets example below for every table.**

| Table | SELECT | INSERT (WITH CHECK) | UPDATE | DELETE |
|---|---|---|---|---|
| `users` | `id = auth.uid()` | `id = auth.uid()` | `id = auth.uid()` | none |
| `families` | `is_family_member(id)` | `created_by = auth.uid()` | `is_family_member(id)` | none |
| `family_members` | `user_id = auth.uid() OR is_family_member(family_id)` | `user_id = auth.uid()` | none | `user_id = auth.uid()` |
| `wallets` | `is_family_member(family_id)` | `is_family_member(family_id)` | `is_family_member(family_id)` | none |
| `transactions` | `is_family_member(family_id)` | `is_family_member(family_id) AND created_by = auth.uid()` | `is_family_member(family_id)` | none |
| `proposals` | `is_family_member((SELECT w.family_id FROM wallets w WHERE w.id = proposals.wallet_id))` | same as SELECT | same as SELECT | none |
| `pending_notifications` | none for authenticated | `is_family_member(family_id)` | none | none |

  Reference example (wallets):
  ```sql
  ALTER TABLE public.wallets ENABLE ROW LEVEL SECURITY;
  CREATE POLICY wallets_select ON public.wallets FOR SELECT TO authenticated
     USING (public.is_family_member(family_id));
  CREATE POLICY wallets_insert ON public.wallets FOR INSERT TO authenticated
     WITH CHECK (public.is_family_member(family_id));
  CREATE POLICY wallets_update ON public.wallets FOR UPDATE TO authenticated
     USING (public.is_family_member(family_id)) WITH CHECK (public.is_family_member(family_id));
  ```
  5. End with `-- +goose Down` + comment `-- Rollback not supported.`
- **DO NOT** add policies for `service_role` (it bypasses RLS by design).
- **DO NOT** run it yet.
- **Commit:** `feat(db): enable row level security with family membership policies`

### TASK-103: Apply migrations + finish DB-side setup (HUMAN-ASSISTED)
- **Steps (agent outputs the exact commands; human runs them):**
  1. `cd apps/backend && DATABASE_URL="<direct connection, port 5432, sslmode=require>" go run ./cmd/migrate`
  2. In Supabase SQL Editor (as postgres), set the app role password: `ALTER ROLE acis_app PASSWORD '<randomly-generated-32-chars>';` → store value as `ACIS_APP_DB_PASSWORD` in the secrets manager.
  3. Build the runtime DSN: `postgresql://acis_app:<ACIS_APP_DB_PASSWORD>@aws-0-ap-northeast-2.pooler.supabase.com:6543/postgres?sslmode=require` (transaction pooler).
- **Acceptance:** `\dt` shows all tables; `SELECT rolname, rolbypassrls FROM pg_roles WHERE rolname='acis_app';` shows `f`.
- **Commit:** none.

### TASK-104: Configure Google OAuth (HUMAN-ASSISTED, agent writes a checklist doc)
- **Create file:** `docs/supabase-google-oauth-setup.md` with this checklist:
  1. Google Cloud Console → OAuth Client ID (Web): Authorized redirect URI = `https://<project-ref>.supabase.co/auth/v1/callback`; Authorized JavaScript origins = `http://localhost:5173` + production Pages URL.
  2. Supabase Dashboard → Authentication → Providers → Google: enable, paste Client ID/Secret.
  3. Supabase Dashboard → Authentication → URL Configuration: Site URL = production Pages URL; Redirect URLs += `http://localhost:5173`.
  4. Supabase Dashboard → Project Settings → API: copy `JWT SECRET` → store as `SUPABASE_JWT_SECRET`.
- **Commit:** `docs: add supabase google oauth setup checklist`

---

## PHASE 2 — Backend (Go) [COMPLETED]

### TASK-201: Config schema update [COMPLETED]
- **Files:** `apps/backend/config/*`, `apps/backend/acis-config.yaml`, `apps/backend/.env.example`.
- **Changes:** added `supabase.jwks_url` (`SUPABASE_JWKS_URL`), `database.app_dsn` (`DATABASE_APP_DSN`), `database.admin_dsn` (`DATABASE_ADMIN_DSN`). Removed `jwt.secret`, `jwt.expiry`, `otp.*` sections. Kept `telegram.*` and `bot.secret`.
- **Commit:** `chore(config): add supabase jwt and dual-pool dsn configuration`

### TASK-202: Dual pool infrastructure + user-scoped transaction helper [COMPLETED]
- **Files:** `apps/backend/infrastructure/database/db.go`, `apps/backend/cmd/api/main.go`.
- **Implement:**
  1. `userDB *sqlx.DB` (from `database.app_dsn`) and `adminDB *sqlx.DB` (from `database.admin_dsn`).
  2. `database.NewDualPool(cfg.AppDSN(), cfg.AdminDSN())`.
  3. `db.WithUserContext(ctx, func(tx *sqlx.Tx) error { ... })` where user claims are extracted directly from `context.Context`.
- **Commit:** `feat(db): add dual connection pools and rls-scoped transaction helper`

### TASK-203: Supabase JWT auth middleware via JWKS (ES256) [COMPLETED]
- **Dependency:** `github.com/lestrrat-go/jwx/v2`.
- **File:** `apps/backend/infrastructure/middleware/supabase_auth.go`.
- **Implementation:**
  1. Asymmetric ES256 key verification against Supabase public JWKS endpoint.
  2. Injects authenticated `userID` and `email` directly into `c.Request.Context()` via `context.WithValue(reqCtx, "auth_user_id", userID)` and `context.WithValue(reqCtx, "auth_user_email", email)` as well as Gin context.
- **Commit:** `feat(auth): verify supabase jwt via jwks es256 middleware`

### TASK-204: Delete legacy auth (Telegram OTP, sessions, OTP crypto) [COMPLETED]
- **Deleted:** `domain/auth/`, AES-GCM OTP cryptography, session cookies, Redis OTP keys, and all legacy login routes.
- **Commit:** `refactor(auth): remove telegram otp and session cookie authentication`

### TASK-205: New auth endpoints (provision + profile) [COMPLETED]
- **Files:** `domain/authentication/handler.go`, `domain/authentication/service.go`, `domain/authentication/repository.go`.
- **Endpoints:**
  1. `POST /api/v1/auth/provision` (behind Supabase JWT middleware): provisions user profile record.
  2. `GET /api/v1/auth/me`: retrieves authenticated user profile and family memberships.
- **Commit:** `feat(auth): add profile provisioning and me endpoints`

### TASK-206: Migrate all user-scoped repositories to WithUserContext [COMPLETED]
- **Files:** `domain/family/repository.go`, `domain/family/service.go`, `domain/transaction/repository.go`, `domain/transaction/service.go`, `domain/authentication/repository.go`.
- **Rule:** Every user-scoped read/write transaction executes inside `s.db.WithUserContext(ctx, func(tx *sqlx.Tx) error { ... })`.
- **Commit:** `fix(arch): properly propagate auth context to rls and fix dual pool wiring`

### TASK-207: Telegram linking endpoint (internal) [COMPLETED]
- **Files:** `domain/bot/handler.go`, `domain/bot/service.go`, `domain/bot/repository.go`.
- **Endpoints:**
  1. `POST /api/v1/telegram/link-code` (requires user JWT): generates 6-char OTP stored in Redis.
  2. `POST /api/v1/internal/telegram/link` (requires `BOT_INTERNAL_SECRET`): consumes code, links `chat_id` to user.
- **Commit:** `feat(bot): add telegram account linking via one-time code`

---

## Backend Architectural Patterns (Phase 2)

Future backend changes MUST adhere strictly to these patterns:

1. **Dual Pool Database Architecture:**
   - Always initialize database via `database.NewDualPool(cfg.AppDSN(), cfg.AdminDSN())`.
   - **`userDB` (via `s.db.WithUserContext`):** Used for ALL user-scoped API operations. Runs inside a transaction where `SELECT set_config('request.jwt.claims', $1, true)` and `SET LOCAL ROLE authenticated` are applied. Supabase Postgres RLS policies (`auth.uid() = ...`) are strictly enforced.
   - **`adminDB` (via `s.db.AdminDB()`):** Used EXCLUSIVELY for background workers (`outboxRepository`), cron jobs, and internal bot webhook endpoints (`/api/v1/internal/telegram/*`) where operations are cross-family or not bound to an individual user JWT.

2. **Context Propagation for RLS:**
   - The Supabase auth middleware (`infrastructure/middleware/supabase_auth.go`) verifies the ES256 JWT and populates `c.Request.Context()`:
     ```go
     reqCtx := c.Request.Context()
     reqCtx = context.WithValue(reqCtx, "auth_user_id", userID)
     reqCtx = context.WithValue(reqCtx, "auth_user_email", email)
     c.Request = c.Request.WithContext(reqCtx)
     ```
   - `db.WithUserContext(ctx, func(tx *sqlx.Tx) error { ... })` extracts `auth_user_id` and `auth_user_email` directly from `ctx`. It **does NOT** take `userID` or `email` as explicit function parameters. If missing from context, it immediately returns an error.
   - Domain repositories and services pass `ctx` to `s.db.WithUserContext` without hardcoding empty strings or mock user IDs.

3. **Authentication & Token Verification:**
   - JWT verification is strictly **ES256 (ECC P-256)** via Supabase public JWKS endpoint (`github.com/lestrrat-go/jwx/v2`).
   - HS256 shared-secret verification is deprecated and disabled.
   - Token claims required: `sub` (User UUID), `aud: authenticated`, `exp` (valid timestamp).

4. **Legacy Auth Removal:**
   - Telegram OTP generation/validation, AES-GCM session cryptography, and session cookies are completely deleted. Do NOT attempt to recreate them.

---

## PHASE 3 — Frontend (Vue 3)

### TASK-301: Supabase client + env
- **Files:** `apps/frontend/package.json` (add `@supabase/supabase-js`), create `apps/frontend/src/lib/supabase.ts`, `.env.example` (`VITE_SUPABASE_URL`, `VITE_SUPABASE_ANON_KEY`).
- **Pattern:**
```ts
import { createClient } from '@supabase/supabase-js'
export const supabase = createClient(
  import.meta.env.VITE_SUPABASE_URL,
  import.meta.env.VITE_SUPABASE_ANON_KEY,
)
```
- **Commit:** `feat(web): add supabase client`

### TASK-302: Auth store (Pinia)
- **Files:** create `stores/auth.ts`; delete old auth store/login state.
- **Implement:** `session` state; `signInWithGoogle()` → `supabase.auth.signInWithOAuth({ provider: 'google', options: { redirectTo: window.location.origin } })`; `signOut()`; `onAuthStateChange` sync; `init()` calls `POST /auth/provision` once per session.
- **Commit:** `feat(web): add pinia auth store with google oauth`

### TASK-303: API client Bearer interceptor
- **Files:** existing API/axios module.
- **Implement:** before each request, attach `Authorization: Bearer <session.access_token>` (refresh via `supabase.auth.getSession()`); on 401 → `signOut()` + redirect to `/login`. Remove `withCredentials` cookie usage.
- **Commit:** `feat(web): attach bearer token to api requests`

### TASK-304: Replace login UI
- **Files:** login view/components.
- **Changes:** remove email+phone form and Telegram deep-link UI; single "Continue with Google" button calling `signInWithGoogle()`; keep DaisyUI styling.
- **Commit:** `feat(web): replace otp login with google oauth button`

---

## PHASE 4 — Telegram Bot

The bot **NO LONGER** handles user authentication, login deep-links, or OTP delivery. Account linking operates via a secure one-time code flow:
1. **Frontend / Webapp:** Authenticated user requests a one-time link code (`POST /api/v1/telegram/link-code` with user Supabase JWT). Backend generates a 6-character code and caches it in Redis for 10 minutes.
2. **Telegram User:** User opens the Telegram bot and sends `/link <code>`.
3. **Bot Backend:** Bot sends `POST /api/v1/internal/telegram/link` (with `BOT_INTERNAL_SECRET`) containing the code and `chat_id`. Backend validates the code and updates `users.telegram_chat_id` using `adminDB`.

### TASK-401: Remove auth flows from bot
- **Files:** `apps/bot/**`.
- **Delete:** `start=auth_<token>` deep-link handling, login OTP delivery, anything calling removed backend auth endpoints.
- **Keep:** transaction/proposal notification handlers and internal API client (`BOT_INTERNAL_SECRET`).
- **Commit:** `refactor(bot): remove authentication flows`

### TASK-402: Add /link command
- **Implement:** `/link <code>` → call `POST /internal/telegram/link` with the sender's `chat_id`; reply with success/failure text.
- **Commit:** `feat(bot): add /link command for account linking`

---

## PHASE 5 — Verification & Security

### TASK-501: RLS negative tests (HUMAN-ASSISTED, agent provides script)
- **Create file:** `scripts/rls_smoke_test.sh` containing curl cases:
  1. No token → 401 on `/auth/me`.
  2. User A token → 200 on own families.
  3. User A token + User B `family_id` → empty result/404 on wallets/transactions/proposals.
  4. `/internal/*` without `BOT_INTERNAL_SECRET` → 401.
- **Acceptance:** all four behave as specified against deployed backend.
- **Commit:** `test: add rls smoke test script`

### TASK-502: OWASP review checklist
- **Create file:** `docs/security-review-supabase-auth.md` verifying: A01 (RLS + repo filters), A02 (no secrets in repo; `acis_app` password manual), A03 (parameterized `set_config`), A07 (Supabase-managed auth, no custom OTP), A05 (env-only config).
- **Commit:** `docs: add security review for supabase auth migration`

---

## PHASE 6 — Deployment

### TASK-601: Koyeb env vars
- **Set:** `DATABASE_APP_DSN` (acis_app pooler DSN), `DATABASE_ADMIN_DSN` (service_role DSN), `SUPABASE_JWKS_URL`, `REDIS_URL` (Upstash), `BOT_INTERNAL_SECRET`, `TELEGRAM_BOT_TOKEN`.
- **Remove:** old OTP/session/JWT vars.

### TASK-602: Cloudflare Pages env vars
- **Set:** `VITE_SUPABASE_URL`, `VITE_SUPABASE_ANON_KEY`. **Never** expose anon key confusion: anon key is safe for frontend; service role key must never appear in `VITE_*`.

---

## DEFINITION OF DONE
- All tasks committed; `go build ./...`, `go test ./...`, `npm run build` green.
- Google OAuth login works end-to-end on localhost and production.
- RLS smoke test passes.
- Telegram bot delivers transaction notifications and `/link` works.