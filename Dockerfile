FROM golang:1.9

RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

RUN apt-get update && apt-get install -yq dnsutils

EXPOSE 9991 8881 80

WORKDIR /go/src/github.com/lucasroesler/builderpoc
COPY . /go/src/github.com/lucasroesler/builderpoc

RUN go-wrapper install ./...

CMD ["builder", "-port=9090", "-host=0.0.0.0", "-registry=''"]

