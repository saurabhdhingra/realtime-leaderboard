#!/bin/bash

# Ensure Python dependencies are installed
pip install -r requirements.txt

# Start the FastAPI application
# Uses the main.py file from the api directory
echo "Starting FastAPI server..."
uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload