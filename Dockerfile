# The base go-image
FROM golang:1.17-alpine

RUN CGO_ENABLED=0

# Create a directory for the app
RUN mkdir /app
RUN mkdir /app/data

# Copy all files from the current directory to the app directory
COPY . /app

# Set working directory
WORKDIR /app

ENV TBS_DB_DSN="dbo:Znmml2mkNuXaHDp@tcp(db:3306)/tinybuildserver?charset=utf8&parseTime=true"

RUN go get -u ./...

# Run command as described:
# go build will build an executable file named server in the current directory
RUN go build -ldflags "-s -w -X 'main.Version=Docker'" -o server cmd/tiny-build-server/main.go

EXPOSE 8271

# Run the server executable
CMD [ "/app/server" ]