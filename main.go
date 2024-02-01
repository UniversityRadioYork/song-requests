package main

import (
	"net/http"
	"os"
	"time"

	"runtime/debug"

	"github.com/UniversityRadioYork/myradio-go"
	"github.com/gorilla/sessions"
	"gopkg.in/yaml.v3"
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:7]
			}
		}
	}
	return ""
}()

var MyRadioSession *myradio.Session
var store Datastore = Datastore{
	RequestsPerPerson: 6, // default
}

var nameCache map[int]struct {
	name      string
	cacheTime time.Time
} = make(map[int]struct {
	name      string
	cacheTime time.Time
})

const cacheExpiryDays int = 2

func GetNameOfUser(id int) string {
	if cacheResult, ok := nameCache[id]; ok {
		if cacheResult.cacheTime.After(time.Now().AddDate(0, 0, -cacheExpiryDays)) {
			return cacheResult.name
		}
	}

	name, err := MyRadioSession.GetUserName(id)
	if err != nil {
		// TODO
		panic(err)
	}

	nameCache[id] = struct {
		name      string
		cacheTime time.Time
	}{
		name:      name,
		cacheTime: time.Now(),
	}

	return name
}

func main() {

	cookiestore.Options = &sessions.Options{
		MaxAge: int(time.Minute * 10),
	}

	time.Sleep(5*time.Second)
	f, err := os.ReadFile("data/data.yaml")
	if err != nil {
		defaultYaml, err := yaml.Marshal(store)
		if err != nil {
			panic(err)
		}
		if err = os.WriteFile("data/data.yaml", defaultYaml, 0644); err != nil {
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

	mux.HandleFunc("/", HandleIndex)
	mux.HandleFunc("/iwant", HandleMakeRequest)
	mux.HandleFunc("/ihaveuploaded", HandleUserUpload)
	mux.HandleFunc("/bought", HandleAdminUpload)
	mux.HandleFunc("/reject", HandleReject)
	mux.HandleFunc("/cancel", HandleUserCancel)
	mux.HandleFunc("/bonus", HandleBonusRequest)
	mux.HandleFunc("/auth", auth)
	mux.HandleFunc("/logout", HandleLogout)

	var host string
	if os.Getenv("SONG_REQUESTS_DEV") == "1" {
		host = "local-development.ury.org.uk"
	} else {
		host = "0.0.0.0"
	}

	http.ListenAndServe(host+":8080", &AuthMiddleware{mux})

}
