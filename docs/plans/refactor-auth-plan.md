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

## PHASE 2 — Backend (Go)

### TASK-201: Config schema update
- **Files:** `apps/backend/config/*` (read existing config package first), `apps/backend/acis-config.yaml.example`, `apps/backend/.env.example`.
- **Changes:** add `supabase.jwks_url` (bound to `SUPABASE_JWKS_URL`, default `https://gkouiqfystipswyqwdu.supabase.co/auth/v1/.well-known/jwks.json`), `database.app_dsn` (the `acis_app` pooler DSN), `database.admin_dsn` (service_role DSN, port 6543 or 5432). Remove `jwt.secret`, `jwt.expiry`, `otp.*` sections. Keep `telegram.*` and `bot.secret`.
- **DO NOT** touch handler/service code yet (compile may break; that is fixed in TASK-204/205).
- **Commit:** `chore(config): add supabase jwt and dual-pool dsn configuration`

### TASK-202: Dual pool infrastructure + user-scoped transaction helper
- **Files:** create/extend under `apps/backend/infrastructure/` (imitate existing repository wiring style).
- **Implement:**
  1. `userDB *sqlx.DB` (from `database.app_dsn`) and `adminDB *sqlx.DB` (from `database.admin_dsn`).
  2. The mandatory transaction wrapper (copy this pattern; it is the RLS activation point):

```go
// WithUserContext runs fn in a transaction scoped to the authenticated user.
// set_config(..., true) and SET LOCAL ROLE are transaction-scoped: they activate
// Supabase RLS (auth.uid()) for every query inside fn and vanish on commit/rollback.
func (p *Postgres) WithUserContext(ctx context.Context, userID, email string, fn func(tx *sqlx.Tx) error) error {
    tx, err := p.userDB.BeginTxx(ctx, nil)
    if err != nil {
        return err
    }
    defer func() { _ = tx.Rollback() }() // no-op after successful commit

    claims, err := json.Marshal(map[string]string{
        "sub": userID, "role": "authenticated", "email": email, "aud": "authenticated",
    })
    if err != nil {
        return err
    }
    // Parameterized: never string-concatenate JWT data into SQL (OWASP A03).
    if _, err := tx.ExecContext(ctx, `SELECT set_config('request.jwt.claims', $1, true)`, string(claims)); err != nil {
        return err
    }
    if _, err := tx.ExecContext(ctx, `SET LOCAL ROLE authenticated`); err != nil {
        return err
    }
    if err := fn(tx); err != nil {
        return err
    }
    return tx.Commit()
}
```

- **Acceptance:** `go build ./...` passes.
- **Commit:** `feat(db): add dual connection pools and rls-scoped transaction helper`

### TASK-203: Supabase JWT auth middleware via JWKS (ES256)
- **Dependency:** add `github.com/lestrrat-go/jwx/v2`.
- **Create:** `apps/backend/infrastructure/middleware/supabase_auth.go` (imitate existing middleware style).
- **Startup behavior:**
  1. `cache := jwk.NewCache(ctx)`
  2. `cache.Register(jwksURL, jwk.WithMinRefreshInterval(15*time.Minute))`
  3. `cache.Refresh(ctx, jwksURL)` once at boot; fail fast on error.
- **Per-request behavior:**
  1. Read `Authorization` header; require prefix `Bearer `, else 401 JSON.
  2. `set, err := cache.Get(ctx, jwksURL)`; on error 401 JSON.
  3. Parse the token with options: `jwt.WithKeySet(set)`, `jwt.WithValidate(true)`, `jwt.WithAcceptableSkew(5*time.Second)`, `jwt.WithAudience("authenticated")`.
  4. On any parse/validation error → 401 JSON.
  5. Extract `sub` and `email` claims; store in Gin context keys `auth_user_id` and `auth_user_email`; call `c.Next()`.
- **Reference pattern:**

```go
// JWKS-backed verification: validates ES256 (ECC P-256) signatures and
// survives Supabase key rotation via cached refresh. Never store a shared
// JWT signing secret in this service (asymmetric trust model, OWASP A02).
tok, err := jwt.Parse([]byte(raw),
    jwt.WithKeySet(set),              // signature check against current ECC key
    jwt.WithValidate(true),           // enforces exp / nbf
    jwt.WithAcceptableSkew(5*time.Second),
    jwt.WithAudience("authenticated"),
)
```

- **DO NOT:** verify with HS256 or any shared secret; accept `alg=none`; add any token-signing capability to this service.
- **Acceptance:**
  1. `go build ./...` passes.
  2. Unit test `supabase_auth_test.go`: generate an ES256 P-256 keypair, serve its public key as a JWKS via `httptest.Server`, then assert: valid token (aud=authenticated, exp=+1h) passes; expired token → 401; token signed by an unknown key → 401. If jwx v2 signatures differ from the pattern above, read the library README from the Go module cache and adapt the code; do NOT change the acceptance criteria.
- **Commit:** `feat(auth): verify supabase jwt via jwks es256 middleware`

### TASK-204: Delete legacy auth (Telegram OTP, sessions, OTP crypto)
- **Files:** locate via grep: `otp`, `AES`, `GCM`, `session`, `cookie`, `auth_` deep-link symbols in `apps/backend` (handlers, services, repositories, shared utils, Redis OTP keys).
- **Delete** all of them, including their unit tests (e.g., OTP cache tests in `shared`).
- **Keep:** everything under the bot internal API (`BOT_INTERNAL_SECRET` protected routes) and all transaction/wallet/proposal business logic.
- **Rewire** DI in the bootstrap file so it compiles without the deleted services.
- **Acceptance:** `go build ./...` and `go test ./...` pass; `grep -ri "otp" apps/backend --include=*.go` returns nothing except comments referencing this migration.
- **Commit:** `refactor(auth): remove telegram otp and session cookie authentication`

### TASK-205: New auth endpoints (provision + profile)
- **Files:** new handler/service/repository under existing Clean Architecture layers.
- **Endpoints:**
  1. `POST /auth/provision` (behind TASK-203 middleware): `INSERT INTO users (id, name, username, avatar_url) VALUES ($1,$2,$3,$4) ON CONFLICT (id) DO NOTHING` via `WithUserContext`; username fallback = `'user_' || left($1, 8)`; then return the profile row.
  2. `GET /auth/me` (middleware): return profile + family memberships.
- **Acceptance:** curl with a real Supabase token returns 200; without token → 401.
- **Commit:** `feat(auth): add profile provisioning and me endpoints`

### TASK-206: Migrate all user-scoped repositories to WithUserContext
- **Files:** every repository touching `users`, `families`, `family_members`, `wallets`, `transactions`, `proposals`.
- **Rule:** every read/write executes inside `WithUserContext(ctx, userID, email, fn)`. The `userID` comes from Gin context (set by middleware). **Keep existing `WHERE family_id = $x` filters** (defense in depth; RLS is the safety net, not the replacement).
- **adminDB usage:** only for `pending_notifications` consumption used by `/internal/*` bot endpoints.
- **Acceptance:** `go build ./...`, `go test ./...` pass; manual curl: user A token cannot see user B family data (expect empty/404).
- **Commit:** `refactor(repo): scope all repositories to rls user context`

### TASK-207: Telegram linking endpoint (internal)
- **Files:** existing internal router + a new Redis-backed link-code store (Upstash).
- **Endpoints:**
  1. `POST /telegram/link-code` (user middleware): generate 6-char code, `SET link:<code> <user_id> EX 600` in Upstash, return code.
  2. `POST /internal/telegram/link` (`BOT_INTERNAL_SECRET`): body `{code, chat_id}`; pop the key; `UPDATE users SET telegram_chat_id = $1 WHERE id = $2`.
- **Acceptance:** wrong secret → 401; expired code → 404.
- **Commit:** `feat(bot): add telegram account linking via one-time code`

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