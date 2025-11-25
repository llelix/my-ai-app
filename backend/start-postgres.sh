#!/bin/bash

# Script to start PostgreSQL container for the AI Knowledge Base app

echo "Starting PostgreSQL container..."

# Run PostgreSQL container
docker run -d \
  --name ai-knowledge-postgres \
  -e POSTGRES_DB=ai_knowledge_db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  postgres:15-alpine

echo "PostgreSQL container started!"
echo "Container name: ai-knowledge-postgres"
echo "Database: ai_knowledge_db"
echo "User: postgres"
echo "Password: password"
echo "Port: 5432"

echo ""
echo "To check if the container is running:"
echo "  docker ps | grep ai-knowledge-postgres"
echo ""
echo "To stop the container:"
echo "  docker stop ai-knowledge-postgres"
echo ""
echo "To remove the container:"
echo "  docker rm ai-knowledge-postgres"