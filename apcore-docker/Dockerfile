FROM postgres:15

ENV POSTGRES_USER=apcore
ENV POSTGRES_PASSWORD=secret
ENV POSTGRES_DB=apcoredb

COPY scripts/init-db.sh /docker-entrypoint-initdb.d/
COPY config/config.ini.template /app/config.ini.template
COPY . /app

WORKDIR /app

RUN chmod +x scripts/entrypoint.sh

ENTRYPOINT ["./scripts/entrypoint.sh"]