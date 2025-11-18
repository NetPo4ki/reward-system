INSERT INTO tasks (code, name, points, active)
VALUES
    ('subscribe_tg', 'Subscribe to Telegram channel', 50, TRUE),
    ('follow_twitter', 'Follow on Twitter/X', 40, TRUE),
    ('enter_referral', 'Enter referral code', 20, TRUE)
ON CONFLICT (code) DO NOTHING;