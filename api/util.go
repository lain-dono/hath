package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var ClientKeyRegexp = regexp.MustCompile(fmt.Sprintf("^[a-zA-Z0-9]{%d}", CLIENT_KEY_LENGTH))

type Login struct {
	Id             int64
	Key            string
	loginValidated bool
}

func (login *Login) CheckId() bool {
	return login.Id > 1000
}
func (login *Login) CheckKey() bool {
	return ClientKeyRegexp.MatchString(login.Key)
}
func (login *API) IsLoginValidated() bool {
	return login.loginValidated
}

type Time struct {
	serverTimeDelta time.Duration
}

func (t *Time) CheckTimeDiff(testtime int64) bool {
	return abs(testtime-t.ServerTime()) > MAX_KEY_TIME_DRIFT
}
func (t *Time) ServerTime() int64 {
	return time.Now().Add(t.serverTimeDelta).Unix()
}
func (t *Time) SetServerTime(unix int64) {
	t.serverTimeDelta = time.Unix(unix, 0).Sub(time.Now())
	log.Debugf("Setting altered: serverTimeDelta=%s", t.serverTimeDelta)
}

func trim(from []string) (to []string) {
	for _, v := range from {
		v = strings.TrimSpace(v)
		if v != "" {
			to = append(to, v)
		}
	}
	return
}

func abs(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}
