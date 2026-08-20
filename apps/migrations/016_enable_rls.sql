-- Migration: 016_enable_rls.sql
-- Description: Create acis_app role, grant permissions, define RLS helper, and enable RLS policies across all tables.
-- +goose Up

-- 1. Create login role (password set manually later in Supabase)
DO $$
BEGIN
   IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'acis_app') THEN
      CREATE ROLE acis_app NOINHERIT NOBYPASSRLS LOGIN;
   END IF;
END $$;
GRANT authenticated TO acis_app; -- enables SET LOCAL ROLE authenticated

-- 2. Grants (deterministic, do not rely on defaults)
GRANT USAGE ON SCHEMA public TO acis_app, authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO acis_app, authenticated;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO acis_app, authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
   GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO acis_app, authenticated;

-- 3. RLS helper (SECURITY DEFINER to avoid recursive RLS on family_members)
CREATE OR REPLACE FUNCTION public.is_family_member(family_uuid uuid)
RETURNS boolean LANGUAGE sql STABLE SECURITY DEFINER
SET search_path = pg_catalog, public AS $$
   SELECT EXISTS (
      SELECT 1 FROM public.family_members fm
      WHERE fm.family_id = family_uuid AND fm.user_id = auth.uid()
   );
$$;
GRANT EXECUTE ON FUNCTION public.is_family_member(uuid) TO authenticated, acis_app;

-- 4. Enable RLS and create policies per table

-- Table: users
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
CREATE POLICY users_select ON public.users FOR SELECT TO authenticated
   USING (id = auth.uid());
CREATE POLICY users_insert ON public.users FOR INSERT TO authenticated
   WITH CHECK (id = auth.uid());
CREATE POLICY users_update ON public.users FOR UPDATE TO authenticated
   USING (id = auth.uid()) WITH CHECK (id = auth.uid());

-- Table: families
ALTER TABLE public.families ENABLE ROW LEVEL SECURITY;
CREATE POLICY families_select ON public.families FOR SELECT TO authenticated
   USING (public.is_family_member(id));
CREATE POLICY families_insert ON public.families FOR INSERT TO authenticated
   WITH CHECK (created_by = auth.uid());
CREATE POLICY families_update ON public.families FOR UPDATE TO authenticated
   USING (public.is_family_member(id)) WITH CHECK (public.is_family_member(id));

-- Table: family_members
ALTER TABLE public.family_members ENABLE ROW LEVEL SECURITY;
CREATE POLICY family_members_select ON public.family_members FOR SELECT TO authenticated
   USING (user_id = auth.uid() OR public.is_family_member(family_id));
CREATE POLICY family_members_insert ON public.family_members FOR INSERT TO authenticated
   WITH CHECK (user_id = auth.uid());
CREATE POLICY family_members_delete ON public.family_members FOR DELETE TO authenticated
   USING (user_id = auth.uid());

-- Table: wallets
ALTER TABLE public.wallets ENABLE ROW LEVEL SECURITY;
CREATE POLICY wallets_select ON public.wallets FOR SELECT TO authenticated
   USING (public.is_family_member(family_id));
CREATE POLICY wallets_insert ON public.wallets FOR INSERT TO authenticated
   WITH CHECK (public.is_family_member(family_id));
CREATE POLICY wallets_update ON public.wallets FOR UPDATE TO authenticated
   USING (public.is_family_member(family_id)) WITH CHECK (public.is_family_member(family_id));

-- Table: transactions
ALTER TABLE public.transactions ENABLE ROW LEVEL SECURITY;
CREATE POLICY transactions_select ON public.transactions FOR SELECT TO authenticated
   USING (public.is_family_member(family_id));
CREATE POLICY transactions_insert ON public.transactions FOR INSERT TO authenticated
   WITH CHECK (public.is_family_member(family_id) AND created_by = auth.uid());
CREATE POLICY transactions_update ON public.transactions FOR UPDATE TO authenticated
   USING (public.is_family_member(family_id)) WITH CHECK (public.is_family_member(family_id));

-- Table: proposals
ALTER TABLE public.proposals ENABLE ROW LEVEL SECURITY;
CREATE POLICY proposals_select ON public.proposals FOR SELECT TO authenticated
   USING (public.is_family_member((SELECT w.family_id FROM public.wallets w WHERE w.id = proposals.wallet_id)));
CREATE POLICY proposals_insert ON public.proposals FOR INSERT TO authenticated
   WITH CHECK (public.is_family_member((SELECT w.family_id FROM public.wallets w WHERE w.id = proposals.wallet_id)));
CREATE POLICY proposals_update ON public.proposals FOR UPDATE TO authenticated
   USING (public.is_family_member((SELECT w.family_id FROM public.wallets w WHERE w.id = proposals.wallet_id)))
   WITH CHECK (public.is_family_member((SELECT w.family_id FROM public.wallets w WHERE w.id = proposals.wallet_id)));

-- Table: pending_notifications
ALTER TABLE public.pending_notifications ENABLE ROW LEVEL SECURITY;
CREATE POLICY pending_notifications_insert ON public.pending_notifications FOR INSERT TO authenticated
   WITH CHECK (true);

-- +goose Down
-- Rollback not supported.
