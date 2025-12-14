# ROCm Monitor - Docker Compose Stack

Complete monitoring stack with Prometheus, Grafana, and Alertmanager for ROCm GPU monitoring.

![rocm monitor main.png](../content/rocm%20monitor%20main.png)

The tool also check ROCm-related driver availability etc:
![rocm monitor system test.png](../content/rocm%20monitor%20system%20test.png)

## Stack Components

- **Prometheus** (port 9090) - Metrics collection and alerting
- **Grafana** (port 3000) - Visualization dashboards
- **Alertmanager** (port 9093) - Alert routing and management

## Quick Start

### 1. Prerequisites

```bash
# Ensure ROCm Monitor is built
cd ..
go build -o rocm-monitor

# Start ROCm Monitor with metrics enabled
./rocm-monitor -metrics &
```

### 2. Start Monitoring Stack

```bash
cd container
docker-compose up -d
```

### 3. Access Dashboards

- **Grafana**: http://localhost:3000
  - Username: `admin`
  - Password: `admin123`
- **Prometheus**: http://localhost:9090
- **Alertmanager**: http://localhost:9093

## Configuration

### Prometheus Configuration

The `prometheus.yml` file configures:
- 15-second scrape intervals for real-time monitoring
- ROCm Monitor target on `host.docker.internal:8080`
- Alert rule loading from `prometheus-rules.yml`
- Alertmanager integration

### Grafana Dashboards

**Pre-configured Dashboard**: ROCm Monitor - GPU Overview
- GPU temperature with warning/critical thresholds
- Power consumption monitoring
- GPU & CPU utilization
- VRAM usage and utilization percentage
- Clock frequencies (SCLK/MCLK)
- System health and collection performance

### Alert Rules

**Critical Alerts** (immediate response):
- GPU temperature >85°C for 2 minutes
- VRAM utilization >95% for 1 minute
- ROCm Monitor service down

**Warning Alerts** (action needed soon):
- GPU temperature >75°C for 5 minutes
- VRAM utilization >80% for 3 minutes
- Collection errors (>5 in 10 minutes)
- Performance degradation (>1000ms collection time)
- Clock frequency drops (<500MHz)

**Info Alerts** (awareness):
- ROCm diagnostic test failures
- High power consumption (>50W sustained)

## Data Retention

- **Prometheus**: 7 days of metrics data with 5GB size limit
- **Grafana**: Persistent dashboard and configuration storage
- **Alertmanager**: Alert state persistence

## Customization

### Adding Custom Dashboards

1. Create JSON dashboard files in `grafana/dashboards/`
2. Restart Grafana: `docker-compose restart grafana`

### Modifying Alert Rules

1. Edit `prometheus-rules.yml`
2. Reload Prometheus: `docker-compose kill -s HUP prometheus`

### Alert Routing

Configure email/webhook alerts in `alertmanager.yml`:

```yaml
# Example email configuration
email_configs:
  - to: 'ops-team@company.com'
    subject: 'ROCm Alert: {{ .GroupLabels.alertname }}'
    body: |
      {{ range .Alerts }}
      Alert: {{ .Annotations.summary }}
      Description: {{ .Annotations.description }}
      {{ end }}
```

## Troubleshooting

### ROCm Monitor Not Found

```bash
# Check if ROCm Monitor is running with metrics
curl http://localhost:8080/metrics

# If not running, start it
./rocm-monitor -metrics &
```

### Prometheus Can't Scrape Metrics

```bash
# Check Prometheus targets
# Go to http://localhost:9090/targets
# Ensure "rocm-monitor" target is UP

# If using Docker Desktop on Windows/Mac
# Change in prometheus.yml:
# - targets: ['host.docker.internal:8080']
```

### Grafana Dashboard Empty

```bash
# Check Prometheus data source
# Go to Configuration > Data Sources
# Test connection to Prometheus

# Verify metrics are available
curl http://localhost:9090/api/v1/query?query=rocm_gpu_temperature_celsius
```

### Container Resource Issues

```bash
# Check container resource usage
docker stats

# Increase memory limits in docker-compose.yml if needed
# For Prometheus:
#   deploy:
#     resources:
#       limits:
#         memory: 2G
```

## Production Deployment

### Security Considerations

1. **Change default passwords**:
   ```yaml
   # In docker-compose.yml
   environment:
     - GF_SECURITY_ADMIN_PASSWORD=strong_password_here
   ```

2. **Enable authentication**:
   ```yaml
   # Add to Prometheus command args
   - '--web.enable-admin-api=false'
   - '--web.external-url=https://prometheus.company.com'
   ```

3. **Use external data sources**:
   - External Prometheus cluster
   - Managed Grafana service
   - External alert routing (PagerDuty, Slack)

### Scaling Configuration

```yaml
# For high-frequency monitoring
prometheus.yml:
  scrape_interval: 5s  # More frequent collection
  evaluation_interval: 5s  # Faster alert evaluation

# Storage scaling
docker-compose.yml:
  command:
    - '--storage.tsdb.retention.time=30d'  # Longer retention
    - '--storage.tsdb.retention.size=50GB'  # More storage
```

### High Availability

```yaml
# Multiple Prometheus instances
version: '3.8'
services:
  prometheus-1:
    # Primary instance
  prometheus-2:
    # Secondary instance
  
  grafana:
    environment:
      - GF_DATABASE_TYPE=postgres  # External database
      - GF_DATABASE_HOST=postgres:5432
```

## Official AMD Dashboards

While this custom dashboard provides comprehensive ROCm monitoring, you can also import the official AMD GPU dashboard:

1. **AMD GPU Nodes Dashboard** (Grafana Labs):
   - Dashboard ID: `18913`
   - URL: https://grafana.com/grafana/dashboards/18913-amd-gpu-nodes/

2. **Import via Grafana UI**:
   - Go to Dashboards > Import
   - Enter Dashboard ID: `18913`
   - Configure data source: `Prometheus`

## Maintenance

### Regular Tasks

```bash
# Update containers
docker-compose pull
docker-compose up -d

# Clean up old data
docker volume prune

# Backup Grafana dashboards
docker exec rocm-grafana grafana-cli admin export-dashboard \
  --homeDashboard > backup.json

# Monitor disk usage
df -h /var/lib/docker/volumes/
```

## Prometheus Query Examples

### Testing ROCm Monitor Connection

First, verify that Prometheus can scrape ROCm Monitor metrics:

```promql
# Check if ROCm Monitor is up and running
up{job="rocm-monitor"}

# Get all available ROCm metrics
{__name__=~"rocm_.*"}

# Check last scrape time
prometheus_target_scrape_duration_seconds{job="rocm-monitor"}
```

### Basic GPU Monitoring Queries

**GPU Temperature Monitoring:**
```promql
# Current GPU temperature
rocm_gpu_temperature_celsius

# GPU temperature for specific GPU
rocm_gpu_temperature_celsius{gpu_id="0"}

# Average temperature across all GPUs
avg(rocm_gpu_temperature_celsius)

# Maximum temperature in last 5 minutes
max_over_time(rocm_gpu_temperature_celsius[5m])
```

**GPU Power Consumption:**
```promql
# Current power consumption
rocm_gpu_power_watts

# Total power consumption across all GPUs
sum(rocm_gpu_power_watts)

# Power consumption rate (change over time)
rate(rocm_gpu_power_watts[1m])

# Average power over last hour
avg_over_time(rocm_gpu_power_watts[1h])
```

**GPU Utilization:**
```promql
# Current GPU utilization
rocm_gpu_usage_percent

# GPU utilization above 80%
rocm_gpu_usage_percent > 80

# Average utilization over 10 minutes
avg_over_time(rocm_gpu_usage_percent[10m])

# CPU vs GPU utilization comparison
rocm_system_cpu_usage_percent and rocm_gpu_usage_percent
```

### Memory Monitoring Queries

**VRAM Usage:**
```promql
# VRAM usage in GB
rocm_gpu_vram_usage_gb

# VRAM utilization percentage
rocm_gpu_vram_utilization_percent

# Available VRAM (total - used)
rocm_gpu_vram_total_gb - rocm_gpu_vram_usage_gb

# VRAM efficiency (usage as percentage of total)
(rocm_gpu_vram_usage_gb / rocm_gpu_vram_total_gb) * 100
```

### Performance Analysis Queries

**Clock Frequencies:**
```promql
# System clock frequency
rocm_gpu_sclk_mhz

# Memory clock frequency  
rocm_gpu_mclk_mhz

# Clock frequency changes over time
rate(rocm_gpu_sclk_mhz[5m])

# Performance scaling (frequency vs utilization)
rocm_gpu_sclk_mhz / rocm_gpu_usage_percent
```

**Performance Efficiency:**
```promql
# Performance per watt (efficiency)
rocm_gpu_usage_percent / rocm_gpu_power_watts

# Thermal efficiency (performance per degree)
rocm_gpu_usage_percent / rocm_gpu_temperature_celsius

# Performance density (utilization per GB VRAM)
rocm_gpu_usage_percent / rocm_gpu_vram_total_gb
```

### System Health Monitoring

**Collection Health:**
```promql
# Collection errors per minute
rate(rocm_monitor_collection_errors_total[1m]) * 60

# Collection duration (performance)
rocm_monitor_collection_duration_ms

# Data points collected per second
rate(rocm_monitor_data_points_total[1m])

# Monitor uptime
rocm_monitor_uptime_seconds
```

**Alert Threshold Queries:**
```promql
# Temperature warnings and critical alerts
rocm_gpu_temperature_warning_threshold
rocm_gpu_temperature_critical_threshold

# VRAM high utilization alerts
rocm_gpu_vram_high_utilization

# Combined alert status
rocm_gpu_temperature_critical_threshold + rocm_gpu_vram_high_utilization
```

### Advanced Multi-GPU Queries

**Multi-GPU Aggregations:**
```promql
# Hottest GPU
max(rocm_gpu_temperature_celsius)

# Coldest GPU
min(rocm_gpu_temperature_celsius)

# GPU temperature variance
stddev(rocm_gpu_temperature_celsius)

# Most utilized GPU
max(rocm_gpu_usage_percent)

# Total system power consumption
sum(rocm_gpu_power_watts)
```

**GPU Load Balancing:**
```promql
# GPU utilization difference (load imbalance)
max(rocm_gpu_usage_percent) - min(rocm_gpu_usage_percent)

# GPU power efficiency ranking
rocm_gpu_usage_percent / rocm_gpu_power_watts

# Temperature-normalized performance
rocm_gpu_usage_percent / rocm_gpu_temperature_celsius
```

### Time-Series Analysis Queries

**Trend Analysis:**
```promql
# Temperature trend over 1 hour
increase(rocm_gpu_temperature_celsius[1h])

# Power consumption trend
deriv(rocm_gpu_power_watts[10m])

# VRAM usage growth rate
rate(rocm_gpu_vram_usage_gb[30m])

# Performance consistency (utilization standard deviation)
stddev_over_time(rocm_gpu_usage_percent[1h])
```

**Capacity Planning:**
```promql
# Peak VRAM usage in last 24 hours
max_over_time(rocm_gpu_vram_usage_gb[24h])

# Average power consumption pattern
avg_over_time(rocm_gpu_power_watts[24h])

# Temperature peaks
quantile_over_time(0.95, rocm_gpu_temperature_celsius[24h])

# Utilization 95th percentile
quantile_over_time(0.95, rocm_gpu_usage_percent[24h])
```

### Debugging and Troubleshooting Queries

**Data Validation:**
```promql
# Check for missing metrics (should return data)
absent(rocm_gpu_temperature_celsius)
absent(rocm_system_gpu_count)

# Verify metric freshness (time since last update)
time() - timestamp(rocm_gpu_temperature_celsius)

# Check for stale data (older than 1 minute)
(time() - timestamp(rocm_gpu_temperature_celsius)) > 60
```

**Performance Debugging:**
```promql
# Collection performance issues
rocm_monitor_collection_duration_ms > 1000

# Error rate monitoring
rate(rocm_monitor_collection_errors_total[5m]) > 0

# Memory usage monitoring
rocm_monitor_memory_usage_mb

# Historical data size
rocm_monitor_history_size_points
```

## Testing Queries in Prometheus

1. **Access Prometheus**: http://localhost:9090
2. **Go to Graph tab**
3. **Enter any query above**
4. **Click "Execute" or press Enter**
5. **View results in Table or Graph format**

### Example Test Sequence:

```bash
# 1. Verify connection
curl "http://localhost:9090/api/v1/query?query=up{job=\"rocm-monitor\"}"

# 2. Get current temperature
curl "http://localhost:9090/api/v1/query?query=rocm_gpu_temperature_celsius"

# 3. Get GPU count
curl "http://localhost:9090/api/v1/query?query=rocm_system_gpu_count"

# 4. Get last 5 minutes of power data
curl "http://localhost:9090/api/v1/query_range?query=rocm_gpu_power_watts&start=$(date -d '5 minutes ago' +%s)&end=$(date +%s)&step=15"
```

### Log Monitoring

```bash
# View service logs
docker-compose logs -f prometheus
docker-compose logs -f grafana
docker-compose logs -f alertmanager

# Monitor ROCm Monitor logs
journalctl -u rocm-monitor -f
```