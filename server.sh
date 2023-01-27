#!/bin/bash

# Set the port
PORT=5000

# Start backend
./ogree_app_backend &

# switch directories
cd build/web/

# Start the frontend server
echo 'Server starting on port' $PORT '...'
python3 -m http.server $PORT
