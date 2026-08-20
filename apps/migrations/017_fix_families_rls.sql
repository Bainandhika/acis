-- Migration: 017_fix_families_rls.sql
-- Description: Update families RLS policies to allow creators to select and update their family rows
-- +goose Up

DROP POLICY IF EXISTS families_select ON public.families;
CREATE POLICY families_select ON public.families FOR SELECT TO authenticated
   USING (created_by = auth.uid() OR public.is_family_member(id));

DROP POLICY IF EXISTS families_update ON public.families;
CREATE POLICY families_update ON public.families FOR UPDATE TO authenticated
   USING (created_by = auth.uid() OR public.is_family_member(id))
   WITH CHECK (created_by = auth.uid() OR public.is_family_member(id));

-- +goose Down
DROP POLICY IF EXISTS families_select ON public.families;
CREATE POLICY families_select ON public.families FOR SELECT TO authenticated
   USING (public.is_family_member(id));

DROP POLICY IF EXISTS families_update ON public.families;
CREATE POLICY families_update ON public.families FOR UPDATE TO authenticated
   USING (public.is_family_member(id))
   WITH CHECK (public.is_family_member(id));
