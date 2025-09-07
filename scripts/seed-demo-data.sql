-- Demo Users (passwords are hashed version of "demo123")
INSERT INTO Users (username, email, password_hash, registration_date) VALUES 
('demo_user', '', '$2a$10$kdxtqkWffXvJ3rcAOKDJbOjBBnSHmuu2GW5wRdQ2gzfD/PPkkXxPa/.og/at2.uheWG/igi', datetime('now')),
('plant_expert', 'expert@plantforum.com', '$2a$10$kdxtqkWffXvJ3rcAOKDJbOjBBnSHmuu2GW5wRdQ2gzfD/.og/at2.uheWG/igi', datetime('now')),
('garden_lover', 'garden@plantforum.com', '$2a$10$kdxtqkWffXvJ3rcAOKDJbOjBBnSHmuu2GW5wRdQ2gzfD/.og/at2.uheWG/igi', datetime('now'));

-- Demo Posts
INSERT INTO Posts (user_id, title, content, creation_date) VALUES 
(1, 'Welcome to Plant Talk Forum!', 'This is a demonstration post showing how our forum works. Feel free to explore all the features!', datetime('now', '-2 days')),
(2, 'Best Succulents for Beginners', 'Here are my top 5 recommended succulents for people just starting their plant journey...', datetime('now', '-1 day')),
(3, 'Watering Tips That Changed My Garden', 'After years of trial and error, these watering techniques transformed my plants...', datetime('now', '-5 hours'));

-- Connect Posts to Categories
INSERT INTO PostCategories (post_id, category_id) VALUES 
(1, 1), -- Welcome post -> Succulents
(2, 1), -- Succulents post -> Succulents  
(2, 5), -- Succulents post -> Plant Care Tips
(3, 5); -- Watering post -> Plant Care Tips
-- Demo Comments
INSERT INTO Comments (post_id, user_id, content, creation_date) VALUES 
(1, 2, 'Great forum! Looking forward to sharing plant knowledge here.', datetime('now', '-1 day')),
(1, 3, 'Thanks for creating this community!', datetime('now', '-12 hours')),
(2, 1, 'Excellent advice on succulents!', datetime('now', '-8 hours')),
(3, 2, 'That watering tip about checking soil moisture - game changer!', datetime('now', '-3 hours'));

-- Demo Likes (optional)
INSERT INTO LikesDislikes (post_id, user_id, vote) VALUES 
(1, 2, 1), -- plant_expert likes welcome post
(1, 3, 1), -- garden_lover likes welcome post  
(2, 1, 1), -- demo_user likes succulents post
(3, 1, 1); -- demo_user likes watering post
