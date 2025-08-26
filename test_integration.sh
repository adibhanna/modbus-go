#!/bin/bash

# Start the server in the background
echo "Starting TCP server..."
go run examples/tcp_server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Run the client
echo ""
echo "Starting TCP client..."
go run examples/tcp_client/main.go

# Kill the server
echo ""
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null

echo "Test complete!"