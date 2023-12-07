package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	user := r.Context().Value(UserCtxKey).(int)

	// Is the User an Admin User?
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

	// User's Details
	bonusCount := 0
	for _, v := range store.BonusRequests {
		if v == user {
			bonusCount++
		}
	}

	songRequests := make([]Request, 0)
	refundRequests := 0
	for _, v := range store.Requests {
		if v.User == user {
			songRequests = append(songRequests, v)
			if v.Uploaded == StateUploaded && v.Cost == 0 {
				refundRequests += 1
			} else if v.Uploaded == StateRejected {
				refundRequests++
			}
		}
	}

	// Requests not done by admin yet
	unuploadedRequests := make([]Request, 0)
	for _, v := range store.Requests {
		if v.Uploaded == StateNotUploaded {
			unuploadedRequests = append(unuploadedRequests, v)
		}
	}

	// Admin User Page
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
		// General
		LoggedInName string
		CommitHash   string

		// Normal User Page
		SongRequests       []Request
		RequestsLeft       int
		UnuploadedRequests []Request

		// Admin User Page
		AdminUser   bool
		TotalCost   string
		AllRequests []Request
	}{
		LoggedInName: GetNameOfUser(user),
		CommitHash:   Commit,

		SongRequests:       songRequests,
		RequestsLeft:       store.RequestsPerPerson - len(songRequests) + refundRequests + bonusCount,
		UnuploadedRequests: unuploadedRequests,

		AdminUser:   adminUser,
		AllRequests: allRequests,
		TotalCost:   fmt.Sprintf("%.2f", totalCost),
	}); err != nil {
		// TODO
		panic(err)
	}
}
