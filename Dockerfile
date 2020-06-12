FROM golang:1.14 AS builder

WORKDIR /build

COPY . .
RUN go build -v ./server.go

FROM ubuntu:20.04

EXPOSE 5000
EXPOSE 5432

ENV DEBIAN_FRONTEND noninteractive
ENV PGPASSWORD="docker"
ENV PGUSERNAME="docker"
ENV PGVER=12

RUN apt-get -y update && apt-get install -y --no-install-recommends apt-utils postgresql-$PGVER;

USER postgres

COPY ./database/create.sql .
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    /etc/init.d/postgresql stop

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

COPY --from=builder ./build .
COPY ./database/create.sql /assets/db/postgres/base.sql
CMD service postgresql start && psql -h localhost -U docker -d docker -p 5432 -a  -f  ./create.sql && ./server
