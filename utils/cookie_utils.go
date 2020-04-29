package utils

import (
	"net/http"
	"time"

	"github.com/spf13/cast"
)

// SetCookie function write a new cookie to the response http
func SetCookie(w http.ResponseWriter, name, value interface{}, expire time.Time, maxAge int, domain, path string, httpOnly bool) {
	strName := cast.ToString(name)
	strValue := cast.ToString(value)

	cookie := &http.Cookie{
		Name:     strName,
		Value:    strValue,
		Expires:  expire,
		HttpOnly: httpOnly,
		MaxAge:   maxAge,
		Domain:   domain,
		Path:     path,
	}
	http.SetCookie(w, cookie)
}

// SetCookies function write a cookie's map to the response http
func SetCookies(w http.ResponseWriter, cookies map[string]interface{}, expire time.Time, maxAge int, domain, path string, httpOnly bool) {
	for name, value := range cookies {
		SetCookie(w, name, value, expire, maxAge, domain, path, httpOnly)
	}
}

// GetCookie will get value of cookie send from request
func GetCookie(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
