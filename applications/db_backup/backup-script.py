#!/usr/bin/env python3
"""
Database Backup Script for PostgreSQL and MySQL
Supports backing up individual databases and uploading to Google Cloud Storage
"""

import os
import sys
import argparse
import subprocess
import logging
import datetime
from pathlib import Path
from google.cloud import storage
import tempfile
import gzip

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class DatabaseBackup:
    def __init__(self, db_type, host, port, username, password, gcs_bucket):
        self.db_type = db_type.lower()
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.gcs_bucket = gcs_bucket
        self.now = datetime.datetime.now()
        self.timestamp = self.now.strftime('%Y%m%d_%H%M%S')
        
        self.storage_client = storage.Client()
        self.bucket = self.storage_client.bucket(gcs_bucket)
        
    def create_postgresql_dump(self, database_name, output_file):
        try:
            env = os.environ.copy()
            env['PGPASSWORD'] = self.password
            
            cmd = [
                'pg_dump',
                '-h', self.host,
                '-p', str(self.port),
                '-U', self.username,
                '-d', database_name,
                '--verbose',
                '--no-password',
                '--format=custom',
                '--compress=9'
            ]
            
            logger.info(f"Creating PostgreSQL dump for database: {database_name}")
            with open(output_file, 'wb') as f:
                subprocess.run(cmd, stdout=f, stderr=subprocess.PIPE, env=env, check=True)
            
            logger.info(f"PostgreSQL dump created successfully: {output_file}")
            return True
            
        except subprocess.CalledProcessError as e:
            logger.error(f"PostgreSQL dump failed: {e.stderr.decode()}")
            return False
        except Exception as e:
            logger.error(f"Unexpected error during PostgreSQL dump: {str(e)}")
            return False
    
    def create_mysql_dump(self, database_name, output_file):
        try:
            cmd = [
                'mysqldump',
                f'--host={self.host}',
                f'--port={self.port}',
                f'--user={self.username}',
                f'--password={self.password}',
                '--single-transaction',
                '--routines',
                '--triggers',
                '--events',
                '--hex-blob',
                '--opt',
                database_name
            ]
            
            logger.info(f"Creating MySQL dump for database: {database_name}")
            with open(output_file, 'w') as f:
                subprocess.run(cmd, stdout=f, stderr=subprocess.PIPE, check=True)
            
            logger.info(f"MySQL dump successfully created: {output_file}")
            return True
            
        except subprocess.CalledProcessError as e:
            logger.error(f"MySQL dump failed: {e.stderr.decode()}")
            return False
        except Exception as e:
            logger.error(f"Unexpected error during MySQL dump: {str(e)}")
            return False
    
    def upload_to_gcs(self, local_file, remote_path):
        try:
            blob = self.bucket.blob(remote_path)
            
            logger.info(f"Uploading {local_file} to gs://{self.gcs_bucket}/{remote_path}")
            blob.upload_from_filename(local_file)
            
            blob.metadata = {
                'backup_timestamp': self.timestamp,
            }
            blob.patch()
            
            logger.info(f"Upload completed successfully")
            return True
            
        except Exception as e:
            logger.error(f"Upload to GCS failed: {str(e)}")
            return False
    
    def backup_database(self, database_name):
        logger.info(f"Starting backup for {self.db_type} database: {database_name}")
        
        with tempfile.NamedTemporaryFile(delete=False, suffix=f'_{database_name}_{self.timestamp}.sql') as temp_file:
            temp_path = temp_file.name
        
        try:
            if self.db_type == 'postgresql':
                success = self.create_postgresql_dump(database_name, temp_path)
            elif self.db_type == 'mysql':
                success = self.create_mysql_dump(database_name, temp_path)
            else:
                logger.error(f"Unsupported database type: {self.db_type}")
                return False
            
            if not success:
                return False
            
            remote_path = f"database-backups/{self.host}/{self.now.year}/{self.now.month}/{database_name}_{self.timestamp}.sql"
            
            upload_success = self.upload_to_gcs(temp_path, remote_path)
            
            if upload_success:
                file_size = os.path.getsize(temp_path)
                logger.info(f"Backup completed for {database_name}. File size: {file_size:,} bytes")
                return True
            else:
                return False
                
        finally:
            if os.path.exists(temp_path):
                os.remove(temp_path)


def main():
    parser = argparse.ArgumentParser(description='Database Backup Tool')
    parser.add_argument('--db-type', required=True, choices=['postgresql', 'mysql'], 
                       help='Database type')
    parser.add_argument('--host', required=True, help='Database host')
    parser.add_argument('--port', type=int, required=True, help='Database port')
    parser.add_argument('--username', required=True, help='Database username')
    parser.add_argument('--password', required=True, help='Database password')
    parser.add_argument('--databases', required=True, help='Comma-separated list of databases to backup')
    parser.add_argument('--gcs-bucket', required=True, help='Google Cloud Storage bucket name')
    
    args = parser.parse_args()
    
    backup_client = DatabaseBackup(
        db_type=args.db_type,
        host=args.host,
        port=args.port,
        username=args.username,
        password=args.password,
        gcs_bucket=args.gcs_bucket
    )
    
    databases_to_backup = [db.strip() for db in args.databases.split(',')]
    
    logger.info(f"Databases to backup: {databases_to_backup}")
    
    failed_backups = []
    successful_backups = []
    
    for database in databases_to_backup:
        if backup_client.backup_database(database):
            successful_backups.append(database)
        else:
            failed_backups.append(database)
    
    logger.info(f"Backup Summary:")
    logger.info(f"  Successful: {len(successful_backups)} - {successful_backups}")
    logger.info(f"  Failed: {len(failed_backups)} - {failed_backups}")
    
    return 0 if not failed_backups else 1

if __name__ == '__main__':
    sys.exit(main())
