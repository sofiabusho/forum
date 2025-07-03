package database

import (
	"time"
)

type User struct {
	UserID           int
	Username         string
	Firstname        string
	Lastname         string
	Email            string
	PasswordHash     string
	RegistrationDate time.Time
}

type Post struct {
	PostID         int
	UserID         int
	Title          string
	PhotoURL       string
	Content        string
	CreationDate   time.Time
	FormatedDate   string
	Categories     []string
	StatusLiked    string
	StatusDisliked string
	Nbrlike        int
	Nbrdislike     int
	Nbrcomments    int
}

type Comment struct {
	CommentID    int
	PostID       int
	UserID       int
	Username     string
	Firstname    string
	Lastname     string
	Content      string
	NbrLike      int
	NbrDislike   int
	CreationDate time.Time
	Formatdate   string
}

type Category struct {
	CategoryID int
	Name       string
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

type Session struct {
	SessionID      int
	UserID         int
	Cookie_value   string
	ExpirationDate time.Time
}
