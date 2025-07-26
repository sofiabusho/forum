-- Users Table: stores registered users
CREATE TABLE IF NOT EXISTS Users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    registration_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Posts Table: stores user posts
CREATE TABLE IF NOT EXISTS Posts (
    post_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    photo_url TEXT,
    content TEXT NOT NULL,
    creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);
-- Comments Table: stores user comments on posts
CREATE TABLE IF NOT EXISTS Comments (
    comment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES Posts(post_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);
-- Categories Table: defines available categories/tags for posts
CREATE TABLE IF NOT EXISTS Categories (
    category_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);
-- PostCategories Table: connects posts to multiple categories (many-to-many)
CREATE TABLE IF NOT EXISTS PostCategories (
    post_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES Posts(post_id),
    FOREIGN KEY (category_id) REFERENCES Categories(category_id)
);
-- LikesDislikes Table: stores user reactions (likes or dislikes) to posts
CREATE TABLE IF NOT EXISTS LikesDislikes (
    likeDislike_id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    vote INTEGER NOT NULL CHECK (vote IN (1, -1)),
    UNIQUE (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES Posts(post_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);
-- CommentLikes Table: stores user reactions (likes or dislikes) to comments
CREATE TABLE IF NOT EXISTS CommentLikes (
    commentlikes_id INTEGER PRIMARY KEY AUTOINCREMENT,
    comment_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    vote INTEGER NOT NULL CHECK (vote IN (1, -1)),
    UNIQUE (comment_id, user_id),
    FOREIGN KEY (comment_id) REFERENCES Comments(comment_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);
-- Notifications Table: stores user notifications
CREATE TABLE IF NOT EXISTS Notifications (
    notification_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('like', 'comment', 'system')),
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    related_post_id INTEGER,
    related_comment_id INTEGER,
    related_user_id INTEGER,
    is_read BOOLEAN DEFAULT FALSE,
    creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(user_id),
    FOREIGN KEY (related_post_id) REFERENCES Posts(post_id),
    FOREIGN KEY (related_comment_id) REFERENCES Comments(comment_id),
    FOREIGN KEY (related_user_id) REFERENCES Users(user_id)
);

-- Images Table: stores uploaded image information
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

-- Sessions Table: manages user sessions (login cookies)
CREATE TABLE IF NOT EXISTS Sessions (
    session_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    cookie_value TEXT UNIQUE NOT NULL,
    expiration_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);

-- Create indexes for performance

-- User table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON Users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON Users(username);

-- Posts table indexes (most important for performance)
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON Posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_creation_date ON Posts(creation_date DESC);
CREATE INDEX IF NOT EXISTS idx_posts_user_date ON Posts(user_id, creation_date DESC);

-- Comments table indexes
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON Comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON Comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_post_date ON Comments(post_id, creation_date ASC);
CREATE INDEX IF NOT EXISTS idx_comments_user_date ON Comments(user_id, creation_date DESC);

-- Categories and PostCategories indexes
CREATE INDEX IF NOT EXISTS idx_post_categories_post_id ON PostCategories(post_id);
CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON PostCategories(category_id);
CREATE INDEX IF NOT EXISTS idx_categories_name ON Categories(name);

-- Likes/Dislikes indexes (for counting likes)
CREATE INDEX IF NOT EXISTS idx_likes_post_id ON LikesDislikes(post_id);
CREATE INDEX IF NOT EXISTS idx_likes_user_id ON LikesDislikes(user_id);
CREATE INDEX IF NOT EXISTS idx_likes_post_vote ON LikesDislikes(post_id, vote);
CREATE INDEX IF NOT EXISTS idx_likes_user_vote ON LikesDislikes(user_id, vote);

-- Comment likes indexes
CREATE INDEX IF NOT EXISTS idx_comment_likes_comment_id ON CommentLikes(comment_id);
CREATE INDEX IF NOT EXISTS idx_comment_likes_user_id ON CommentLikes(user_id);
CREATE INDEX IF NOT EXISTS idx_comment_likes_comment_vote ON CommentLikes(comment_id, vote);

-- Notifications indexes (already exists, but here for completeness)
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON Notifications(user_id, is_read, creation_date DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON Notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_related_post ON Notifications(related_post_id);

-- Sessions indexes (important for authentication)
CREATE INDEX IF NOT EXISTS idx_sessions_cookie_value ON Sessions(cookie_value);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON Sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expiration ON Sessions(expiration_date);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_posts_category_date ON PostCategories(category_id, post_id);
CREATE INDEX IF NOT EXISTS idx_user_posts_likes ON Posts(user_id, post_id);

-- Indexes for Images table
CREATE INDEX IF NOT EXISTS idx_images_user_id ON Images(user_id);
CREATE INDEX IF NOT EXISTS idx_images_filename ON Images(filename);
CREATE INDEX IF NOT EXISTS idx_images_upload_date ON Images(upload_date DESC);

-- Index for Posts with images
CREATE INDEX IF NOT EXISTS idx_posts_image_id ON Posts(image_id);

-- This will store the reference to uploaded images
ALTER TABLE Posts ADD COLUMN image_id INTEGER REFERENCES Images(image_id);


-- Insert starter categories
INSERT OR IGNORE INTO Categories (name) VALUES 
    ('Succulents'),
    ('Tropical Plants'),
    ('Herb Garden'),
    ('Indoor Plants'),
    ('Plant Care Tips'),
    ('Plant Diseases'),
    ('Propagation'),
    ('Flowering Plants');

-- Insert notification types
INSERT INTO Notifications (user_id, type, title, message)
VALUES (1, 'like', 'New Like!', 'Someone liked your post! Check it out!'),
       (1, 'comment', 'New Comment!', 'Someone commented on your post!'),
       (1, 'system', 'Welcome to the Forum!', 'Thank you for joining our plant community!');