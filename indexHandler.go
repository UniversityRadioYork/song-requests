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

	adminUser := isAdminUser(user)

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
			} else if v.Uploaded == StateRejected || v.Uploaded == StateCancelled {
				refundRequests++
			}
		}
	}

	// Requests not done by admin yet
	unuploadedRequests := make([]Request, 0)
	completedRequests := make([]Request, 0)
	for _, v := range store.Requests {
		if v.Uploaded == StateNotUploaded {
			unuploadedRequests = append(unuploadedRequests, v)
		} else {
			completedRequests = append(completedRequests, v)
		}
	}

	// Admin User Page
	totalCost := 0.00

	type userRemainingRequestsObject struct {
		NumRequests int
		UserID      int
	}
	userRemainingRequests := make(map[string]userRemainingRequestsObject)

	if adminUser {
		totalCost += store.InitialSpending

		var e userRemainingRequestsObject

		for _, v := range store.Requests {
			totalCost += v.Cost

			if _, ok := userRemainingRequests[v.UserName()]; !ok {
				userRemainingRequests[v.UserName()] = struct {
					NumRequests int
					UserID      int
				}{NumRequests: store.RequestsPerPerson, UserID: v.User}
			}

			e = userRemainingRequests[v.UserName()]
			e.NumRequests--
			userRemainingRequests[v.UserName()] = e

			if v.Uploaded == StateUploaded && v.Cost == 0 {
				e = userRemainingRequests[v.UserName()]
				e.NumRequests++
				userRemainingRequests[v.UserName()] = e
			} else if v.Uploaded == StateRejected || v.Uploaded == StateCancelled {
				e = userRemainingRequests[v.UserName()]
				e.NumRequests++
				userRemainingRequests[v.UserName()] = e
			}
		}

		for _, v := range store.BonusRequests {
			if _, ok := userRemainingRequests[GetNameOfUser(v)]; !ok {
				userRemainingRequests[GetNameOfUser(v)] = userRemainingRequestsObject{NumRequests: store.RequestsPerPerson, UserID: v}
			}
			e = userRemainingRequests[GetNameOfUser(v)]
			e.NumRequests++
			userRemainingRequests[GetNameOfUser(v)] = e
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
		AdminUser             bool
		TotalCost             string
		CompletedRequests     []Request
		UserRemainingRequests map[string]userRemainingRequestsObject
		PreviousYearsData     []string
	}{
		LoggedInName: GetNameOfUser(user),
		CommitHash:   Commit,

		SongRequests:       songRequests,
		RequestsLeft:       store.RequestsPerPerson - len(songRequests) + refundRequests + bonusCount,
		UnuploadedRequests: unuploadedRequests,

		AdminUser:             adminUser,
		CompletedRequests:     completedRequests,
		TotalCost:             fmt.Sprintf("%.2f", totalCost),
		UserRemainingRequests: userRemainingRequests,
		PreviousYearsData:     previousYearCSVs,
	}); err != nil {
		// TODO
		panic(err)
	}
}
