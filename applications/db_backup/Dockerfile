FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    postgresql-client \
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
