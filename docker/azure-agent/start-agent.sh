#!/bin/bash
set -e

# Validate required environment variables
if [ -z "$AZP_URL" ]; then
  echo "ERROR: Missing AZP_URL environment variable"
  echo "Set it to your Azure DevOps organization URL (e.g., https://dev.azure.com/yourorg)"
  exit 1
fi

if [ -z "$AZP_TOKEN" ]; then
  echo "ERROR: Missing AZP_TOKEN environment variable"
  echo "Create a Personal Access Token (PAT) with Agent Pools (read, manage) scope"
  exit 1
fi

# Set agent name if not provided
if [ -z "$AZP_AGENT_NAME" ]; then
  AZP_AGENT_NAME="docker-agent-$(hostname)"
fi

echo "=========================================="
echo "Azure DevOps Agent Configuration"
echo "=========================================="
echo "Organization URL: $AZP_URL"
echo "Agent Pool: $AZP_POOL"
echo "Agent Name: $AZP_AGENT_NAME"
echo "Go Version: $(go version)"
echo "=========================================="

# Clean up any previous agent configuration
if [ -f "/home/agent/.agent" ]; then
  echo "Removing previous agent registration..."
  ./config.sh remove --unattended --auth pat --token "$AZP_TOKEN" || true
fi

# Configure the agent
echo "Configuring Azure DevOps agent..."
./config.sh \
  --unattended \
  --url "$AZP_URL" \
  --auth pat \
  --token "$AZP_TOKEN" \
  --pool "$AZP_POOL" \
  --agent "$AZP_AGENT_NAME" \
  --acceptTeeEula \
  --replace

# Cleanup function for graceful shutdown
cleanup() {
  echo "Removing agent..."
  ./config.sh remove --unattended --auth pat --token "$AZP_TOKEN" || true
}

trap cleanup EXIT SIGTERM SIGINT

# Start the agent
echo "Starting Azure DevOps agent..."
./run.sh "$@" & wait $!
