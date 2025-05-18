#!/bin/bash

# Reset WLTrak SQLite Database
#
# This script removes the local SQLite database file and its
# containing directory, allowing for a clean start.

echo "WLTrak Database Reset Tool"
echo "=========================="

DB_PATH="data/wltrak.db"
DATA_DIR="data"

# Check if the database file exists
if [ -f "$DB_PATH" ]; then
    echo "Found database at $DB_PATH"
    
    # Ask for confirmation
    read -p "Are you sure you want to delete the database? This will remove all your data. (y/N): " confirm
    
    if [[ "$confirm" == [yY] || "$confirm" == [yY][eE][sS] ]]; then
        # Delete the database file
        echo "Removing database file..."
        rm "$DB_PATH"
        echo "Database file removed."
        
        # Check if data directory is empty and remove if it is
        if [ -z "$(ls -A "$DATA_DIR" 2>/dev/null)" ]; then
            echo "Data directory is empty, removing it..."
            rmdir "$DATA_DIR"
            echo "Data directory removed."
        else
            echo "Data directory contains other files, keeping the directory."
        fi
        
        echo "Database reset complete. A new database will be created when you next run the application."
    else
        echo "Database reset cancelled."
    fi
else
    echo "Database file does not exist at $DB_PATH"
    
    # Check if the data directory exists but is empty
    if [ -d "$DATA_DIR" ] && [ -z "$(ls -A "$DATA_DIR" 2>/dev/null)" ]; then
        echo "Empty data directory found, removing it..."
        rmdir "$DATA_DIR"
        echo "Data directory removed."
    elif [ -d "$DATA_DIR" ]; then
        echo "Data directory exists and contains files."
    else
        echo "Data directory does not exist."
    fi
    
    echo "No action needed. A new database will be created when you run the application."
fi

echo "=========================="