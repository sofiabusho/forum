package database

import (
	"database/sql"
)

type Table interface {
	ScanRows(rows *sql.Rows) error
}

// User structure
func (u *User) ScanRows(rows *sql.Rows) error {
	return rows.Scan(&u.UserID, &u.Username, &u.Email, &u.PasswordHash, &u.RegistrationDate)
}

// Post structure
func (p *Post) ScanRows(rows *sql.Rows) error {
	return rows.Scan(&p.PostID, &p.UserID, &p.Title, &p.PhotoURL, &p.Content, &p.CreationDate)
}

// Comment structure
func (c *Comment) ScanRows(rows *sql.Rows) error {
	return rows.Scan(&c.CommentID, &c.PostID, &c.UserID, &c.Content, &c.CreationDate)
}

// CAtegory structure
func (cat *Category) ScanRows(rows *sql.Rows) error {
	return rows.Scan(&cat.CategoryID, &cat.Name)
}

// PstCategory structure
func (pc *PostCategory) ScanRows(rows *sql.Rows) error {
	return rows.Scan(&pc.PostID, &pc.CategoryID)
}

// LikeDislike structure
func (ld *LikeDislike) ScanRows(rows *sql.Rows) error {
	return rows.Scan(&ld.LikeDislikeID, &ld.PostID, &ld.CommentID, &ld.UserID, &ld.LikeDislikeType, &ld.CreationDate)
}

// Session structure
func (s *Session) ScanRows(rows *sql.Rows) error {
	return rows.Scan(&s.SessionID, &s.UserID, &s.Cookie_value, &s.ExpirationDate)
}
