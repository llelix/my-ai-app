-- init.sql

-- 创建数据库 (如果不存在)
-- 注意: docker-compose中的POSTGRES_DB环境变量通常会自动创建数据库
-- 这里显式写出是为了完整性和可移植性
SELECT 'CREATE DATABASE ai_knowledge_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ai_knowledge_db')\gexec

-- 连接到新创建的数据库
\c ai_knowledge_db;

-- 安装扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";
CREATE EXTENSION IF NOT EXISTS "vector";

-- 设置时区
ALTER DATABASE ai_knowledge_db SET timezone TO 'Asia/Shanghai';

-- 创建一个初始表 (可选, GORM的AutoMigrate会处理)
-- 这里可以留空，让GORM来管理表结构
-- 或者创建一个简单的版本表来跟踪数据库初始化状态
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT NOT NULL,
    dirty BOOLEAN NOT NULL,
    PRIMARY KEY (version)
);

-- 可以在这里添加其他数据库级别的初始化操作
-- 例如: 创建特定的角色、模式(schema)等

-- 提示: GORM的AutoMigrate会自动创建表结构
-- 所以这里不需要手动创建categories, tags, knowledges等表
-- 除非有特殊的列类型或约束是GORM不支持的

-- 结束
