# syntax=docker/dockerfile:1
## We specify the base image we need for our
## go application
FROM golang:1.18.6-alpine3.16

# Move to /dist directory as the place for resulting binary folder
#RUN mkdir /dist
#WORKDIR /dist

# create directories in container, can use -p if there are sub directories
#RUN mkdir /{build,config}

# copy all files to dependencies and go files /server
# WORKDIR defines the container directory, copy copies from host to container
# Directory is created if it is not already there
WORKDIR /build
COPY /server/go.mod .
COPY /server/go.sum .
RUN go mod download
COPY /server/*.go .

# build binary in /server
RUN go build -o foodserver .
WORKDIR /config
#copy the env file into container directory /config, copy will create the directory if not already there
# Note on COPY <src> <dst>
#If <dst> not starts with / then it is considered as absolute path. 
#If <dst> starts with / then it is considered as relative path to the work directory. 
#If <dst> ends with / the it is considered as directory. 
#If <dst> does not ends with / the it is considered as file. A file without any extension will be created. 
COPY /config/* .

# Export necessary port
#API
EXPOSE 5000
#mySQL
#EXPOSE 3307/tcp

## Our start command which kicks off
## our newly created binary executable
CMD ["/build/foodserver"]