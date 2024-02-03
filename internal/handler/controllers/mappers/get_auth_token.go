package mappers

import "net/http"

func GetAuthToken(req *http.Request) string {
	// check if token exists in cookie
	cookie, err := req.Cookie("_token")
	if err != nil {
		return ""
	}

	return cookie.Value
}
