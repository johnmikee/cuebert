#!/bin/bash

# Database migration script for applying critical fixes
# This script applies the database improvements in a safe, controlled manner

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default database configuration (override with environment variables)
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-cue}
DB_USER=${DB_USER:-cue}
DB_PASSWORD=${DB_PASSWORD:-cue}

# Migration directory
MIGRATION_DIR="./sql"

echo -e "${GREEN}Database Migration Script${NC}"
echo "================================"
echo "Host: $DB_HOST:$DB_PORT"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo ""

# Function to run a migration
run_migration() {
    local migration_file=$1
    local description=$2
    
    echo -e "${YELLOW}Applying: $description${NC}"
    echo "File: $migration_file"
    
    if [ ! -f "$MIGRATION_DIR/$migration_file" ]; then
        echo -e "${RED}Error: Migration file not found: $migration_file${NC}"
        return 1
    fi
    
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
        -f "$MIGRATION_DIR/$migration_file" \
        -v ON_ERROR_STOP=1 \
        --single-transaction
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Successfully applied: $description${NC}"
    else
        echo -e "${RED}✗ Failed to apply: $description${NC}"
        return 1
    fi
    echo ""
}

# Function to check if migration was already applied
check_migration_applied() {
    local migration_number=$1
    
    # Check if migration history table exists and has this migration
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
        -tAc "SELECT 1 FROM schema_migrations WHERE version >= $migration_number LIMIT 1" 2>/dev/null || echo "0"
}

# Create backup before starting
echo -e "${YELLOW}Creating backup...${NC}"
BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
PGPASSWORD=$DB_PASSWORD pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > $BACKUP_FILE

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Backup created: $BACKUP_FILE${NC}"
else
    echo -e "${RED}✗ Failed to create backup${NC}"
    exit 1
fi
echo ""

# Ask for confirmation
echo -e "${YELLOW}This script will apply the following migrations:${NC}"
echo "  1. Fix SQL injection vulnerabilities (00009)"
echo "  2. Add foreign key constraints (00009)"
echo "  3. Fix data type inconsistencies (00010)"
echo "  4. Add performance indexes (00011)"
echo "  5. Implement timezone-aware timestamps (00012)"
echo "  6. Create unified tracking table (00013)"
echo "  7. Add audit logging system (00014)"
echo ""
read -p "Do you want to proceed? (yes/no): " -r
echo ""

if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Migration cancelled."
    exit 0
fi

# Track if any migration fails
MIGRATION_FAILED=0

# Apply migrations in order

# Phase 1: Critical Security and Integrity Fixes
echo -e "${GREEN}=== Phase 1: Critical Fixes ===${NC}"

run_migration "00009_add_foreign_keys.up.sql" \
    "Add foreign key constraints for data integrity" || MIGRATION_FAILED=1

run_migration "00010_fix_data_types.up.sql" \
    "Fix data type inconsistencies" || MIGRATION_FAILED=1

# Phase 2: Performance Optimization
echo -e "${GREEN}=== Phase 2: Performance Optimization ===${NC}"

run_migration "00011_add_performance_indexes.up.sql" \
    "Add missing indexes for query performance" || MIGRATION_FAILED=1

# Phase 3: Data Quality Improvements
echo -e "${GREEN}=== Phase 3: Data Quality ===${NC}"

run_migration "00012_timezone_aware_timestamps.up.sql" \
    "Implement timezone-aware timestamps" || MIGRATION_FAILED=1

# Phase 4: Architectural Improvements (Optional)
echo -e "${GREEN}=== Phase 4: Architectural Improvements ===${NC}"
echo -e "${YELLOW}The following migrations are optional but recommended:${NC}"
read -p "Apply unified tracking table? (yes/no): " -r
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    run_migration "00013_unified_tracking_table.up.sql" \
        "Create unified tracking table" || MIGRATION_FAILED=1
fi

read -p "Apply audit logging system? (yes/no): " -r
if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    run_migration "00014_audit_logging.up.sql" \
        "Add comprehensive audit logging" || MIGRATION_FAILED=1
fi

# Summary
echo ""
echo "================================"
if [ $MIGRATION_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All migrations completed successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Test the application thoroughly"
    echo "2. Monitor query performance"
    echo "3. Verify data integrity"
    echo "4. Update application code to use new features"
else
    echo -e "${RED}✗ Some migrations failed${NC}"
    echo ""
    echo "To rollback, run:"
    echo "  psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME < $BACKUP_FILE"
fi

echo ""
echo "Backup file: $BACKUP_FILE"