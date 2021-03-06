package handler

import (
	"net/http"
	"strings"
	"sync"

	"github.com/Tomasen/realip"
	"github.com/dragonflylee/gocms/model"
)

var (
	tokenMap   = make(map[int64]string)
	tokenMutex sync.RWMutex
)

// Login 登录页
func Login(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessName)
	if err == nil {
		if _, exist := session.Values["user"]; exist {
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
	}
	if r.Method == http.MethodGet {
		t.ExecuteTemplate(w, "login.tpl", r.Referer())
		return
	}
	if err = r.ParseForm(); err != nil {
		jRsp(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	email := strings.TrimSpace(r.PostForm.Get("username"))
	password := strings.TrimSpace(r.PostForm.Get("password"))
	if !emailRegexp.MatchString(email) {
		jRsp(w, http.StatusBadRequest, "邮箱格式非法", nil)
		return
	}
	if !md5Regexp.MatchString(password) {
		jRsp(w, http.StatusBadRequest, "密码不正确", nil)
		return
	}
	user, err := model.Login(email, password, realip.FromRequest(r))
	if err != nil {
		jRsp(w, http.StatusForbidden, err.Error(), nil)
		return
	}
	session.Values["user"] = user
	if _, exist := r.PostForm["remember"]; !exist {
		session.Options.MaxAge = 3600
	}
	if err = session.Save(r, w); err != nil {
		jRsp(w, http.StatusForbidden, err.Error(), nil)
		return
	}
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	tokenMap[user.ID] = session.ID
	jRsp(w, http.StatusOK, "登录成功", r.Form.Get("refer"))
}

// Logout 登出
func Logout(w http.ResponseWriter, r *http.Request) {
	if session, err := store.Get(r, sessName); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
	} else {
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
