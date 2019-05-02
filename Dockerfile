FROM golang:1.12-alpine3.9

RUN apk add git

ADD . /rate/.

WORKDIR /rate

RUN go install -mod=vendor ./...

ENTRYPOINT [ "rate" ]

CMD [ "-rpm=100" ]
