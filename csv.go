package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func CreateNewYear() {
	todayDateString := time.Now().Format(time.DateOnly)

	// Create a CSV from current data
	csvFile, err := os.Create(fmt.Sprintf("data/%s.csv", todayDateString))
	if err != nil {
		// TODO
	}
	defer csvFile.Close()

	wr := csv.NewWriter(csvFile)
	wr.Write([]string{
		"Title",
		"Artist",
		"Requester",
		"Request Date",
		"Other Info",
		"Cost",
		"State",
		"Uploader",
	})

	for _, request := range store.Requests {
		wr.Write([]string{
			request.Title,
			request.Artist,
			request.UserName(),
			request.Datetime.Format(time.DateOnly),
			request.OtherInfo,
			fmt.Sprintf("%v", request.Cost),
			string(request.Uploaded),
			request.UploadedByName(),
		})
	}

	wr.Flush()
	previousYearCSVs = append(previousYearCSVs, todayDateString)

	// Reset all the data
	currentDataFile, err := os.Open("data/data.yaml")
	if err != nil {
		// TODO
	}
	defer currentDataFile.Close()

	backupDataFile, err := os.Create(fmt.Sprintf("data/backup-%s.yaml", todayDateString))
	if err != nil {
		// TODO
	}
	defer backupDataFile.Close()

	_, err = io.Copy(backupDataFile, currentDataFile)
	if err != nil {
		// TODO
	}

	err = os.Remove("data/data.yaml")
	if err != nil {
		// TODO
	}

	store.lock.Lock()
	store = Datastore{
		RequestsPerPerson: DefaultRequestsPerPerson,
	}
	makeNewDatafile()
}

func populatePreviousYearCSVs() {
	if err := filepath.Walk("data/", func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name()[len(info.Name())-4:] == ".csv" {
			previousYearCSVs = append(previousYearCSVs, info.Name()[:len(info.Name())-4])
		}

		return nil
	}); err != nil {
		// TODO
		panic(err)
	}
}
