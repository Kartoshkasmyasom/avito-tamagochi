-- +goose Up

ALTER TABLE pets ADD COLUMN cooldown_ends_at TIMESTAMPTZ;
ALTER TABLE pets ADD COLUMN last_decay_date DATE NOT NULL DEFAULT CURRENT_DATE;

CREATE TABLE daily_tasks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_date DATE NOT NULL,
    activity_type VARCHAR(32) NOT NULL,
    progress INTEGER NOT NULL CHECK (progress >= 0),
    target INTEGER NOT NULL CHECK (target > 0),
    xp_reward INTEGER NOT NULL CHECK (xp_reward >= 0),
    completed_at TIMESTAMPTZ,
    UNIQUE (user_id, task_date, activity_type)
);

CREATE TABLE activity_events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(32) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, id)
);

CREATE TABLE xp_events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source VARCHAR(32) NOT NULL,
    xp_awarded INTEGER NOT NULL CHECK (xp_awarded >= 0),
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE user_rewards (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_type VARCHAR(32) NOT NULL,
    required_level INTEGER NOT NULL CHECK (required_level >= 1),
    unlocked_at TIMESTAMPTZ,
    claimed_at TIMESTAMPTZ,
    UNIQUE (user_id, reward_type)
);

CREATE INDEX daily_tasks_user_date_idx ON daily_tasks(user_id, task_date);
CREATE INDEX xp_events_user_time_idx ON xp_events(user_id, occurred_at);
CREATE INDEX user_rewards_user_idx ON user_rewards(user_id);

-- +goose Down

DROP TABLE user_rewards;
DROP TABLE xp_events;
DROP TABLE activity_events;
DROP TABLE daily_tasks;
ALTER TABLE pets DROP COLUMN last_decay_date;
ALTER TABLE pets DROP COLUMN cooldown_ends_at;
