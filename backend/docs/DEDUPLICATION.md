# File Deduplication Implementation

## Overview

The file deduplication system prevents storing duplicate files in MinIO/S3-compatible storage by using file hash and size as deduplication keys. When a duplicate file is uploaded, the system creates a new document record that references the existing physical file instead of storing a new copy.

## Key Features

### 1. Hash-Based Deduplication
- Files are identified by their SHA-256 hash and file size
- When a file with the same hash and size is uploaded, it's considered a duplicate
- Only one physical copy is stored in MinIO/S3 storage

### 2. Reference Counting
- Each document record has a `ref_count` field that tracks how many documents reference the same physical file
- When a duplicate is detected, a new document record is created with the same `file_path` (S3 object key)
- The original document's reference count is incremented

### 3. Safe Deletion
- When a document is deleted, the system checks if other documents reference the same file
- The physical file is only removed from storage when no other documents reference it
- This prevents accidental deletion of files that are still needed by other documents

## Implementation Details

### Database Schema Changes
- Added `ref_count` column to the `documents` table (default: 1)
- Added indexes on `file_hash`, `file_size` for efficient deduplication queries
- Added index on `ref_count` for cleanup operations

### Key Methods

#### `CheckFile(hash, size)`
- Checks if a file with the given hash and size already exists
- Returns the existing document if found

#### `CreateDuplicateReference(originalDoc, fileName, originalName)`
- Creates a new document record that references an existing file
- Increments the reference count of the original document
- Verifies file integrity before creating the reference

#### `VerifyObjectIntegrity(filePath, expectedHash)`
- Verifies that an object exists in storage and matches the expected hash
- Used to ensure file integrity before creating duplicate references

#### `Delete(id)`
- Implements reference-counted deletion
- Only removes the physical file when no other documents reference it
- Uses database transactions to ensure consistency

#### `GetDeduplicationStats()`
- Returns statistics about file deduplication
- Shows total documents vs unique files
- Calculates space saved through deduplication

#### `CleanupOrphanedObjects()`
- Removes objects from storage that have no database references
- Helps maintain storage consistency

## Usage Examples

### Upload Duplicate File
```go
// First upload
doc1, err := service.Upload(file1) // Creates new file in storage

// Second upload with same content
doc2, err := service.Upload(file2) // Creates duplicate reference, no new storage
```

### Check Deduplication Stats
```go
stats, err := service.GetDeduplicationStats()
// Returns: total_documents, unique_files, space_saved_bytes, etc.
```

### Safe Deletion
```go
// Delete first document
err := service.Delete(doc1.ID) // File remains in storage

// Delete second document  
err := service.Delete(doc2.ID) // File is removed from storage
```

## Benefits

1. **Storage Efficiency**: Eliminates duplicate files, saving storage space
2. **Cost Reduction**: Reduces storage costs in cloud environments
3. **Performance**: Faster uploads for duplicate files (instant deduplication)
4. **Data Integrity**: Ensures files are not accidentally deleted while still referenced
5. **Transparency**: Works seamlessly with existing API endpoints

## Migration

The system includes database migrations to add the `ref_count` column to existing installations:
- `002_add_deduplication_support.sql` (PostgreSQL)
- `002_add_deduplication_support_sqlite.sql` (SQLite)

Existing documents are automatically assigned `ref_count = 1` during migration.