FROM golang:alpine

# Set necessary environmet variables needed for our image
ENV GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
COPY api .
COPY cmd .
COPY internal .
COPY pkg .

RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o main cmd/main.go

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/main .
RUN rm -rf /build/*

# Command to run when starting the container
CMD ["/dist/main"]
