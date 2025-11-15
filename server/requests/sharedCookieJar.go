package requests

import (
	"net/http"
	"net/url"
)

type SharedCookieJar struct {
	CookieSlice []*http.Cookie
}

func (jar *SharedCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.CookieSlice = append(jar.CookieSlice, cookies...)
}

func (jar *SharedCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return jar.CookieSlice
}
