#!/bin/bash

set -e

# Find the database file
DB_FILE=""
POSSIBLE_LOCATIONS=(
  "wltrack.db"
  "data/wltrack.db"
  "../data/wltrack.db"
)

for loc in "${POSSIBLE_LOCATIONS[@]}"; do
  if [ -f "$loc" ]; then
    DB_FILE="$loc"
    break
  fi
done

if [ -z "$DB_FILE" ]; then
  echo "Error: Database file not found. Please specify the path with -d option."
  echo "Usage: $0 [-d db_path] [-r] [-m] [-s]"
  exit 1
fi

# Set defaults
DO_RESET=0
DO_MIGRATE=0
DO_SAMPLE=0

# Process command line arguments
while getopts "d:rms" opt; do
  case $opt in
    d) DB_FILE="$OPTARG" ;;
    r) DO_RESET=1 ;;
    m) DO_MIGRATE=1 ;;
    s) DO_SAMPLE=1 ;;
    *)
      echo "Usage: $0 [-d db_path] [-r] [-m] [-s]"
      echo "  -d path   : Specify database path"
      echo "  -r        : Reset database (WARNING: deletes all data)"
      echo "  -m        : Migrate existing database to new schema"
      echo "  -s        : Add sample data"
      exit 1
      ;;
  esac
done

# Ensure database exists
if [ ! -f "$DB_FILE" ] && [ $DO_RESET -eq 0 ] && [ $DO_MIGRATE -eq 0 ]; then
  echo "Database file does not exist at $DB_FILE."
  echo "Would you like to create a new database? [y/N]"
  read -r answer
  if [[ "$answer" =~ ^[Yy]$ ]]; then
    DO_RESET=1
  else
    echo "Exiting without changes."
    exit 0
  fi
fi

# Ensure user wants to proceed with potentially destructive operations
if [ $DO_RESET -eq 1 ]; then
  echo "WARNING: You are about to RESET the database at $DB_FILE."
  echo "This will DELETE ALL DATA. Are you sure? [y/N]"
  read -r answer
  if [[ ! "$answer" =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
  fi
fi

if [ $DO_MIGRATE -eq 1 ] && [ $DO_RESET -eq 1 ]; then
  echo "Error: Cannot use -r and -m together. Choose either reset (-r) or migrate (-m)."
  exit 1
fi

# Backup the database
if [ -f "$DB_FILE" ]; then
  BACKUP_FILE="${DB_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
  echo "Creating backup at $BACKUP_FILE"
  cp "$DB_FILE" "$BACKUP_FILE"
fi

# Reset database if requested
if [ $DO_RESET -eq 1 ]; then
  echo "Resetting database schema..."
  sqlite3 "$DB_FILE" < "reset_database.sql"
fi

# Migrate database if requested
if [ $DO_MIGRATE -eq 1 ]; then
  echo "Migrating database schema..."
  sqlite3 "$DB_FILE" < "migrate_schema.sql"
fi

# Add sample data if requested
if [ $DO_SAMPLE -eq 1 ]; then
  echo "Adding sample data..."
  sqlite3 "$DB_FILE" < "sample_data.sql"
fi

echo "Schema changes applied successfully to $DB_FILE"
echo "Done!"
