#!/bin/bash

set -e               # Exit on error
cd "$(dirname "$0")" # Go to script directory

IMAGE_NAME="d.isotronic.de/project/apartmenthunter"
TAG="latest"

echo "Building Go binary..."
GOOS=linux GOARCH=amd64 go build -o app ./cmd

echo "Building Docker image..."
docker build -t "$IMAGE_NAME:$TAG" .

echo "Pushing Docker image..."
docker push "$IMAGE_NAME:$TAG"

# Restart the Nomad job
echo "Restarting Nomad job..."
nomad job stop projects.apartment-hunter || true
sleep 5  # Wait a few seconds to allow push propagation
nomad job run apartment-hunter.nomad

echo "Deployment completed!"

