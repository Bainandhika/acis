-- Migration: 011_seed_test_users.sql
-- Description: Seed three test users with isolated OTP bypass, test family, and zero-balance real wallets
-- Date: 2026-08-16

-- +goose Up

-- 1. Seed 3 Test Users (Phone numbers: 082123456781, 082123456782, 082123456783)
INSERT INTO users (id, username, phone_number, name, telegram_chat_id, created_at, updated_at)
VALUES 
    ('a0000000-0000-0000-0000-000000000001', 'admin_user', '+6282123456781', 'Admin User', 100000001, NOW(), NOW()),
    ('a0000000-0000-0000-0000-000000000002', 'sarah_member', '+6282123456782', 'Sarah Member', 100000002, NOW(), NOW()),
    ('a0000000-0000-0000-0000-000000000003', 'alex_member', '+6282123456783', 'Alex Member', 100000003, NOW(), NOW())
ON CONFLICT (phone_number) DO UPDATE 
SET 
    username = EXCLUDED.username,
    name = EXCLUDED.name,
    telegram_chat_id = EXCLUDED.telegram_chat_id,
    updated_at = NOW();

-- 2. Seed Test Family (Monthly Income initialized to 0.00)
INSERT INTO families (id, name, invite_code, monthly_income, created_by, created_at, updated_at)
VALUES 
    (
        'b0000000-0000-0000-0000-000000000001', 
        'Smith Family', 
        'SMTH01', 
        0.00, 
        (SELECT id FROM users WHERE phone_number = '+6282123456781' LIMIT 1), 
        NOW(), 
        NOW()
    )
ON CONFLICT (id) DO UPDATE 
SET 
    name = EXCLUDED.name,
    invite_code = EXCLUDED.invite_code,
    monthly_income = 0.00,
    created_by = (SELECT id FROM users WHERE phone_number = '+6282123456781' LIMIT 1),
    updated_at = NOW();

-- 3. Seed Family Members (Admin + 2 Members)
INSERT INTO family_members (id, family_id, user_id, role, joined_at)
VALUES 
    ('c0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', (SELECT id FROM users WHERE phone_number = '+6282123456781' LIMIT 1), 'admin', NOW()),
    ('c0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', (SELECT id FROM users WHERE phone_number = '+6282123456782' LIMIT 1), 'member', NOW()),
    ('c0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', (SELECT id FROM users WHERE phone_number = '+6282123456783' LIMIT 1), 'member', NOW())
ON CONFLICT (id) DO UPDATE 
SET 
    user_id = EXCLUDED.user_id,
    role = EXCLUDED.role;

-- 4. Seed Real Initial Wallets (Balances set to 0.00)
INSERT INTO wallets (id, family_id, name, description, current_balance, minimum_limit, created_by, created_at, updated_at)
VALUES 
    ('d0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'Food & Groceries', 'Daily meals and weekly grocery shopping', 0.00, 0.00, (SELECT id FROM users WHERE phone_number = '+6282123456781' LIMIT 1), NOW(), NOW()),
    ('d0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'Shopping & Leisure', 'Family recreational and shopping envelope', 0.00, 0.00, (SELECT id FROM users WHERE phone_number = '+6282123456781' LIMIT 1), NOW(), NOW()),
    ('d0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', 'Fuel & Transport', 'Vehicle fuel, public transit, and parking', 0.00, 0.00, (SELECT id FROM users WHERE phone_number = '+6282123456781' LIMIT 1), NOW(), NOW())
ON CONFLICT (id) DO UPDATE 
SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    created_by = (SELECT id FROM users WHERE phone_number = '+6282123456781' LIMIT 1),
    updated_at = NOW();

-- +goose Down
-- Rollback script
