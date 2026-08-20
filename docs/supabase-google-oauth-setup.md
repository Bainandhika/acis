# Supabase Google OAuth Setup Checklist

This document details the configuration steps required to set up Google OAuth authentication with Supabase for ACIS.

---

## 1. Google Cloud Console Configuration
1. Navigate to **Google Cloud Console** → **APIs & Services** → **Credentials**.
2. Create or select an **OAuth 2.0 Client ID** (Application type: **Web application**).
3. Under **Authorized redirect URIs**, add:
   - `https://<project-ref>.supabase.co/auth/v1/callback` (replace `<project-ref>` with your Supabase Project Reference ID)
4. Under **Authorized JavaScript origins**, add:
   - `http://localhost:5173` (Frontend local development)
   - `https://<your-production-pages-domain>.pages.dev` (Cloudflare Pages production URL)
5. Save and copy the **Client ID** and **Client Secret**.

---

## 2. Supabase Authentication Provider Setup
1. Go to **Supabase Dashboard** → Select your Project.
2. Navigate to **Authentication** → **Providers** → **Google**.
3. Toggle Google provider to **Enabled**.
4. Paste the **Client ID** and **Client Secret** obtained from Google Cloud Console.
5. Click **Save**.

---

## 3. Supabase URL Configuration
1. In the Supabase Dashboard, navigate to **Authentication** → **URL Configuration**.
2. Set **Site URL** to your production frontend URL (e.g., `https://<your-production-pages-domain>.pages.dev`).
3. Under **Redirect URLs**, ensure the following are included:
   - `http://localhost:5173`
   - `http://localhost:5173/**`
   - `https://<your-production-pages-domain>.pages.dev/**`
4. Click **Save**.

---

## 4. Retrieve Supabase JWT Secret
1. In the Supabase Dashboard, go to **Project Settings** (gear icon) → **API**.
2. Locate the **JWT Settings** section.
3. Reveal and copy the **JWT Secret**.
4. Store this value securely as `SUPABASE_JWT_SECRET` in your backend environment configuration (and secrets manager for production deployment).
