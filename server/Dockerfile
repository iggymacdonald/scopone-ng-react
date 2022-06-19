FROM golang:alpine 
RUN apk add git

# RUN mkdir /app
# ADD . /app/
# WORKDIR /app
# RUN go mod download
# RUN go build -o main ./...
# CMD ["/app/main"]

RUN mkdir /scopone-server
ADD . /scopone-server/
WORKDIR /scopone-server
RUN go mod download
RUN go build -o scopone-server ./...
CMD ["/scopone-server/server"]