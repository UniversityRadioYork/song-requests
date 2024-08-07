package main

import (
	"fmt"
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
		Uploaded:  StateNotUploaded,
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
			store.Requests[i].Uploaded = StateUploaded
			store.Requests[i].UploadedBy = user
		}
	}
	store.update()
	http.Redirect(w, r, "/", http.StatusFound)

}

func HandleUserCancel(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	store.lock.Lock()
	for i, v := range store.Requests {
		if v.ID.String() == r.FormValue("id") {
			store.Requests[i].Uploaded = StateCancelled
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
			store.Requests[i].Uploaded = StateUploaded
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

func HandleReject(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	user := r.Context().Value(UserCtxKey).(int)

	store.lock.Lock()
	for i, v := range store.Requests {
		if v.ID.String() == r.FormValue("id") {
			store.Requests[i].Uploaded = StateRejected
			store.Requests[i].UploadedBy = user
		}
	}
	store.update()
	http.Redirect(w, r, "/", http.StatusFound)

}

func HandleStartNewYear(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserCtxKey).(int)
	if !isAdminUser(user) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Forbidden")
		return
	}

	if r.URL.Query().Get("confirm") != "on" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	CreateNewYear()
	http.Redirect(w, r, "/", http.StatusFound)
}

func HandleCSV(w http.ResponseWriter, r *http.Request) {
	req := r.URL.Query().Get("date")

	if req == "" {
		// return 400
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad Request")
		return
	}

	http.ServeFile(w, r, fmt.Sprintf("data/%s.csv", req))

}
