#!/bin/bash
# Quick start script for Azure DevOps agent in Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "owlctl Azure DevOps Agent - Quick Start"
echo "=========================================="
echo ""

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "⚠ .env file not found"
    echo ""
    echo "Creating .env from template..."
    cp .env.example .env
    echo "✓ Created .env file"
    echo ""
    echo "📝 Please edit .env and fill in your values:"
    echo "   - AZP_URL: Your Azure DevOps organization URL"
    echo "   - AZP_TOKEN: Personal Access Token (Agent Pools: Read & Manage)"
    echo "   - OWLCTL_USERNAME: VBR username"
    echo "   - OWLCTL_PASSWORD: VBR password"
    echo "   - OWLCTL_URL: VBR server URL"
    echo ""
    echo "After editing .env, run this script again."
    exit 0
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "✗ Docker is not running"
    echo "Please start Docker Desktop and try again"
    exit 1
fi

echo "✓ Docker is running"

# Check if required variables are set
source .env

MISSING_VARS=0

if [ -z "$AZP_URL" ]; then
    echo "✗ AZP_URL not set in .env"
    MISSING_VARS=1
fi

if [ -z "$AZP_TOKEN" ]; then
    echo "✗ AZP_TOKEN not set in .env"
    MISSING_VARS=1
fi

if [ $MISSING_VARS -eq 1 ]; then
    echo ""
    echo "Please edit .env and fill in the required values"
    exit 1
fi

echo "✓ Configuration loaded"

# Check if agent is already running
if docker compose ps | grep -q "owlctl-azure-agent.*Up"; then
    echo "⚠ Agent is already running"
    echo ""
    read -p "Do you want to restart it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Stopping existing agent..."
        docker compose down
    else
        echo "Keeping existing agent running"
        echo ""
        echo "View logs: docker compose logs -f"
        exit 0
    fi
fi

# Build and start
echo ""
echo "Building agent image..."
docker compose build

echo ""
echo "Starting agent..."
docker compose up -d

echo ""
echo "Waiting for agent to register..."
sleep 5

# Show logs
echo ""
echo "=========================================="
echo "Agent Status"
echo "=========================================="
docker compose logs azure-agent | tail -20

echo ""
echo "=========================================="
echo "Next Steps"
echo "=========================================="
echo ""
echo "1. Verify agent is online:"
echo "   Go to: $AZP_URL/_settings/agentpools"
echo "   Pool: ${AZP_POOL:-Default}"
echo "   Look for: ${AZP_AGENT_NAME:-owlctl-docker-agent}"
echo ""
echo "2. Test with the example pipeline:"
echo "   Import test-pipeline.yml to your Azure DevOps project"
echo ""
echo "3. View live logs:"
echo "   docker compose logs -f azure-agent"
echo ""
echo "4. Stop agent:"
echo "   docker compose down"
echo ""
echo "=========================================="
