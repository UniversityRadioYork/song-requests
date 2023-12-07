package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func HandleMakeRequest(w http.ResponseWriter, r *http.Request) {
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
}

func HandleUserUpload(w http.ResponseWriter, r *http.Request) {
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

}

func HandleAdminUpload(w http.ResponseWriter, r *http.Request) {
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
}

func HandleBonusRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	newRequest, err := strconv.Atoi(r.FormValue("bonus"))
	if err != nil {
		return
	}

	store.lock.Lock()
	store.BonusRequests = append(store.BonusRequests, newRequest)
	store.update()
	http.Redirect(w, r, "/", http.StatusFound)
}
