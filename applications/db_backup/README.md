# Database Backup CronJob

This application is a K8s CronJob meant to be deployed to a kubernetes cluster that runs MySQL and PostgreSQL on it. It supports backing up MySQL and PostgreSQL databases individually into a GCS bucket.

## Setup Instructions

### 1. Setting up Google Cloud

1. Create a bucket for these backups;
2. Create a service account in Google Cloud Console;
3. Grant it the following permissions on your backup bucket:
   - Storage Object Creator;
   - Storage Object Viewer;
4. Generate and download the JSON key file;

### 2. Create Kubernetes Secrets

```bash
kubectl create secret generic gcs-service-account \
  --from-file=key.json=/path/to/your/service-account.json \
  --namespace=postgresql-database-namespace

kubectl create secret generic gcs-service-account \
  --from-file=key.json=/path/to/your/service-account.json \
  --namespace=mysql-database-namespace
```

### 3. Configure Backup Job

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgresql-backup
  namespace: postgresql-database-namespace
spec:
  schedule: "0 2 * * *"
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      backoffLimit: 2
      template:
        metadata:
          labels:
            app: postgresql-backup
        spec:
          restartPolicy: Never
          containers:
          - name: backup
            image: ghcr.io/duck-labs/upduck-db-backup:latest
            imagePullPolicy: Always
            args:
            - "--db-type=postgresql"
            - "--host=pgsql-service.postgresql-database-namespace.svc.cluster.local"
            - "--port=5432"
            - "--username=$(DB_USERNAME)"
            - "--password=$(DB_PASSWORD)"
            - "--databases=$(DATABASES)"
            - "--gcs-bucket=$(GCS_BUCKET)"
            envFrom:
            - secretRef:
                name: database-backup-secrets
            - name: GOOGLE_APPLICATION_CREDENTIALS
              value: "/etc/gcp/key.json"
            volumeMounts:
            - name: gcp-service-account
              mountPath: /etc/gcp
              readOnly: true
          volumes:
          - name: gcp-service-account
            secret:
              secretName: gcs-service-account
```
> ./postgresql-backup.yaml

### 4. Deploy the CronJobs

```bash
kubectl apply -f postgresql-backup.yaml
```

### Manual Backup Execution

```bash
kubectl create job postgresql-manual-backup --from=cronjob/postgresql-backup -n database-postgresql
```

## Backup Retention

Consider implementing a retention policy in your GCS bucket to automatically delete old backups and control storage costs. You can configure lifecycle rules in the Google Cloud Console or using gsutil.

Example lifecycle rule to delete backups older than 30 days:
```json
{
  "rule": [
    {
      "action": {"type": "Delete"},
      "condition": {
        "age": 7,
        "matchesPrefix": ["database-backups/"]
      }
    }
  ]
}
```
