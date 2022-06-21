FROM golang:alpine 
RUN apk add git

# RUN mkdir /app
# ADD . /app/
# WORKDIR /app
# RUN go mod download
# RUN go build -o main ./...
# CMD ["/app/main"]

RUN mkdir /tarabish-server
ADD . /tarabish-server/
WORKDIR /tarabish-server
RUN go mod download
RUN go build -o tarabish-server ./...
CMD ["/tarabish-server/server"]