FROM python:3.11-slim

RUN apt-get update && apt-get install -y wget gnupg2 lsb-release

RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list
    
RUN apt-get update

RUN apt-get install -y \
    postgresql-client-16 \
    default-mysql-client \
    curl

RUN pip install --no-cache-dir \
    google-cloud-storage==2.10.0 \
    psycopg2-binary==2.9.7 \
    PyMySQL==1.1.0

WORKDIR /app

COPY backup-script.py /app/backup-script.py

RUN chmod +x /app/backup-script.py

RUN useradd -m -u 1000 backupuser && chown -R backupuser:backupuser /app
USER backupuser

ENTRYPOINT ["python3", "/app/backup-script.py"]
