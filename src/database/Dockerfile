FROM postgres:15
ENV POSTGRES_PASSWORD=pass
ENV POSTGRES_DB=bactic
ENV POSTGRES_USER=postgres
COPY migrations/001.sql /docker-entrypoint-initdb.d/001.sql