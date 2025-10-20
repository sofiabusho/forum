# FORUM PROJECT

🪴🌵🌱🌷🌿 Plant Talk Forum 🌿🌷🌱🌵🪴

A web forum for plant enthusiasts to connect, share knowledge, and build community. built with Go, SQLite, and vanilla JavaScript.

---

**Forum,**
**Forum-Authentication,**
**Forum-Image-Upload, and**
**Forum-Advanced-Features**

## 🌱 Features

**- User Authentication & Authorization:** Secure registration and login with bcrypt password hashing

**- Posts & Comments System:** Create, view, edit, and delete posts and comments

**- Multi-Category Support:** Organize posts by plant types (Succulents, Tropical Plants, etc.)

**- Like/Dislike System:** React to both posts and comments with voting functionality

**- Advanced Filtering:** Filter posts by categories, user's own posts, or liked posts

**- Session Management:** Secure cookie-based sessions with expiration dates

---

### Authentication Methods (OAuth Integration)

Traditional Registration: Email, username, and password registration

Google OAuth: Sign in with Google account

GitHub OAuth: Sign in with GitHub account

Password Recovery: Forgot password functionality with secure reset tokens

Session Security: UUID-based session tokens with automatic expiration

### Image Upload System

Multi-Format Support: JPEG, PNG, and GIF image uploads

Size Validation: 20MB maximum file size with proper error handling

Thumbnail Generation: Automatic thumbnail creation for optimized display

Image Management: Users can upload, view, and delete their own images

Post Integration: Attach images to posts with preview functionality

### Advanced Features

Real-time Notifications: Notify users when their content is liked/disliked or commented on

Activity Tracking: Comprehensive user activity page showing:

User's created posts

Posts where user left likes/dislikes

Comments made by the user with context

Content Management: Edit and delete posts and comments

Notification Management: Mark notifications as read, view notification history

### Technical Features

Responsive Design: Mobile-friendly interface using Bootstrap 5.3.2

Docker Ready: Containerized application with Docker Compose support

Database Optimization: Comprehensive indexing for performance

Error Handling: Proper HTTP status codes and user-friendly error messages

Security Best Practices: CSRF protection, input validation, and secure session management

## Usage

**Option 1: Using Docker (Recommended)**

1. Clone the repository

```bash
git clone https://platform.zone01.gr/git/ttarara/forum

cd forum
```

2. Build and run with Docker Compose

```bash
docker compose up -d --build
docker images
docker ps -a
docker compose down -v
```

3. Access the forum

```bash
Open your browser and visit: http://localhost:8080

```

**Option 2: Local Development**

Prerequisites

Go 1.21 or higher
SQLite3

1. Install dependencies
```bash
mod tidy
```

2. Run the application

```bash
go run .
```
3. Access the forum
```bash
Open your browser and visit: http://localhost:8080
```

## Database Schema

The application uses SQLite with the following optimized database structure:

### Core Tables

**_Users:_** User accounts, authentication, and profile information
**_Posts:_** Forum posts with image references
**_Comments:_** Post comments and replies
**_Categories:_** Available post categories
**_PostCategories:_** Many-to-many relationship for post categorization

### Interaction Tables

**_LikesDislikes:_** Post voting system
**_CommentLikes:_** Comment voting system
**_Sessions:_** Secure session management
**_Images:_** Image upload metadata and file tracking
**_Notifications:_** User notification system

## Performance Features

- Comprehensive database indexing for optimal query performance
- WAL (Write-Ahead Logging) mode for concurrent read/write operations
- Optimized composite indexes for complex filtering operations

---

**🎨 Frontend Technology**
The frontend implementation uses:
```bash
Bootstrap 5.3.2: Responsive design framework
Vanilla JavaScript: Dynamic functionality and AJAX requests
Template System: Reusable HTML components (header, footer)
Responsive Design: Mobile-first approach with adaptive layouts
```

Key Frontend Features
```bash
Dynamic post loading with infinite scroll
Real-time notification updates
Image upload with drag-and-drop interface
Responsive navigation with user state management
Form validation and error handling
Modal dialogs for confirmations
```
---

**🧪 Testing the Application**
User Registration & Authentication
```bash
Traditional Registration: Use /register with email, username, and password
OAuth Login: Test Google and GitHub authentication flows
Password Recovery: Use forgot password functionality
Session Management: Verify automatic logout on session expiration
```
Content Creation & Management
```bash
Create Posts: Test post creation with categories and images
Image Upload: Upload JPEG, PNG, and GIF files (test size limits)
Comment System: Add comments and replies to posts
Edit/Delete: Modify your own content
Like/Dislike: React to posts and comments
```
Advanced Features
```bsh
Filtering: Test category filters, "my posts", and "my likes"
Notifications: Create interactions and check notification system
Activity Tracking: Review your activity page
Image Management: Upload, view, and delete images
```
---

**🚀 Deployment**

Docker Deployment

```bash Build the image
docker build -t forum .

# Run with volume for data persistence
docker run -p 8080:8080 -v forum-data:/app/data plant-talk-forum

```

---

**Production Considerations**
```bash
sqlite3 - Database driver
bcrypt - Password hashing
google/uuid - UUID generation
golang.org/x/oauth2 - OAuth2 implementation
golang.org/x/image - Image processing
google.golang.org/api - Google API client
```
---

**📝 Project Requirements Compliance**

This forum implementation satisfies all requirements from the original project specifications:

```bash
✅ Basic Forum: SQLite database, authentication, posts, comments, likes, filtering
✅ Forum-Authentication: Google and GitHub OAuth integration
✅ Forum-Image-Upload: JPEG, PNG, GIF support with 20MB size limit
✅ Forum-Advanced-Features: Notifications, activity tracking, edit/delete functionality
✅ Docker: Complete containerization with Docker Compose
✅ Security: Bcrypt hashing, session management, input validation
✅ Performance: Database optimization, indexing, efficient queries
```

**📄 License**
This project is licensed under the MIT License - see the LICENSE file for details.

---

## ✍️ Authors

🌸 Theocharoula Tarara 🪴

🌱 Sofia Busho 🌺

---

                   Enjoy being part of the Plant Talk Community!

                                      🌵
