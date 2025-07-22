package authentication

import (
	"net/http"

	"github.com/gorilla/securecookie"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

// SetSession creates a secure session cookie for the authenticated user.
func SetSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/admin/",
		}
		http.SetCookie(response, cookie)
	}
}

// GetUserName extracts the username from the session cookie in the request.
func GetUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

// ClearSession removes the session cookie by setting it to expire.
func ClearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/admin/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}
