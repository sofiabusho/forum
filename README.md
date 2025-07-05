FORUM PROJECTPlant Talk Forum
A modern web forum for plant enthusiasts built with Go, SQLite, and vanilla JavaScript.
🌱 Features

User Authentication: Secure registration and login with bcrypt password hashing
Posts & Comments: Create, view, and comment on plant-related discussions
Categories: Organize posts by plant types (Succulents, Tropical, etc.)
Like/Dislike System: React to posts and comments
Filtering: Filter posts by categories, your own posts, or liked posts
Responsive Design: Works on desktop and mobile devices
Docker Ready: Containerized for easy deployment

🚀 Quick Start
Option 1: Using Docker (Recommended)

Clone the repository
bashgit clone https://github.com/yourorg/plant-talk.git
cd plant-talk

Build and run with Docker Compose
bashdocker-compose up --build

Access the forum
Open your browser and visit: http://localhost:8080

Option 2: Local Development

Prerequisites

Go 1.21 or higher
SQLite3


Install dependencies
bashgo mod tidy

Run the application
bashgo run main.go

Access the forum
Open your browser and visit: http://localhost:8080

📁 Project Structure
plant-talk/
├── main.go                     # Main application entry point
├── go.mod                      # Go module dependencies
├── go.sum                      # Go module checksums
├── Dockerfile                  # Docker container configuration
├── docker-compose.yml          # Docker Compose setup
├── internals/
│   ├── database/
│   │   ├── database.go         # Database connection
│   │   ├── sqlstruct.go        # Database structures
│   │   ├── scanfunc.go         # Row scanning functions
│   │   └── table.sql           # Database schema
│   ├── handlers/
│   │   ├── loginHandler.go     # User login
│   │   ├── registerHandler.go  # User registration
│   │   ├── postHandler.go      # Post creation & display
│   │   ├── commentHandler.go   # Comment management
│   │   ├── likeHandler.go      # Like/dislike functionality
│   │   └── filterHandler.go    # Post filtering
│   └── utils/
│       └── utils.go            # Utility functions
└── frontend/
    ├── templates/              # HTML templates
    └── css/                    # Stylesheets and images
🔧 Configuration
Environment Variables

PORT: Server port (default: 8080)
DB_PATH: SQLite database file path (default: ./forum.db)

Database
The application uses SQLite with the following main tables:

Users: User accounts and authentication
Posts: Forum posts
Comments: Post comments
Categories: Post categories
LikesDislikes: Post reactions
CommentLikes: Comment reactions
Sessions: User sessions

🛠 API Endpoints
Authentication

POST /login - User login
POST /register - User registration
GET /logout - User logout

Posts

GET /api/posts - Get all posts (with optional filtering)
POST /new-post - Create new post
POST /api/posts/like - Like/dislike a post

Comments

GET /api/comments?post_id=X - Get comments for a post
POST /api/comments/create - Create new comment
POST /api/comments/like - Like/dislike a comment

Categories

GET /api/categories - Get all categories

Filtering

GET /api/posts?filter=my-posts - Get user's posts (authenticated)
GET /api/posts?filter=my-likes - Get user's liked posts (authenticated)
GET /api/posts?filter=categories&value=CategoryName - Filter by category

🎨 Frontend
The frontend uses:

Bootstrap 5.3.2 for responsive design
Vanilla JavaScript for dynamic functionality
Template partials for shared components (header, footer)

Key Features:

Dynamic post loading with AJAX
Real-time filtering
Responsive navigation
User authentication state management

🧪 Testing

Register a new account

Go to /register
Fill in username, email, and password
Confirm registration


Create posts

Login and click "Create Post"
Select a category and write your post
Publish and view on homepage


Interact with content

Like/dislike posts and comments
Add comments to posts
Filter posts by categories or your activity



🚀 Deployment
Docker Deployment

Build the image
bashdocker build -t plant-talk .

Run the container
bashdocker run -p 8080:8080 -v plant-talk-data:/app/data plant-talk


Production Considerations

Use a reverse proxy (nginx) for SSL termination
Set up proper logging and monitoring
Configure backup for the SQLite database
Consider using PostgreSQL for high-traffic scenarios

🤝 Contributing

Fork the repository
Create a feature branch (git checkout -b feature/amazing-feature)
Commit your changes (git commit -m 'Add amazing feature')
Push to the branch (git push origin feature/amazing-feature)
Open a Pull Request

📝 License
This project is licensed under the MIT License - see the LICENSE file for details.
🙏 Acknowledgments

Inspired by the plant-loving community
Uses Go, SQLite, and modern web technologies