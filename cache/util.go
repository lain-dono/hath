package cache

import (
	"os"
	"path"
	"syscall"
	"time"
)

func currentTimeMs() int64 {
	now := time.Now()
	return now.Unix()*1000 + int64(now.Nanosecond()/(1000*1000))
}

func currentTimeS() int64 {
	return time.Now().Unix()
}

func freeSpace(wd string) int64 {
	var stat syscall.Statfs_t
	//wd, err := os.Getwd()
	syscall.Statfs(wd, &stat)
	return int64(stat.Bavail * uint64(stat.Bsize))
}
func getFreeSpace(*os.File) int {
	return 44444
}

func removeFile(args ...string) error {
	return os.Remove(path.Join(args...))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type settings struct {
}

func (s *settings) isForceDirty() bool {
	return false
}
func (s *settings) isVerifyCache() bool {
	return false
}
func (s *settings) isUseLessMemory() bool {
	return false
}
func (s *settings) isStaticRange(fileid string) bool {
	return false
}
func (s *settings) isSkipFreeSpaceCheck() bool {
	return false
}
func (s *settings) DataDirAhsolutePath() string {
	return "data"
}
func (s *settings) getDiskLimitBytes() int64 {
	return 15968745168
}
func (s *settings) getDiskMinRemainingBytes() int {
	return 48941236
}

var Settings = &settings{}
