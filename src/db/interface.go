package db

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: expose more APIs if needed
type IDB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	PingContext(ctx context.Context) error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	// Conn(ctx context.Context) (*Conn, error)
	// Driver() driver.Driver
	// SetConnMaxIdleTime(d time.Duration)
	// SetConnMaxLifetime(d time.Duration)
	// SetMaxIdleConns(n int)
	// SetMaxOpenConns(n int)
	// Stats() DBStats
}

type IDBFunctions interface {
	Init(ctx context.Context, adminName, adminPwd string, config *SiteConfig) error
	IDBLockable
	IUserDB
	IFileDB
	IUploadDB
	ISharingDB
	IConfigDB
}

type IDBLockable interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type IUserDB interface {
	IsInited() bool
	AddUser(ctx context.Context, user *User) error
	DelUser(ctx context.Context, id uint64) error
	GetUser(ctx context.Context, id uint64) (*User, error)
	GetUserByName(ctx context.Context, name string) (*User, error)
	SetPwd(ctx context.Context, id uint64, pwd string) error
	SetInfo(ctx context.Context, id uint64, user *User) error
	SetPreferences(ctx context.Context, id uint64, prefers *Preferences) error
	SetUsed(ctx context.Context, id uint64, incr bool, capacity int64) error
	ResetUsed(ctx context.Context, id uint64, used int64) error
	ListUsers(ctx context.Context) ([]*User, error)
	ListUserIDs(ctx context.Context) (map[string]string, error)
	AddRole(role string) error
	DelRole(role string) error
	ListRoles() (map[string]bool, error)
}

type IFileDB interface {
	AddFileInfo(ctx context.Context, userId uint64, itemPath string, info *FileInfo) error
	// DelFileInfo(ctx context.Context, itemPath string) error
	DelFileInfo(ctx context.Context, userId uint64, itemPath string) error
	GetFileInfo(ctx context.Context, userId uint64, itemPath string) (*FileInfo, error)
	SetSha1(ctx context.Context, userId uint64, itemPath, sign string) error
	MoveFileInfos(ctx context.Context, userID uint64, oldPath, newPath string, isDir bool) error
	ListFileInfos(ctx context.Context, itemPaths []string) (map[string]*FileInfo, error)
}
type IUploadDB interface {
	AddUploadInfos(ctx context.Context, userId uint64, tmpPath, filePath string, info *FileInfo) error
	DelUploadingInfos(ctx context.Context, userId uint64, realPath string) error
	// MoveUploadingInfos(ctx context.Context, userId uint64, uploadPath, itemPath string) error
	SetUploadInfo(ctx context.Context, user uint64, filePath string, newUploaded int64) error
	GetUploadInfo(ctx context.Context, userId uint64, filePath string) (string, int64, int64, error)
	ListUploadInfos(ctx context.Context, user uint64) ([]*UploadInfo, error)
}

type ISharingDB interface {
	IsSharing(ctx context.Context, userId uint64, dirPath string) bool
	GetSharingDir(ctx context.Context, hashID string) (string, error)
	AddSharing(ctx context.Context, userId uint64, dirPath string) error
	DelSharing(ctx context.Context, userId uint64, dirPath string) error
	ListUserSharings(ctx context.Context, userId uint64) (map[string]string, error)
}

type IConfigDB interface {
	SetClientCfg(ctx context.Context, cfg *ClientConfig) error
	GetCfg(ctx context.Context) (*SiteConfig, error)
}
