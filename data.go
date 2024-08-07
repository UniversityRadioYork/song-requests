package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type UploadState string

const (
	StateUploaded    UploadState = "UPLOADED"
	StateNotUploaded UploadState = "NOTUPLOADED"
	StateCancelled   UploadState = "CANCELLED"
	StateRejected    UploadState = "REJECTED"
)

type Request struct {
	Datetime   time.Time
	ID         uuid.UUID
	User       int
	Title      string
	Artist     string
	OtherInfo  string
	UploadedBy int
	Uploaded   UploadState
	Cost       float64
}

func (r Request) TimeStr() string {
	return r.Datetime.Format("02/01/2006")
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

func (s *Datastore) update() {
	defer s.lock.Unlock()

	dataFile, err := os.OpenFile("data/data.yaml", os.O_WRONLY, os.ModeAppend)
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

func isAdminUser(userId int) bool {
	// Is the User an Admin User?
	management, err := MyRadioSession.GetTeamWithOfficers("management")
	if err != nil {
		panic(err)
	}

	for _, v := range management.Officers {
		if v.User.MemberID == userId {
			return true
		}
	}

	return false
}
