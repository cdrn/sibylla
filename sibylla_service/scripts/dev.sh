#!/bin/bash
# Check for Reflex and start dev server
if ! command -v reflex &> /dev/null
then
    echo "Reflex could not be found. Install it with 'go install github.com/cespare/reflex@latest'"
    exit
else
    echo "Reflex is installed."
fi

# Start Reflex to watch .go files and reload on change
echo "Starting Reflex to watch .go files..."
reflex -r '\.go$' -s -- go run cmd/sibylla_service/main.go
