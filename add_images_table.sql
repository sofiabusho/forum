CREATE TABLE IF NOT EXISTS Images (
    image_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    filename TEXT UNIQUE NOT NULL,
    original_name TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    file_type TEXT NOT NULL CHECK (file_type IN ('JPEG', 'PNG', 'GIF')),
    image_url TEXT NOT NULL,
    thumbnail_url TEXT,
    upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);

CREATE INDEX IF NOT EXISTS idx_images_user_id ON Images(user_id);
CREATE INDEX IF NOT EXISTS idx_images_filename ON Images(filename);
CREATE INDEX IF NOT EXISTS idx_images_upload_date ON Images(upload_date DESC);

ALTER TABLE Posts ADD COLUMN image_id INTEGER REFERENCES Images(image_id);
CREATE INDEX IF NOT EXISTS idx_posts_image_id ON Posts(image_id);
