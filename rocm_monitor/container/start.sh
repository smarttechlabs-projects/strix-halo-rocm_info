#!/bin/bash

# ROCm Monitor - Monitoring Stack Startup Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🚀 Starting ROCm Monitor Stack..."

# Check if ROCm Monitor is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ ROCm Monitor is not running!"
    echo "Please start ROCm Monitor first:"
    echo "  cd .."
    echo "  ./rocm-monitor -metrics &"
    exit 1
fi

# Check if metrics endpoint is available
if ! curl -s http://localhost:8080/metrics > /dev/null 2>&1; then
    echo "❌ ROCm Monitor metrics endpoint not found!"
    echo "Please restart ROCm Monitor with -metrics flag:"
    echo "  ./rocm-monitor -metrics &"
    exit 1
fi

echo "✅ ROCm Monitor is running with metrics enabled"

# Create necessary directories
mkdir -p prometheus-data grafana-data alertmanager-data
chmod 777 prometheus-data grafana-data alertmanager-data

# Start the monitoring stack
echo "📊 Starting Prometheus, Grafana, and Alertmanager..."

# Try regular docker-compose first
if docker-compose up -d 2>/dev/null; then
    echo "✅ Started with bridge networking"
    COMPOSE_FILE="docker-compose.yml"
else
    echo "⚠️  Bridge networking failed, trying host networking..."
    
    # Stop any containers that might have started
    docker-compose down 2>/dev/null || true
    
    # Use host networking version
    if docker-compose -f docker-compose-host.yml up -d; then
        echo "✅ Started with host networking"
        COMPOSE_FILE="docker-compose-host.yml"
    else
        echo "❌ Both networking modes failed. Please check Docker installation."
        echo "Try running: sudo systemctl restart docker"
        exit 1
    fi
fi

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 10

# Check service status
echo "🔍 Checking service status..."

# Check Prometheus
if curl -s http://localhost:9090/-/healthy > /dev/null 2>&1; then
    echo "✅ Prometheus is running at http://localhost:9090"
else
    echo "❌ Prometheus failed to start"
    docker-compose logs prometheus
    exit 1
fi

# Check Grafana
if curl -s http://localhost:3000/api/health > /dev/null 2>&1; then
    echo "✅ Grafana is running at http://localhost:3000"
    echo "   Username: admin"
    echo "   Password: admin123"
else
    echo "❌ Grafana failed to start"
    docker-compose logs grafana
    exit 1
fi

# Check Alertmanager
if curl -s http://localhost:9093/-/healthy > /dev/null 2>&1; then
    echo "✅ Alertmanager is running at http://localhost:9093"
else
    echo "❌ Alertmanager failed to start"
    docker-compose logs alertmanager
    exit 1
fi

# Check Prometheus targets
echo "🎯 Checking Prometheus targets..."
sleep 5
if curl -s "http://localhost:9090/api/v1/targets" | grep -q "rocm-monitor"; then
    echo "✅ ROCm Monitor target configured in Prometheus"
else
    echo "⚠️  ROCm Monitor target not found in Prometheus"
    echo "Check the prometheus.yml configuration"
fi

echo ""
echo "🎉 ROCm Monitor Stack is ready!"
echo ""
echo "📊 Access Points:"
echo "   • Grafana:      http://localhost:3000 (admin/admin123)"
echo "   • Prometheus:   http://localhost:9090"
echo "   • Alertmanager: http://localhost:9093"
echo "   • ROCm Monitor: http://localhost:8080"
echo ""
echo "📋 Management Commands:"
echo "   • View logs:    docker-compose logs -f"
echo "   • Stop stack:   docker-compose down"
echo "   • Restart:      docker-compose restart"
echo ""
echo "📈 Pre-configured Dashboard: 'ROCm Monitor - GPU Overview'"
echo "📢 Alert rules active for temperature, VRAM, and system health"