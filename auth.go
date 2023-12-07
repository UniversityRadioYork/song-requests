package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key         = []byte("super-secret-key")
	cookiestore = sessions.NewCookieStore(key)
)

const AuthRealm string = "ury-song-requests"

type AuthMiddleware struct {
	handler http.Handler
}

type ContextKey string

const UserCtxKey ContextKey = "user"

func auth(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.ServeFile(w, r, "login.html")
		return
	} else if r.Method == "POST" {
		session, _ := cookiestore.Get(r, AuthRealm)

		memberid := r.FormValue("memberid")
		if memberid == "" {
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		var err error
		session.Values["memberid"], err = strconv.Atoi(memberid)
		if err != nil {
			panic(err)
		}
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)

	}
	fmt.Fprint(w, "hmmm")
}

func (a *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/auth" {
		a.handler.ServeHTTP(w, r)
		return
	}

	session, _ := cookiestore.Get(r, AuthRealm)
	if auth, ok := session.Values["memberid"].(int); !ok || auth == 0 {
		// redirect to auth
		http.Redirect(w, r, "/auth", http.StatusFound)
	} else {
		ctx := context.WithValue(context.Background(), UserCtxKey, auth)
		a.handler.ServeHTTP(w, r.WithContext(ctx))
	}

}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := cookiestore.Get(r, AuthRealm)
	session.Values["memberid"] = 0
	session.Save(r, w)
	http.Redirect(w, r, "https://ury.org.uk/myradio/MyRadio/logout", http.StatusFound)
}
