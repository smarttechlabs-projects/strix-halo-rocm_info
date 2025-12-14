# Docker iptables Issue - Quick Fixes

## Issue
```
iptables failed: iptables --wait -t filter -A DOCKER-ISOLATION-STAGE-1 -i br-xxx ! -o br-xxx -j DOCKER-ISOLATION-STAGE-2: iptables v1.8.10 (nf_tables): Chain 'DOCKER-ISOLATION-STAGE-2' does not exist
```

## Quick Fixes (try in order)

### Option 1: Restart Docker
```bash
sudo systemctl restart docker
cd /home/jfey/Projects/rocm-info/rocm_monitor/container
./start.sh
```

### Option 2: Use Host Networking (if Option 1 fails)
```bash
cd /home/jfey/Projects/rocm-info/rocm_monitor/container
docker-compose -f docker-compose-host.yml up -d
```

### Option 3: Fix iptables (if both fail)
```bash
cd /home/jfey/Projects/rocm-info/rocm_monitor/container
sudo ./fix-iptables.sh
./start.sh
```

### Option 4: Manual Container Start
```bash
# Start each service individually
docker run -d --name rocm-prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml:ro \
  -v $(pwd)/prometheus-rules.yml:/etc/prometheus/rules.yml:ro \
  prom/prometheus:v2.45.0

docker run -d --name rocm-grafana \
  -p 3000:3000 \
  -e GF_SECURITY_ADMIN_PASSWORD=admin123 \
  -v $(pwd)/grafana/provisioning:/etc/grafana/provisioning:ro \
  -v $(pwd)/grafana/dashboards:/var/lib/grafana/dashboards:ro \
  grafana/grafana:10.1.0
```

## Verification
After any fix:
```bash
# Check if services are running
curl http://localhost:3000/api/health  # Grafana
curl http://localhost:9090/-/healthy   # Prometheus

# Access URLs:
# Grafana: http://localhost:3000 (admin/admin123)
# Prometheus: http://localhost:9090
```