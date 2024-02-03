package controllers

import (
	"net/http"
	"time"
)

func LogoutHandler(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    "_token",
		Expires: time.Unix(0, 0),
		Value:   "",
		Path:    "/",
	})

	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/login", http.StatusFound)
}
