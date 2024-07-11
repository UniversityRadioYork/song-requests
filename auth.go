package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
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

	jwtString := r.URL.Query().Get("jwt")

	if jwtString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "no jwt")
		return
	}

	// parse the token
	token, err := jwt.Parse(jwtString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(os.Getenv("MYRADIO_SIGNING_KEY")), nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	session, _ := cookiestore.Get(r, AuthRealm)

	memberid := claims["uid"].(float64)
	name := claims["name"].(string)

	nameCache[int(memberid)] = struct {
		name      string
		cacheTime time.Time
	}{
		name:      name,
		cacheTime: time.Now(),
	}

	session.Values["memberid"] = int(memberid)
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)

}

func (a *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/auth" {
		a.handler.ServeHTTP(w, r)
		return
	}

	session, _ := cookiestore.Get(r, AuthRealm)
	if auth, ok := session.Values["memberid"].(int); !ok || auth == 0 {
		// redirect to auth
		http.Redirect(w, r, fmt.Sprintf("https://ury.org.uk/myradio/MyRadio/jwt?redirectto=%s/auth", os.Getenv("HOST")), http.StatusFound)
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
