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
    type TEXT NOT NULL CHECK (type IN ('like', 'comment', 'mention', 'system')),
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
-- Index for better performance
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON Notifications(user_id, is_read, creation_date DESC);
-- Sessions Table: manages user sessions (login cookies)
CREATE TABLE IF NOT EXISTS Sessions (
    session_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    cookie_value TEXT UNIQUE NOT NULL,
    expiration_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);
-- Optional: insert some starter categories
INSERT
    OR IGNORE INTO Categories (name)
VALUES ('Succulents'),
    ('Tropical Plants'),
    ('Herb Garden'),
    ('Indoor Plants'),
    ('Plant Care Tips'),
    ('Plant Diseases'),
    ('Propagation'),
    ('Flowering Plants');