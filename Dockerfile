FROM golang:1.16 as build

WORKDIR /src
COPY go.* ./
RUN go mod download

COPY . .
RUN go build -o /go/bin/main 

FROM gcr.io/distroless/base:nonroot 
COPY --from=build /go/bin/main /app/
WORKDIR /app
ENTRYPOINT ["/app/main"]