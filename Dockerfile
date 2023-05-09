FROM golang:1.20.4-alpine3.16

WORKDIR /app
COPY . .
WORKDIR /app/cmd
RUN go build
EXPOSE 6001
CMD [ "./cmd" ]

