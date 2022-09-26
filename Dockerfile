# syntax=docker/dockerfile:1
## We specify the base image we need for our
## go application
FROM golang:1.18.6-alpine3.16

# Move to /dist directory as the place for resulting binary folder
#RUN mkdir /dist
#WORKDIR /dist

# create sub directory
RUN mkdir /build
RUN mkdir /config

# copy all files to dependencies and go files /server
WORKDIR /build
COPY /server/go.mod .
COPY /server/go.sum .
RUN go mod download
COPY /server/*.go .

# build binary in /server
RUN go build -o foodserver .

#copy the env file into image
WORKDIR /config
COPY /config/* .

# Export necessary port
#API
EXPOSE 5000
#mySQL
#EXPOSE 3307/tcp

## Our start command which kicks off
## our newly created binary executable
CMD ["/build/foodserver"]