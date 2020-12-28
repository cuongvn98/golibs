// Package inetauth provides ...
package inetauth

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	sidRw        sync.RWMutex
	sidMemorizer = make(map[string]*userPayload)
)

func init() {
	go func() {
		for {
			// fmt.Println("start cleaning sid repo_memorize")
			time.Sleep(60 * time.Minute)
			sidRw.Lock()
			for key, val := range sidMemorizer {
				if val.IsExpire() {
					delete(sidMemorizer, key)
				}
			}
			sidRw.Unlock()
		}
	}()
}

// UserExchange - iNET payload
type UserExchange struct {
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}

var (
	errSidNotFound = errors.New("sid not found")
)

func exchange(sid string) (UserExchange, error) {
	if sid == "" {
		return UserExchange{}, errSidNotFound
	}
	if val, exists := getFromMemorize(sid); exists {
		return val.User, nil
	}
	code, u, err := exchangeSid(sid)
	if err != nil {
		return u, errors.Wrap(err, "exchange sid")
	}
	if code != http.StatusOK && code != http.StatusCreated {
		return u, errors.New(http.StatusText(code))
	}
	setToMemorize(sid, newUserPayload(u))
	return u, nil
}

func getFromMemorize(sid string) (*userPayload, bool) {
	sidRw.RLock()
	defer sidRw.RUnlock()
	val, ok := sidMemorizer[sid]
	// if ok {
	// 	fmt.Println("get " + val.User.Email)
	// }
	return val, ok
}

type userPayload struct {
	User UserExchange
	Tll  time.Time
}

func (u *userPayload) IsExpire() bool {
	return time.Until(u.Tll).Seconds() <= 0
	// return u.Tll.Sub(time.Now()).Seconds() <= 0
}

func newUserPayload(usr UserExchange) *userPayload {
	return &userPayload{
		User: usr,
		Tll:  time.Now().Add(24 * time.Hour),
	}
}
func setToMemorize(sid string, usr *userPayload) {
	sidRw.Lock()
	defer sidRw.Unlock()
	// fmt.Println("set " + usr.User.Email)
	sidMemorizer[sid] = usr
}

const exchangeSidAPI = "https://sso.inet.vn/api/admin/v1/account/verifysid"

func sidToBody(sid string) io.Reader {
	b, _ := json.Marshal(map[string]string{
		"sid": sid,
	})
	return bytes.NewBuffer(b)
}

func exchangeSid(sid string) (int, UserExchange, error) {
	user := UserExchange{}
	body := sidToBody(sid)
	resp, err := http.Post(exchangeSidAPI, "application/json", body)
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		} else {
			if resp.Body != nil {
				_ = resp.Body.Close()
			}
		}
	}()
	if err != nil {
		return resp.StatusCode, user, err
	}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return http.StatusInternalServerError, user, err
	}
	if user.Email == "" {
		return http.StatusUnauthorized, user, errors.New("email not found")
	}
	return resp.StatusCode, user, nil
}
