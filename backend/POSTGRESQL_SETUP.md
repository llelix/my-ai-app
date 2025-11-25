# PostgreSQL Setup Instructions

## Prerequisites
- Docker installed on your system
- Internet connection to pull PostgreSQL image

## Option 1: Using Docker Compose (Recommended)

1. Make sure you have docker-compose installed:
   ```bash
   docker-compose --version
   ```

2. Start PostgreSQL container:
   ```bash
   docker-compose up -d
   ```

3. Check if the container is running:
   ```bash
   docker-compose ps
   ```

## Option 2: Using Docker Commands

1. Run the provided script:
   ```bash
   ./start-postgres.sh
   ```

2. Or run the command manually:
   ```bash
   docker run -d \
     --name ai-knowledge-postgres \
     -e POSTGRES_DB=ai_knowledge_db \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=password \
     -p 5432:5432 \
     postgres:15-alpine
   ```

## Option 3: Using Existing PostgreSQL Instance

If you already have a PostgreSQL server running:

1. Create a database:
   ```sql
   CREATE DATABASE ai_knowledge_db;
   ```

2. Create a user (optional):
   ```sql
   CREATE USER ai_user WITH PASSWORD 'password';
   GRANT ALL PRIVILEGES ON DATABASE ai_knowledge_db TO ai_user;
   ```

## Environment Configuration

The `.env` file has already been configured to use PostgreSQL:

```env
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ai_knowledge_db
```

## Testing Connection

After starting PostgreSQL, test the connection by running the application:

```bash
go run cmd/server/main.go
```

The application will automatically:
1. Connect to PostgreSQL
2. Create tables (auto-migration)
3. Insert seed data if the database is empty

## Troubleshooting

### Network Issues
If you encounter network issues with Docker:
1. Check your Docker daemon settings
2. Configure Docker to use your proxy if behind a corporate firewall
3. Try pulling the image manually:
   ```bash
   docker pull postgres:15-alpine
   ```

### Connection Refused
If you get connection refused errors:
1. Check if the container is running:
   ```bash
   docker ps | grep postgres
   ```
2. Check container logs:
   ```bash
   docker logs ai-knowledge-postgres
   ```
3. Verify port availability:
   ```bash
   netstat -tlnp | grep 5432
   ```