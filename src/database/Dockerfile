FROM postgres:15
ENV POSTGRES_PASSWORD=password
ENV POSTGRES_DB=database
ENV POSTGRES_USER=username
COPY migrations/001.sql /docker-entrypoint-initdb.d/001.sql