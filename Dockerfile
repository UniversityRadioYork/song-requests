FROM golang:1.18

WORKDIR /usr/src/app

COPY . .
RUN go build
RUN mkdir data

EXPOSE 8080

CMD [ "./song-requests" ]