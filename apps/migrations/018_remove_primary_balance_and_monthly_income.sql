-- Migration: 018_remove_primary_balance_and_monthly_income.sql
-- Description: Drop primary_balance and monthly_income columns from families table and allow reading member names
-- +goose Up

ALTER TABLE public.families DROP COLUMN IF EXISTS primary_balance;
ALTER TABLE public.families DROP COLUMN IF EXISTS monthly_income;

-- Update users_select policy so family members can read member names in the member list
DROP POLICY IF EXISTS users_select ON public.users;
CREATE POLICY users_select ON public.users FOR SELECT TO authenticated
   USING (true);

-- +goose Down
ALTER TABLE public.families ADD COLUMN IF NOT EXISTS monthly_income NUMERIC(15, 2) DEFAULT 0.00;
ALTER TABLE public.families ADD COLUMN IF NOT EXISTS primary_balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00;
