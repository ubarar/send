FROM golang:latest AS builder

COPY . /app

RUN cd /app && go mod tidy && go build -o send .

# multi-stage build
FROM ubuntu:latest

RUN mkdir /app

COPY --from=builder /app/send /app
COPY --from=builder /app/index.html /app
COPY --from=builder /app/view.html /app
COPY --from=builder /app/assets/style.css /app/assets/style.css


EXPOSE 8080

WORKDIR /app

CMD /app/send -storage=/storage
