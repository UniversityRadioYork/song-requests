# Song Requests
This is a simple Go app that handles member requests for the station to purchase songs.

# Getting started

## Without Docker
Make sure Go is installed.
```bash
mkdir data
go build
HOST=http://localhost:8080 MYRADIO_SIGNING_KEY=abcde ./song-request
```
It is then running on `localhost:8080`

## With Docker
```bash
docker build -t songrec .
docker run -d -p 8080:8080 -e HOST=http://localhost:8080 -e MYRADIO_SIGNING_KEY=abcde songrec
```
This will error if you have already made a data folder so make sure to delete this first if it is there.
