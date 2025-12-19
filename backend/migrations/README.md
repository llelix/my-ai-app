# Database Migrations

This directory contains database migration scripts for the AI Knowledge Application.

## Migration Files

### 001_add_minio_support.sql
- **Purpose**: Add MinIO support to database models
- **Changes**:
  - Adds `upload_id` column to `upload_sessions` table for storing MinIO multipart upload IDs
  - Creates index on `upload_id` for better query performance
  - Adds documentation comment to `documents.file_path` column

### 001_add_minio_support_sqlite.sql
- **Purpose**: SQLite-compatible version of the MinIO support migration
- **Changes**: Same as PostgreSQL version but adapted for SQLite syntax

## How Migrations Work

1. Migrations are automatically run during application startup
2. The system detects the database type (PostgreSQL or SQLite) and runs the appropriate migration files
3. Migration failures are logged as warnings but don't stop the application (since GORM AutoMigrate handles most schema changes)
4. Manual migrations are run before GORM AutoMigrate to handle complex schema changes

## Adding New Migrations

1. Create a new SQL file with a descriptive name and incremental number
2. For PostgreSQL compatibility, create a `.sql` file
3. For SQLite compatibility, create a `_sqlite.sql` file
4. Test both versions thoroughly
5. Document the changes in this README

## Migration Safety

- Migrations should be idempotent (safe to run multiple times)
- Use `IF NOT EXISTS` or similar constructs where possible
- Always backup your database before running migrations in production
- Test migrations on a copy of production data first