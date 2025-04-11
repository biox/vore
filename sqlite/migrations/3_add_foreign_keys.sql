ALTER TABLE subscribe ADD CONSTRAINT fk_subscribe_user FOREIGN KEY (user_id) REFERENCES user (id);

ALTER TABLE subscribe ADD CONSTRAINT fk_subscribe_feed FOREIGN KEY (feed_id) REFERENCES feed (id);

ALTER TABLE saved_item ADD CONSTRAINT fk_saved_item_user FOREIGN KEY (user_id) REFERENCES user (id);

CREATE INDEX IF NOT EXISTS idx_subscribe_user_feed ON subscribe (user_id, feed_id);

CREATE INDEX IF NOT EXISTS idx_saved_item_user ON saved_item (user_id);

CREATE INDEX IF NOT EXISTS idx_saved_item_url ON saved_item (item_url);

CREATE INDEX IF NOT EXISTS idx_saved_item_created ON saved_item (created_at);
