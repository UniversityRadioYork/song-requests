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

func GetNameOfUser(id int) string {
	name, err := MyRadioSession.GetUserName(id)
	if err != nil {
		// TODO
		panic(err)
	}

	return name
}

func main() {

	cookiestore.Options = &sessions.Options{
		MaxAge: int(time.Minute * 10),
	}

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
	mux.HandleFunc("/bonus", HandleBonusRequest)
	mux.HandleFunc("/auth", auth)
	mux.HandleFunc("/logout", HandleLogout)

	http.ListenAndServe("local-development.ury.org.uk:8080", &AuthMiddleware{mux})

}
