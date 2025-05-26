CREATE TABLE IF NOT EXISTS follows (
    follower_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    PRIMARY KEY (follower_id, following_id),

    -- Check constraint to prevent self-follow
    CONSTRAINT prevent_self_follow CHECK (follower_id <> following_id)
);

