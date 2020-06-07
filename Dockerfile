FROM golang:1.13-stretch AS builder

# Building project
WORKDIR /build

COPY . .
RUN go build -v ./server.go

FROM ubuntu:18.04

# Expose server & database ports
EXPOSE 5000
EXPOSE 5432

RUN apt-get -y update && apt-get -y install postgresql
USER postgres
ENV PGPASSWORD="docker"

COPY /database/create.sql .
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql -U docker -d docker -f ./create.sql &&\
    /etc/init.d/postgresql stop

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/10/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/10/main/postgresql.conf

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

# Copying built binary
COPY --from=builder /build/server .
CMD service postgresql start && ./server