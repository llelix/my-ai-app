-- Create database if it doesn't exist
CREATE DATABASE ai_knowledge_db;

-- Connect to the database
\c ai_knowledge_db;

-- Create tables will be handled by GORM auto-migration
-- This file can be extended with custom initializations if needed