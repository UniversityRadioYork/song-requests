package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/UniversityRadioYork/myradio-go"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const AuthRealm string = "ury-song-requests"

var MyRadioSession *myradio.Session

func GetNameOfUser(id int) string {
	name, err := MyRadioSession.GetUserName(id)
	if err != nil {
		// TODO
		panic(err)
	}

	return name
}

type Request struct {
	Datetime   time.Time
	ID         uuid.UUID
	User       int
	Title      string
	Artist     string
	OtherInfo  string
	UploadedBy int
	Uploaded   bool
	Cost       float64
}

func (r Request) TimeStr() string {
	return r.Datetime.Format(time.RFC1123)
}

func (r Request) UserName() string {
	return GetNameOfUser(r.User)
}

func (r Request) UploadedByName() string {
	return GetNameOfUser(r.UploadedBy)
}

func (r Request) FormatCost() string {
	return fmt.Sprintf("%.2f", r.Cost)
}

type Datastore struct {
	lock              sync.RWMutex
	InitialSpending   float64
	RequestsPerPerson int
	Requests          []Request
	BonusRequests     []int
}

type AuthMiddleware struct {
	handler http.Handler
}

type ContextKey string

const UserCtxKey ContextKey = "user"

func (a *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()

	if username == "logout" {
		http.Error(w, "Please close this site now to fully log out.", http.StatusUnauthorized)
		return
	}

	var user *myradio.User

	if ok {
		var err error
		user, err = MyRadioSession.UserCredentialsTest(username, password)
		if err != nil {
			ok = false
		}

		if user == nil {
			ok = false
		}
	}

	if ok {
		ctx := context.WithValue(context.Background(), UserCtxKey, user.MemberID)
		a.handler.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s", charset="UTF-8"`, AuthRealm))
	http.Error(w, "Unauthorized", http.StatusUnauthorized)

}

func (s *Datastore) update() {
	defer s.lock.Unlock()

	dataFile, err := os.OpenFile("data.yaml", os.O_WRONLY, os.ModeAppend)
	if err != nil {
		// TODO
		panic(err)
	}
	defer dataFile.Close()

	d, err := yaml.Marshal(s)
	if err != nil {
		// TODO
		panic(err)
	}

	_, err = dataFile.Write(d)
	if err != nil {
		// TODO
		panic(err)
	}
}

func main() {

	store := Datastore{
		RequestsPerPerson: 6, // default
	}

	f, err := os.ReadFile("data.yaml")
	if err != nil {
		defaultYaml, err := yaml.Marshal(store)
		if err != nil {
			panic(err)
		}
		if err = os.WriteFile("data.yaml", defaultYaml, 0644); err != nil {
			panic(err)
		}

	}

	err = yaml.Unmarshal(f, &store)
	if err != nil {
		panic(err)
	}

	MyRadioSession, err = myradio.NewSessionFromKeyFile()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		store.lock.RLock()
		defer store.lock.RUnlock()

		user := r.Context().Value(UserCtxKey).(int)
		adminUser := false
		management, err := MyRadioSession.GetTeamWithOfficers("management")
		if err != nil {
			panic(err)
		}

		for _, v := range management.Officers {
			if v.User.MemberID == user {
				adminUser = true
			}
		}

		bonusCount := 0
		for _, v := range store.BonusRequests {
			if v == user {
				bonusCount++
			}
		}

		songRequests := make([]Request, 0)
		requestsUploadedNotBought := 0
		for _, v := range store.Requests {
			if v.User == user {
				songRequests = append(songRequests, v)
				if v.Uploaded && v.Cost == 0 {
					requestsUploadedNotBought += 1
				}
			}
		}

		unuploadedRequests := make([]Request, 0)
		for _, v := range store.Requests {
			if !v.Uploaded {
				unuploadedRequests = append(unuploadedRequests, v)
			}
		}

		// admin user stuffs
		totalCost := 0.00
		allRequests := []Request{}
		if adminUser {
			allRequests = store.Requests
			totalCost += store.InitialSpending
			for _, v := range store.Requests {
				totalCost += v.Cost
			}
		}

		t := template.Must(template.New("index.html").ParseFiles("index.html"))
		if err := t.Execute(w, struct {
			LoggedInName       string
			SongRequests       []Request
			RequestsLeft       int
			UnuploadedRequests []Request
			AdminUser          bool
			TotalCost          string
			AllRequests        []Request
		}{
			LoggedInName:       GetNameOfUser(user),
			SongRequests:       songRequests,
			RequestsLeft:       store.RequestsPerPerson - len(songRequests) + requestsUploadedNotBought + bonusCount,
			UnuploadedRequests: unuploadedRequests,
			AdminUser:          adminUser,
			AllRequests:        allRequests,
			TotalCost:          fmt.Sprintf("%.2f", totalCost),
		}); err != nil {
			// TODO
			panic(err)
		}

	})

	mux.HandleFunc("/iwant", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		user := r.Context().Value(UserCtxKey).(int)

		store.lock.Lock()
		store.Requests = append(store.Requests, Request{
			Datetime:  time.Now(),
			ID:        uuid.New(),
			User:      user,
			Title:     r.FormValue("song-title"),
			Artist:    r.FormValue("artist"),
			OtherInfo: r.FormValue("other-info"),
		})

		store.update()

		http.Redirect(w, r, "/", http.StatusFound)
	})

	mux.HandleFunc("/ihaveuploaded", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		user := r.Context().Value(UserCtxKey).(int)

		store.lock.Lock()
		for i, v := range store.Requests {
			if v.ID.String() == r.FormValue("id") {
				store.Requests[i].Uploaded = true
				store.Requests[i].UploadedBy = user
			}
		}
		store.update()
		http.Redirect(w, r, "/", http.StatusFound)

	})

	mux.HandleFunc("/bought", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		user := r.Context().Value(UserCtxKey).(int)
		cost, _ := strconv.ParseFloat(r.FormValue("cost"), 64)

		store.lock.Lock()
		for i, v := range store.Requests {
			if v.ID.String() == r.FormValue("id") {
				store.Requests[i].Uploaded = true
				store.Requests[i].UploadedBy = user
				store.Requests[i].Cost = cost
			}
		}
		store.update()
		http.Redirect(w, r, "/", http.StatusFound)

	})

	mux.HandleFunc("/bonus", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		newRequest, err := strconv.Atoi(r.FormValue("bonus"))
		if err != nil {
			return
		}

		store.lock.Lock()
		store.BonusRequests = append(store.BonusRequests, newRequest)
		store.update()
		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.ListenAndServe(":8080", &AuthMiddleware{mux})

}
