package database

import (
	"time"
)

type User struct {
	UserID           int
	Username         string
	Email            string
	PasswordHash     string
	RegistrationDate time.Time
	ResetToken       *string
}

type UserProfile struct {
	UserID        int    `json:"userId"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	JoinDate      string `json:"joinDate"`
	PostCount     int    `json:"postCount"`
	CommentCount  int    `json:"commentCount"`
	LikesGiven    int    `json:"likesGiven"`
	LikesReceived int    `json:"likesReceived"`
	DislikesGiven    int    `json:"dislikesGiven"`
	DislikesReceived int    `json:"dislikesReceived"`
	ProfileImage  string `json:"profileImage,omitempty"`
	Bio           string `json:"bio"`
}

type UserActivity struct {
	RecentPosts    []PostResponse    `json:"recentPosts"`
	RecentComments []CommentActivity `json:"recentComments"`
}

type CommentActivity struct {
	ID        int    `json:"id"`
	PostID    int    `json:"postId"`
	PostTitle string `json:"postTitle"`
	Content   string `json:"content"`
	TimeAgo   string `json:"timeAgo"`
}

type Post struct {
	PostID         int
	UserID         int
	Title          string
	PhotoURL       string
	Content        string
	ImageID        *int
	CreationDate   time.Time
	FormatedDate   string
	Categories     []string
	StatusLiked    string
	StatusDisliked string
	Nbrlike        int
	Nbrdislike     int
	Nbrcomments    int
}

type PostResponse struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Author       string   `json:"author"`
	TimeAgo      string   `json:"timeAgo"`
	Tags         []string `json:"tags"`
	Comments     int      `json:"comments"`
	Likes        int      `json:"likes"`
	Dislikes     int      `json:"dislikes"` 
	Excerpt      string   `json:"excerpt"`
	ImageURL     string   `json:"imageUrl,omitempty"`
	ThumbnailURL string   `json:"thumbnailUrl,omitempty"`
	UserVote     int      `json:"userVote,omitempty"` 
	IsAuthor     bool     `json:"isAuthor,omitempty"`
}

type Image struct {
	ImageID      int       `json:"id"`
	UserID       int       `json:"userId"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"originalName"`
	FileSize     int64     `json:"fileSize"`
	FileType     string    `json:"fileType"`
	ImageType    string    `json:"imageType"`
	ImageURL     string    `json:"imageUrl"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	UploadDate   time.Time `json:"uploadDate"`
}

type ImageResponse struct {
	ID                int    `json:"id"`
	Filename          string `json:"filename"`
	OriginalName      string `json:"originalName"`
	FileSize          int64  `json:"fileSize"`
	FileSizeFormatted string `json:"fileSizeFormatted"`
	FileType          string `json:"fileType"`
	ImageURL          string `json:"imageUrl"`
	ThumbnailURL      string `json:"thumbnailUrl"`
	UploadDate        string `json:"uploadDate"`
}

type Comment struct {
	CommentID    int
	PostID       int
	UserID       int
	Username     string
	Content      string
	NbrLike      int
	NbrDislike   int
	CreationDate time.Time
	Formatdate   string
}

type CommentResponse struct {
	ID           int    `json:"id"`
	PostID       int    `json:"postId"`
	Author       string `json:"author"`
	Content      string `json:"content"`
	TimeAgo      string `json:"timeAgo"`
	LikeCount    int    `json:"likeCount"`
	DislikeCount int    `json:"dislikeCount"`
	UserVote     int    `json:"userVote"`
	IsAuthor     bool   `json:"isAuthor"`
}

type Category struct {
	CategoryID int
	Name       string
}

type CategoryResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type PostCategory struct {
	PostID     int
	CategoryID int
}

type LikeDislike struct {
	LikeDislikeID   int
	PostID          int
	CommentID       int
	UserID          int
	LikeDislikeType string
	CreationDate    time.Time
}

type CommentLike struct {
	CommentID    int
	PostID       int
	UserID       int
	NbrLike      int
	NbrDislike   int
	CreationDate time.Time
}

type LikeResponse struct {
	Success      bool `json:"success"`
	LikeCount    int  `json:"likeCount"`
	DislikeCount int  `json:"dislikeCount"`
	UserVote     int  `json:"userVote"` // 1 for like, -1 for dislike, 0 for none
}

type Session struct {
	SessionID      int
	UserID         int
	Cookie_value   string
	ExpirationDate time.Time
}

type Notification struct {
	NotificationID   int       `json:"id"`
	UserID           int       `json:"userId"`
	Type             string    `json:"type"`
	Title            string    `json:"title"`
	Message          string    `json:"message"`
	RelatedPostID    *int      `json:"relatedPostId,omitempty"`
	RelatedCommentID *int      `json:"relatedCommentId,omitempty"`
	RelatedUserID    *int      `json:"relatedUserId,omitempty"`
	IsRead           bool      `json:"isRead"`
	CreationDate     time.Time `json:"creationDate"`
	TimeAgo          string    `json:"timeAgo"`
}

type NotificationResponse struct {
	Unread []Notification `json:"unread"`
	Read   []Notification `json:"read"`
}
