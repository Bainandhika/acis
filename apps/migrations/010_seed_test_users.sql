-- Migration: 010_seed_test_users.sql
-- Description: Seed three test users with isolated OTP bypass, test family, and zero-balance real wallets
-- Date: 2026-08-16

-- +goose Up

-- 1. Seed 3 Test Users
INSERT INTO users (id, email, phone_number, name, telegram_chat_id, created_at, updated_at)
VALUES 
    ('a0000000-0000-0000-0000-000000000001', 'admin@acis.test', '100000001', 'Admin User', 100000001, NOW(), NOW()),
    ('a0000000-0000-0000-0000-000000000002', 'member1@acis.test', '100000002', 'Sarah Member', 100000002, NOW(), NOW()),
    ('a0000000-0000-0000-0000-000000000003', 'member2@acis.test', '100000003', 'Alex Member', 100000003, NOW(), NOW())
ON CONFLICT (email, phone_number) DO UPDATE 
SET 
    name = EXCLUDED.name,
    telegram_chat_id = EXCLUDED.telegram_chat_id,
    updated_at = NOW();

-- 2. Seed Test Family (Monthly Income initialized to 0.00)
INSERT INTO families (id, name, invite_code, monthly_income, created_by, created_at, updated_at)
VALUES 
    ('b0000000-0000-0000-0000-000000000001', 'Smith Family', 'SMTH01', 0.00, 'a0000000-0000-0000-0000-000000000001', NOW(), NOW())
ON CONFLICT (id) DO UPDATE 
SET 
    name = EXCLUDED.name,
    invite_code = EXCLUDED.invite_code,
    monthly_income = 0.00,
    updated_at = NOW();

-- 3. Seed Family Members (Admin + 2 Members)
INSERT INTO family_members (id, family_id, user_id, role, joined_at)
VALUES 
    ('c0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'admin', NOW()),
    ('c0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002', 'member', NOW()),
    ('c0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000003', 'member', NOW())
ON CONFLICT (family_id, user_id) DO UPDATE 
SET 
    role = EXCLUDED.role;

-- 4. Seed Real Wallets with ZERO initial balances
INSERT INTO wallets (id, family_id, name, description, initial_balance, current_balance, minimum_limit, created_by, created_at, updated_at)
VALUES 
    ('d0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'Food & Groceries', 'Monthly food, dining and grocery expenses', 0.00, 0.00, 0.00, 'a0000000-0000-0000-0000-000000000001', NOW(), NOW()),
    ('d0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'Shopping', 'Personal shopping, clothing and household essentials', 0.00, 0.00, 0.00, 'a0000000-0000-0000-0000-000000000001', NOW(), NOW()),
    ('d0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', 'Fuel & Transport', 'Fuel, public transit and vehicle maintenance', 0.00, 0.00, 0.00, 'a0000000-0000-0000-0000-000000000001', NOW(), NOW())
ON CONFLICT (id) DO UPDATE 
SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    initial_balance = EXCLUDED.initial_balance,
    current_balance = EXCLUDED.current_balance,
    minimum_limit = EXCLUDED.minimum_limit,
    updated_at = NOW();

-- +goose Down
-- Rollback test seed
DELETE FROM wallets WHERE family_id = 'b0000000-0000-0000-0000-000000000001';
DELETE FROM family_members WHERE family_id = 'b0000000-0000-0000-0000-000000000001';
DELETE FROM families WHERE id = 'b0000000-0000-0000-0000-000000000001';
DELETE FROM users WHERE email IN ('admin@acis.test', 'member1@acis.test', 'member2@acis.test');
