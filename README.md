# Song Requests
This is a simple Go app that handles a calendar front-end for off-air studio booking at URY. 

# Getting started

## Without Docker
Make sure Go is installed.
```bash
mkdir data
go build
./song-request
```
It is then running on `localhost:8080`

## With Docker
```bash
docker build -t songrec .
docker run -d -p 8080:8080 songrec
```
This will error if you have already made a data folder so make sure to delete this first if it is there.
