package service

import "net/http"

type CookieService interface {
	CreateSetAuthCookie(refreshToken string) *http.Cookie
	CreateClearAuthCookie() *http.Cookie
}

type cookieService struct {
	refreshPath string
	cookieName  string
	refreshTTL  int
	mode        http.SameSite
	secureMode  bool
}

func (cs *cookieService) CreateSetAuthCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     cs.cookieName,
		Value:    refreshToken,
		Path:     cs.refreshPath,
		MaxAge:   cs.refreshTTL,
		HttpOnly: true,
		Secure:   cs.secureMode,
		SameSite: cs.mode,
	}
}

func (cs *cookieService) CreateClearAuthCookie() *http.Cookie {
	return &http.Cookie{
		Name:     cs.cookieName,
		Value:    "",
		Path:     cs.refreshPath,
		MaxAge:   0,
		HttpOnly: true,
		Secure:   cs.secureMode,
		SameSite: cs.mode,
	}
}

func NewCookieService(refreshPath, cookieName string, refreshTTL int, sameSiteMode http.SameSite, secureMode bool) CookieService {
	return &cookieService{
		refreshPath: refreshPath,
		cookieName:  cookieName,
		refreshTTL:  refreshTTL,
		mode:        sameSiteMode,
		secureMode:  secureMode,
	}
}
