-- lmao sqlite doesn't support altering fk relationships inline
CREATE TABLE IF NOT EXISTS subscribe_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    feed_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user (id),
    FOREIGN KEY (feed_id) REFERENCES feed (id)
);

CREATE TABLE IF NOT EXISTS saved_item_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    item_url TEXT NOT NULL,
    item_title TEXT NOT NULL,
    archive_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user (id)
);

-- Copy data from old tables to new tables
INSERT INTO
    subscribe_new
SELECT
    id,
    user_id,
    feed_id,
    created_at
FROM
    subscribe;

INSERT INTO
    saved_item_new
SELECT
    id,
    user_id,
    item_url,
    item_title,
    archive_url,
    created_at
FROM
    saved_item;

-- Drop old tables
DROP TABLE subscribe;

DROP TABLE saved_item;

-- Rename new tables to original names
ALTER TABLE subscribe_new
RENAME TO subscribe;

ALTER TABLE saved_item_new
RENAME TO saved_item;

CREATE INDEX IF NOT EXISTS idx_subscribe_user_feed ON subscribe (user_id, feed_id);

CREATE INDEX IF NOT EXISTS idx_saved_item_user ON saved_item (user_id);

CREATE INDEX IF NOT EXISTS idx_saved_item_url ON saved_item (item_url);

CREATE INDEX IF NOT EXISTS idx_saved_item_created ON saved_item (created_at);
