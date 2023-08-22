FROM golang:1.19-alpine as builder

ENV GO111MODULE=on
ENV PORT 8082
ENV APP reading

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY ${APP}/ ${APP}
COPY tracing/ tracing

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/${APP} ./${APP}

# final stage
FROM scratch
COPY --from=builder /out/${APP} /app/

EXPOSE ${PORT}
ENTRYPOINT ["/app/reading"]
