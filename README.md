# ROCm GPU Monitor

A real-time GPU monitoring tool for AMD GPUs using ROCm, with a web-based dashboard for 
visualization and data export. Providing /metrics endpoint for Prometheus and Grafana
for related dashboards.

![rocm monitor main.png](rocm_monitor/content/rocm%20monitor%20main.png)

The tool also check ROCm-related driver availability etc:
![rocm monitor system test.png](rocm_monitor/content/rocm%20monitor%20system%20test.png)

## Stack Components

o check ROCm-related driver availability etc:

## Features

- ✅ Real-time GPU monitoring (temperature, power, usage, VRAM)
- ✅ Multi-GPU support with individual GPU selection
- ✅ Web-based dashboard with interactive charts
- ✅ **ROCm System Diagnostics** - Comprehensive ROCm installation testing
- ✅ Data export (CSV, JSON, Prometheus metrics)
- ✅ Configurable monitoring intervals
- ✅ Time-windowed data views (5min, 15min, 30min, 1h, all)
- ✅ Dark/Light theme support
- ✅ REST API for programmatic access
- ✅ Graceful error handling and retry logic
- ✅ Command-line configuration options

## Requirements

- Ubuntu 24.04 LTS (or compatible Linux distribution)
- ROCm 6.4.x installed and configured
- AMD GPU(s) supported by ROCm
- Go 1.19+ (for building from source)
- `rocm-smi` command available in PATH

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/smarttechlabs-projects/strix-halo-rocm_info.git
cd rocm_monitor

# Build the application
go build -o rocm-monitor

# Run the monitor
./rocm-monitor
```

### Using Make

```bash
# Build
make build

# Run
make run

# Clean
make clean
```

## Usage

### Basic Usage

```bash
# Start with default settings
./rocm-monitor

# Custom port
./rocm-monitor -port 9090

# Custom monitoring interval
./rocm-monitor -interval 10s

# Enable Prometheus metrics
./rocm-monitor -metrics

# Restrict CORS origin
./rocm-monitor -cors "http://localhost:3000"
```

### Command-line Options

```
-port int
    HTTP server port (default 8080)
-interval duration
    Collection interval (default 5s)
-history int
    Maximum history size (default 1000)
-cors string
    CORS allowed origin (default "*")
-metrics
    Enable Prometheus metrics endpoint
```

## API Endpoints

### Data Endpoints

- `GET /api/stats` - Get full history of GPU statistics
- `GET /api/stats?window=5m` - Get statistics for specific time window
- `GET /api/latest` - Get only the latest data point
- `GET /api/health` - Health check endpoint
- `GET /api/config` - Get current configuration
- `POST /api/config` - Update configuration (interval)

### Export Endpoints

- `GET /api/export.csv` - Export data as CSV
- `GET /api/export.json` - Export data as JSON
- `GET /metrics` - Comprehensive Prometheus metrics for Grafana integration (if enabled)

### Testing Endpoints

- `POST /api/rocm-test` - Run comprehensive ROCm system diagnostics

### Example API Usage

```bash
# Get latest GPU data
curl http://localhost:8080/api/latest

# Get last 5 minutes of data
curl http://localhost:8080/api/stats?window=5m

# Update monitoring interval
curl -X POST http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{"interval": "10s"}'

# Export data as CSV
curl http://localhost:8080/api/export.csv > gpu_data.csv
```

## Web Dashboard

Access the comprehensive monitoring dashboard at `http://localhost:8080`

### 📊 Dashboard Overview

The web interface provides real-time monitoring with professional charts and comprehensive system diagnostics.

#### Header Controls
- **🔧 Test ROCm** - Launch comprehensive system diagnostics
- **📥 Export CSV** - Download monitoring data in CSV format
- **📥 Export JSON** - Download monitoring data in JSON format
- **Interval Selector** - Choose collection frequency (1s, 5s, 10s, 30s, 60s)
- **Time Window** - Select data range (5min, 15min, 30min, 1h, all)
- **🌓 Theme** - Toggle between dark and light themes

#### Status Indicators
- **Connection Status** - Green dot indicates active connection to ROCm monitor
- **GPU Count** - Shows number of detected GPUs (e.g., "1 GPU detected")
- **Error Messages** - Red banner appears if connection or data issues occur

#### GPU Information Panel
Displays static GPU information when available:
- **Product Name** - GPU model (e.g., "AMD Radeon Graphics")
- **Vendor** - GPU manufacturer
- **Serial Number** - Hardware serial (if available)
- **VRAM Vendor** - Memory manufacturer
- **Bus Info** - PCIe bus location
- **Firmware Versions** - Various firmware component versions

#### Real-Time Charts

**🌡️ Temperature Chart**
- Displays GPU edge temperature in Celsius
- Typical range: 30-85°C
- Red zone: >80°C indicates potential thermal issues
- Multiple GPU support with color-coded lines

**⚡ Power Consumption Chart**  
- Shows real-time power usage in Watts
- Includes socket-level power measurements
- Useful for monitoring power efficiency and thermal design power (TDP)
- Helps identify power-hungry workloads

**🎮 GPU Usage Chart**
- GPU utilization percentage (0-100%)
- Indicates how busy the GPU cores are
- Useful for performance monitoring and bottleneck identification
- Shows compute workload intensity

**🖥️ CPU Usage Chart**
- Overall system CPU utilization percentage
- Helps correlate GPU and CPU workloads
- Useful for identifying system bottlenecks
- Single line showing aggregate CPU usage

**⏱️ GPU Clock Frequencies Chart**
- **SCLK** - System/Shader Clock (solid lines)
- **MCLK** - Memory Clock (dashed lines)  
- Frequencies shown in MHz
- Indicates performance states and boost behavior
- Multiple clock domains for different GPU functions

**💾 VRAM Usage Chart**
- Video memory utilization in GB
- Shows used vs. total VRAM capacity
- Critical for memory-intensive applications
- Helps prevent out-of-memory conditions

#### Interactive Features

**Multi-GPU Selection**
- Checkbox controls appear when multiple GPUs detected
- "All" checkbox toggles all GPUs simultaneously
- Individual GPU checkboxes for selective monitoring
- Real-time chart updates when selection changes

**Chart Interactions**
- **Hover** - Shows exact values and timestamps in tooltips
- **Responsive Design** - Adapts to screen size and mobile devices
- **Smooth Updates** - Real-time data with minimal animation delay
- **Auto-Scaling** - Y-axis automatically adjusts to data ranges

### 🔧 ROCm System Diagnostics

**Quick Access** - Click the prominent "🔧 Test ROCm" button in the dashboard header

**What It Does:**
- Runs 11 comprehensive diagnostic tests
- Validates ROCm installation and configuration
- Identifies common issues and provides solutions
- Tests take 10-30 seconds to complete
- Results displayed in professional modal popup

**When to Use:**
- After ROCm installation or updates
- When experiencing GPU performance issues
- Before deploying GPU workloads
- When troubleshooting system problems
- For system validation and health checks

## ROCm System Diagnostics

The built-in ROCm Test feature provides comprehensive diagnostics to verify your ROCm installation and identify configuration issues. Access via the "🔧 Test ROCm" button in the dashboard or API endpoint.

### Test Suite Overview

The diagnostics run **11 comprehensive tests** covering all aspects of ROCm functionality:

#### 🔍 Core System Tests

**1. ROCm Info** - `rocminfo`
- **Purpose**: Validates HSA (Heterogeneous System Architecture) runtime and device detection
- **What it checks**: 
  - ROCk kernel module loading status
  - HSA runtime version and capabilities  
  - GPU agents and their properties
  - Memory pools and accessibility
  - Instruction Set Architecture (ISA) support
- **Success indicators**: 
  - "ROCk module is loaded" message present
  - HSA Agents section with detected devices
  - GPU device type properly identified
  - Runtime version information available
- **Typical duration**: 100-200ms
- **Critical for**: Verifying basic ROCm installation

**2. ROCm SMI** - `rocm-smi`
- **Purpose**: Basic GPU information and operational status
- **What it checks**:
  - GPU device detection and enumeration
  - Basic metrics availability (temperature, power, usage)
  - Device power state and performance levels
  - Overall system health
- **Success indicators**: 
  - GPU devices listed with IDs
  - Temperature readings available
  - Power consumption data
  - Performance metrics accessible
- **Common warnings**: "GPU in low-power state" (normal when idle)
- **Typical duration**: 50-100ms
- **Critical for**: Basic GPU functionality verification

**3. ROCm SMI Detailed** - `rocm-smi -a`
- **Purpose**: Comprehensive GPU metrics and detailed hardware information
- **What it checks**:
  - All available GPU sensors and metrics
  - Hardware identification (Device ID, VBIOS, PCIe info)
  - Clock frequencies and voltage information
  - Memory subsystem details
  - Firmware versions and capabilities
- **Success indicators**:
  - Detailed GPU information displayed
  - Multiple metric categories available
  - Hardware identifiers present
- **Common informational messages** (not errors):
  - "Clock exists but EMPTY! Likely driver error!" - normal on APUs
  - "Not supported on the given system" - expected for many APU features
  - "Failed to retrieve GPU metrics" - normal when metric version unsupported
- **Typical duration**: 100-150ms
- **Critical for**: Deep hardware analysis and troubleshooting

#### 📋 Device Enumeration Tests

**4. GPU List** - `rocm-smi -l`
- **Purpose**: Enumerate all available ROCm-compatible GPU devices
- **What it checks**:
  - Device discovery and listing
  - GPU accessibility by ROCm stack
  - Device power profiles and capabilities
- **Success indicators**: Clean device enumeration
- **Typical duration**: 50-80ms
- **Critical for**: Multi-GPU system validation

#### 🌡️ Sensor and Monitoring Tests

**5. Temperature Monitoring** - `rocm-smi -t`
- **Purpose**: Validate thermal sensor functionality
- **What it checks**:
  - Temperature sensor availability
  - Thermal reading accuracy
  - Sensor communication with driver
- **Success indicators**: Temperature values in reasonable range (20-90°C)
- **Typical duration**: 50-80ms
- **Critical for**: Thermal management verification

**6. Power Monitoring** - `rocm-smi -p`
- **Purpose**: Verify power measurement capabilities
- **What it checks**:
  - Power sensor functionality
  - Performance level reporting
  - Power management features
- **Success indicators**: Power readings and performance levels
- **Typical duration**: 50-80ms
- **Critical for**: Power management validation

**7. Clock Frequency Monitoring** - `rocm-smi -c`
- **Purpose**: Validate clock frequency reporting and control
- **What it checks**:
  - System clock (SCLK) frequency reporting
  - Memory clock (MCLK) frequency reporting  
  - System-on-chip clock (SOCCLK) information
  - Dynamic frequency scaling
- **Success indicators**: Clock frequencies reported in MHz
- **Typical duration**: 50-80ms
- **Critical for**: Performance monitoring capabilities

**8. Memory Usage Monitoring** - `rocm-smi -u`
- **Purpose**: Validate VRAM usage reporting
- **What it checks**:
  - Video memory utilization reporting
  - VRAM capacity detection
  - Memory controller communication
- **Success indicators**: VRAM usage percentages
- **Typical duration**: 50-80ms
- **Critical for**: Memory management validation

#### 🏗️ Runtime and Platform Tests

**9. HIP Version Check** - `hipconfig --version`
- **Purpose**: Verify HIP (Heterogeneous-Compute Interface for Portability) runtime
- **What it checks**:
  - HIP runtime installation
  - Version compatibility
  - Runtime library availability
- **Expected output**: Version number format (e.g., "7.1.25424-4179531dcd")
- **Success indicators**: Valid version string returned
- **Typical duration**: 30-50ms
- **Critical for**: HIP application compatibility

**10. HIP Platform Detection** - `hipconfig --platform`
- **Purpose**: Validate HIP platform configuration  
- **What it checks**:
  - Platform backend detection (AMD vs NVIDIA)
  - Runtime configuration validity
  - Backend library availability
- **Expected output**: Platform identifier ("amd", "nvidia", etc.)
- **Success indicators**: Valid platform name returned
- **Typical duration**: 30-50ms
- **Critical for**: Platform-specific optimization

**11. Device Capability Query** - `rocminfo` (second run)
- **Purpose**: Detailed device capabilities and architecture verification
- **What it checks**:
  - Complete device feature enumeration
  - Architecture-specific capabilities
  - Memory hierarchy and access patterns
  - Compute unit organization
- **Success indicators**: Detailed capability information
- **Typical duration**: 100-200ms
- **Critical for**: Application optimization and compatibility

### How to Use ROCm Tests

#### Via Web Dashboard
1. **Open Dashboard** - Navigate to `http://localhost:8080`
2. **Click Test Button** - Click "🔧 Test ROCm" in the header
3. **Wait for Results** - Tests run for 10-30 seconds
4. **Review Output** - Click individual tests to expand details

#### Via API
```bash
# Run diagnostics via API
curl -X POST http://localhost:8080/api/rocm-test

# Example response structure:
{
  "overall_success": true,
  "summary": "✅ All ROCm tests passed - Tests: 11 total, 11 passed, 0 failed (completed in 1240ms)",
  "test_results": [
    {
      "command": "rocminfo",
      "success": true,
      "output": "[full command output]",
      "duration_ms": 156,
      "issues": [],
      "summary": "✅ ROCm Info - Success"
    }
  ]
}
```

### Understanding Test Results

#### 🎯 Overall Test Summary

The test modal displays a comprehensive summary at the top:

**✅ Green Summary Examples:**
- "✅ All ROCm tests passed - Tests: 11 total, 11 passed, 0 failed (completed in 1240ms)"
- Indicates: ROCm is fully functional and properly configured

**⚠️ Yellow Summary Examples:**  
- "⚠️ ROCm tests passed with warnings - Tests: 11 total, 11 passed, 0 failed, 3 with warnings (completed in 1340ms)"
- Indicates: ROCm works but has minor issues or informational warnings (usually normal on APUs)

**❌ Red Summary Examples:**
- "❌ ROCm tests failed - Tests: 11 total, 7 passed, 4 failed (completed in 890ms)"
- Indicates: Critical issues requiring immediate attention

#### 📋 Individual Test Result Format

Each test result is displayed as an expandable card showing:

**Test Header (Always Visible):**
- **Status Icon**: ✅ (Success), ❌ (Failure), ⚠️ (Warning)
- **Test Name**: Descriptive name (e.g., "ROCm Info - Success")
- **Warning Indicator**: Additional ⚠️ if issues detected but test passed
- **Execution Time**: Duration in milliseconds (e.g., "156ms")
- **Expand Arrow**: ▼ (collapsed) / ▲ (expanded)

**Expandable Details (Click to View):**
- **Command**: Exact command executed (e.g., `rocminfo` or `rocm-smi -a`)
- **Issues Detected**: Specific warnings or problems found (if any)
- **Output**: Complete raw output from the command
- **Error Output**: Error messages if command failed

#### 🔍 Interpreting Test Results

**✅ Success with No Issues**
```
✅ ROCm SMI Temperature - Success                    78ms ▼
```
- Command executed successfully
- No warnings or issues detected
- Output contains expected data

**✅ Success with Warnings**  
```
✅ ROCm SMI Detailed - Success ⚠️                   118ms ▼
Issues detected:
⚠️ Some metrics unavailable (temperature, power, etc.)
```
- Command executed successfully  
- Minor issues detected (often normal on APUs)
- Functionality still working correctly

**❌ Failure**
```
❌ HIP Version - Failed                              45ms ▼
Issues detected:
❌ Command not found - ROCm may not be installed
```
- Command failed to execute
- Critical issue requiring attention
- Specific guidance provided in issues section

#### 📊 Performance Indicators

**Execution Times:**
- **Fast (< 50ms)**: hipconfig commands, simple queries
- **Normal (50-150ms)**: rocm-smi commands, system queries  
- **Slower (> 150ms)**: rocminfo commands with full device enumeration

**Abnormal Timing Indicators:**
- **> 1000ms**: May indicate system performance issues
- **Timeout (30s)**: Critical system problems, driver issues
- **Variable timing**: Inconsistent performance, potential instability

#### 🎨 Visual Interface Elements

**Modal Layout:**
- **Header**: Test title with close button (×)
- **Summary Bar**: Color-coded overall result
- **Test List**: Expandable cards for each test
- **Scrollable Content**: Handle large output easily

**Color Coding:**
- **Green**: Success states, healthy systems
- **Yellow**: Warning states, minor issues  
- **Red**: Error states, critical issues
- **Blue**: Informational elements, neutral states

**Interactive Elements:**
- **Click Test Cards**: Expand/collapse detailed information
- **Hover Effects**: Visual feedback on interactive elements  
- **Responsive Design**: Works on desktop and mobile devices
- **Keyboard Support**: Accessible via keyboard navigation

### 📖 Dashboard Usage Guide

#### 🚀 Getting Started

1. **Launch Application**
   ```bash
   ./rocm-monitor
   ```

2. **Open Dashboard** 
   - Navigate to `http://localhost:8080`
   - Wait for "Connected" status indicator

3. **Verify System Health**
   - Click "🔧 Test ROCm" for comprehensive diagnostics
   - Review any warnings or issues

4. **Monitor Real-Time Data**
   - Observe temperature, power, and usage trends
   - Adjust time window and refresh interval as needed

#### ⚙️ Configuration Recommendations

**For Development/Testing:**
- **Interval**: 1-5 seconds for responsive monitoring
- **Time Window**: 5-15 minutes for recent trends
- **Export**: Regular JSON exports for analysis

**For Production Monitoring:**
- **Interval**: 10-30 seconds to reduce overhead
- **Time Window**: 30 minutes to 1 hour for trend analysis
- **Metrics**: Enable Prometheus endpoint (`-metrics` flag)

**For Troubleshooting:**
- **Interval**: 1 second for maximum responsiveness
- **Time Window**: 5 minutes for immediate issue correlation
- **Testing**: Run ROCm tests during problem periods

#### 📈 Monitoring Best Practices

**Temperature Monitoring:**
- **Normal Range**: 30-70°C for most workloads
- **Concerning**: >80°C sustained temperatures
- **Critical**: >90°C temperatures (thermal throttling likely)

**Power Monitoring:**
- **Baseline**: Note idle power consumption (typically 5-15W)
- **Load Testing**: Monitor power spikes during workloads
- **Efficiency**: Correlate power usage with performance metrics

**Memory Monitoring:**
- **Capacity Planning**: Keep VRAM usage <80% for performance
- **Leak Detection**: Watch for gradually increasing memory usage
- **Allocation Patterns**: Monitor usage spikes during operations

**Performance Correlation:**
- **GPU + CPU**: High GPU usage should correlate with application CPU usage
- **Clock Scaling**: Observe frequency changes under load
- **Thermal Throttling**: Watch for clock reduction when temperature rises

#### 🔄 Workflow Integration

**Development Workflow:**
1. Start monitoring before launching GPU workloads
2. Run ROCm tests after driver updates
3. Export data for performance regression testing
4. Use real-time monitoring during development

**DevOps Integration:**
1. Monitor during deployment and scaling
2. Set up automated ROCm testing in CI/CD
3. Export metrics to monitoring infrastructure
4. Create alerts for temperature/power thresholds

**Research/Analysis Workflow:**
1. Baseline system before experiments
2. Monitor continuously during long-running tasks
3. Export detailed data for analysis
4. Correlate performance with hardware metrics

### Common Issues and Solutions

#### 🔧 ROCm Installation Issues

**❌ "Command not found" Errors**
```
Issue: rocminfo: command not found
Solution: Install ROCm runtime
```
```bash
# Ubuntu 24.04 installation
sudo apt update
sudo apt install rocminfo rocm-smi-lib hip-runtime-dev
```

**❌ "No AMD/ROCm GPU devices detected"**
```
Issue: No compatible GPU found
Solutions:
1. Verify GPU compatibility: https://rocm.docs.amd.com/en/latest/release/gpu_os_support.html
2. Check if GPU is detected: lspci | grep AMD
3. Install GPU drivers: sudo apt install amdgpu-dkms
```

#### 🛡️ Permission Issues

**❌ "Permission denied - add user to render group"**
```
Issue: User lacks GPU access permissions
Solution: Add user to render group
```
```bash
sudo usermod -a -G render $USER
sudo usermod -a -G video $USER
# Log out and back in, or reboot
```

#### ⚙️ Driver and Runtime Issues

**⚠️ "ROCk module not loaded"**
```
Issue: ROCm kernel driver not loaded
Solutions:
1. Restart system after ROCm installation
2. Load modules manually: sudo modprobe amdgpu
3. Check module status: lsmod | grep amdgpu
```

**⚠️ "HSA runtime version not found"**
```
Issue: HSA runtime not properly installed
Solution: Install HSA runtime
```
```bash
sudo apt install hsa-rocr-dev hsa-runtime-dev
```

**⚠️ "No GPU device type detected"**
```
Issue: GPU not detected by HSA runtime
Solutions:
1. Check ROCm compatibility
2. Verify driver installation
3. Restart system
```

#### 💾 HIP Runtime Issues

**❌ "HIP not properly installed"**
```
Issue: HIP runtime missing or broken
Solution: Install/reinstall HIP
```
```bash
sudo apt install hip-runtime-dev hipcc
```

**⚠️ "Invalid HIP version output"** (Fixed in latest version)
```
Issue: Test was expecting wrong output format from hipconfig commands
Note: hipconfig --version returns "7.1.25424-4179531dcd" (version numbers)
      hipconfig --platform returns "amd" (platform name)
Solution: Update to latest version - this was a test detection bug
```

#### ⚡ Performance and Hardware Issues

**⚠️ "GPU is in a low-power state"** (Warning)
```
Issue: GPU running in power-saving mode
Note: This is normal when GPU is idle
Action: No action required unless performance is poor
```

**⚠️ "Some metrics unavailable"** (Warning)
```
Issue: Certain GPU sensors not supported
Note: Common on APUs and some GPU models
Action: Normal behavior, not a critical issue
```

**⚠️ "Likely driver error!" in ROCm SMI output** (Informational)
```
Issue: Empty clock frequency domains (dcefclk, mclk) 
Note: Normal on APU systems where certain clocks are managed differently
Examples: "Clock [mclk] on device [0] exists but EMPTY! Likely driver error!"
Action: No action required - this is expected behavior on APUs
```

**⚠️ "Failed to retrieve GPU metrics"** (Informational)
```
Issue: GPU metrics version not supported for this device
Note: Common when monitoring interface version doesn't match GPU generation
Action: No action required - basic monitoring still works
```

#### 🚨 Critical Errors

**❌ "Critical driver error detected"**
```
Issue: Driver malfunction or corruption
Solutions:
1. Restart system
2. Reinstall ROCm: sudo apt remove --purge rocm-* && sudo apt install rocm
3. Check system logs: dmesg | grep amdgpu
```

**❌ "Fatal error detected"**
```
Issue: System-level failure
Solutions:
1. Check system logs: journalctl -b | grep rocm
2. Verify hardware: memtest86+
3. Reinstall drivers and ROCm
```

### Diagnostic Tips

#### Pre-Test Checklist
1. **Verify Installation**
   ```bash
   which rocminfo rocm-smi hipconfig
   ```

2. **Check Groups**
   ```bash
   groups $USER  # Should include 'render' and 'video'
   ```

3. **Test Basic Commands**
   ```bash
   rocminfo | head -10
   rocm-smi
   ```

#### Advanced Troubleshooting
```bash
# Check ROCm installation status
dpkg -l | grep rocm

# Verify GPU detection at hardware level
lspci -nn | grep AMD

# Check kernel modules
lsmod | grep amdgpu

# Review system logs for errors
dmesg | grep -i rocm
journalctl -b | grep -i amd
```

### Test Automation

#### Scheduled Testing
```bash
# Run tests every hour via cron
0 * * * * curl -X POST http://localhost:8080/api/rocm-test > /var/log/rocm-test.log 2>&1
```

#### CI/CD Integration
```bash
#!/bin/bash
# ROCm validation script for CI/CD
response=$(curl -s -X POST http://localhost:8080/api/rocm-test)
success=$(echo "$response" | jq -r '.overall_success')

if [ "$success" = "true" ]; then
    echo "✅ ROCm tests passed"
    exit 0
else
    echo "❌ ROCm tests failed"
    echo "$response" | jq '.summary'
    exit 1
fi
```

## Architecture

The application is structured into modular components:

- **main.go** - HTTP server and route handlers
- **collector.go** - Data collection service with rocm-smi integration
- **rocm_data.go** - Data structures and parsing logic
- **exporter.go** - Export functionality (CSV, JSON, Prometheus)
- **test_rocm.go** - ROCm diagnostics and system testing
- **static/index.html** - Web dashboard

### Security Features

- Command execution timeout (3 seconds)
- Input validation for all parameters
- Configurable CORS origins
- Graceful error handling
- No shell injection vulnerabilities

### Performance Optimizations

- Pre-compiled regex patterns
- Efficient circular buffer for history
- Minimal memory allocations
- Concurrent-safe data access
- Optimized chart updates

## Troubleshooting

### rocm-smi not found
Ensure ROCm is properly installed and `rocm-smi` is in your PATH:
```bash
which rocm-smi
```

### Permission denied
The application needs permission to execute `rocm-smi`. Run with appropriate permissions or add your user to the `video` group:
```bash
sudo usermod -a -G video $USER
```

### No GPU data
Check if your GPU is detected by ROCm:
```bash
rocm-smi
```

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go best practices
- Tests are included for new features
- Documentation is updated
- Security considerations are addressed

## Observability

### Overview

ROCm Monitor provides comprehensive observability capabilities for GPU systems through multiple interfaces:

- **📊 Real-time Dashboard** - Interactive web interface with live charts
- **📈 Prometheus Metrics** - Enterprise-grade metrics export for Grafana
- **🔍 ROCm Diagnostics** - Comprehensive system testing and validation
- **📁 Data Export** - Historical data export in CSV/JSON formats
- **⚡ REST API** - Programmatic access to all monitoring data

### Monitoring Layers

#### 1. Hardware Layer
- **GPU Metrics**: Temperature, power, utilization, clock frequencies
- **Memory Metrics**: VRAM usage, capacity, utilization percentages
- **Thermal Metrics**: Temperature thresholds and thermal management
- **Performance Metrics**: Clock speeds, performance states, boost behavior

#### 2. System Layer
- **Host Metrics**: CPU utilization, system resource usage
- **Collection Health**: Data collection performance and reliability
- **Service Metrics**: Monitor uptime, memory usage, error tracking
- **Connectivity**: GPU detection and communication health

#### 3. Application Layer
- **ROCm Stack**: Runtime validation, driver status, platform detection
- **Diagnostic Results**: Test execution metrics, validation outcomes
- **Historical Analysis**: Trend analysis and capacity planning data
- **Alert Conditions**: Threshold monitoring and anomaly detection

### Metrics Export Formats

#### 1. Prometheus Metrics (Production Ready)
```bash
# Enable comprehensive metrics
./rocm-monitor -metrics

# Access metrics endpoint
curl http://localhost:8080/metrics
```

**21+ Production Metrics:**
- GPU hardware metrics with rich labels
- System health and performance indicators  
- Collection reliability and error tracking
- Performance threshold monitoring
- Multi-GPU support with device identification

#### 2. CSV Export (Data Analysis)
```bash
# Export historical data
curl http://localhost:8080/api/export.csv > gpu_data.csv
```

**Includes:**
- Timestamp-indexed data points
- All GPU metrics per collection interval
- System CPU utilization correlation
- Clock frequency tracking over time

#### 3. JSON Export (API Integration)
```bash
# Export with metadata
curl http://localhost:8080/api/export.json > gpu_data.json
```

**Structured Format:**
- Export metadata and statistics
- Complete historical dataset
- Collector performance metrics
- Data validation and integrity info

### Prometheus & Grafana Integration

#### Comprehensive Metrics Export

The enhanced `/metrics` endpoint provides 21+ metrics for complete observability:

**GPU Hardware Metrics:**
- `rocm_gpu_temperature_celsius` - GPU edge temperature in Celsius
- `rocm_gpu_power_watts` - GPU power consumption in watts  
- `rocm_gpu_usage_percent` - GPU compute utilization percentage
- `rocm_gpu_vram_usage_gb` / `rocm_gpu_vram_total_gb` - VRAM capacity metrics
- `rocm_gpu_vram_utilization_percent` - VRAM utilization percentage
- `rocm_gpu_sclk_mhz` / `rocm_gpu_mclk_mhz` - System and memory clock frequencies
- `rocm_gpu_fan_speed_percent` - GPU fan speed percentage

**System Metrics:**
- `rocm_system_cpu_usage_percent` - System CPU utilization
- `rocm_system_gpu_count` - Number of detected GPUs

**Monitoring Health Metrics:**
- `rocm_monitor_collection_errors_total` - Total collection errors (counter)
- `rocm_monitor_collection_duration_ms` - Collection time in milliseconds
- `rocm_monitor_data_points_total` - Total data points collected (counter)
- `rocm_monitor_uptime_seconds` - Monitor uptime in seconds
- `rocm_monitor_memory_usage_mb` - Monitor memory usage in MB
- `rocm_monitor_history_size_points` - Historical data points stored

**Performance Thresholds:**
- `rocm_gpu_temperature_warning_threshold` - Temperature warning (>70°C)
- `rocm_gpu_temperature_critical_threshold` - Temperature critical (>80°C)
- `rocm_gpu_vram_high_utilization` - VRAM high usage alert (>80%)

**ROCm Test Metrics** (when tests are executed):
- `rocm_test_suite_success` - Overall test suite success status
- `rocm_test_suite_duration_ms` - Total test execution time
- `rocm_test_suite_total_tests` / `rocm_test_suite_passed_tests` - Test counts
- `rocm_test_success` - Individual test success status
- `rocm_test_duration_ms` - Individual test execution times
- `rocm_test_issues_count` - Number of issues detected per test

**Rich Labels for Multi-GPU Support:**
All GPU metrics include comprehensive labels:
- `gpu_id` - GPU identifier (0, 1, 2...)
- `product_name` - GPU model name (AMD Radeon Graphics)
- `vendor` - GPU vendor (AMD, NVIDIA)
- `serial_number` - Hardware serial number
- `vram_vendor` - VRAM manufacturer

#### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'rocm-monitor'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
    scrape_timeout: 10s
    honor_labels: true
```

### Quick Setup with Docker Compose

For rapid deployment, use the included Docker Compose stack:

```bash
# Start ROCm Monitor with metrics
./rocm-monitor -metrics &

# Launch complete monitoring stack
cd container
./start.sh

# Access dashboards
# Grafana: http://localhost:3000 (admin/admin123)
# Prometheus: http://localhost:9090
```

**Included Components:**
- **Prometheus** - Metrics collection with 7-day retention
- **Grafana** - Pre-configured ROCm GPU dashboard
- **Alertmanager** - Production-ready alert rules
- **Custom Dashboard** - Temperature, power, VRAM, performance metrics

See `container/README.md` for detailed setup and configuration options.

#### Grafana Dashboard Queries

**GPU Performance Overview:**
```promql
# Temperature monitoring
rocm_gpu_temperature_celsius{gpu_id="0"}

# Power consumption trends
rocm_gpu_power_watts{gpu_id="0"}

# GPU utilization
rocm_gpu_usage_percent{gpu_id="0"}

# VRAM utilization
rocm_gpu_vram_utilization_percent{gpu_id="0"}
```

**Multi-GPU Monitoring:**
```promql
# All GPUs temperature
rocm_gpu_temperature_celsius{gpu_id=~".*"}

# Average GPU utilization across all devices
avg(rocm_gpu_usage_percent)

# Total system power consumption
sum(rocm_gpu_power_watts)
```

**System Health Dashboard:**
```promql
# Collection reliability
rate(rocm_monitor_collection_errors_total[5m])

# Monitor performance
rocm_monitor_collection_duration_ms

# System uptime
rocm_monitor_uptime_seconds

# Data collection rate
rate(rocm_monitor_data_points_total[5m])
```

**Performance Analysis:**
```promql
# Clock frequency trends
rocm_gpu_sclk_mhz{gpu_id="0"}
rocm_gpu_mclk_mhz{gpu_id="0"}

# Power efficiency (performance per watt)
rocm_gpu_usage_percent{gpu_id="0"} / rocm_gpu_power_watts{gpu_id="0"}

# Thermal efficiency (performance vs temperature)
rocm_gpu_usage_percent{gpu_id="0"} / rocm_gpu_temperature_celsius{gpu_id="0"}
```

#### Production Alert Rules

```yaml
groups:
  - name: rocm_critical_alerts
    rules:
      - alert: GPUTemperatureCritical
        expr: rocm_gpu_temperature_celsius > 85
        for: 2m
        labels:
          severity: critical
          component: hardware
        annotations:
          summary: "GPU {{ $labels.gpu_id }} temperature critically high"
          description: "GPU temperature {{ $value }}°C exceeds safe operating limits"
          
      - alert: VRAMExhaustion
        expr: rocm_gpu_vram_utilization_percent > 95
        for: 1m
        labels:
          severity: critical
          component: memory
        annotations:
          summary: "GPU {{ $labels.gpu_id }} VRAM near exhaustion"
          description: "VRAM utilization {{ $value }}% may cause out-of-memory errors"

  - name: rocm_warning_alerts
    rules:
      - alert: GPUTemperatureWarning
        expr: rocm_gpu_temperature_celsius > 75
        for: 5m
        labels:
          severity: warning
          component: hardware
        annotations:
          summary: "GPU {{ $labels.gpu_id }} running warm"
          description: "GPU temperature {{ $value }}°C approaching thermal limits"
          
      - alert: VRAMHighUtilization  
        expr: rocm_gpu_vram_utilization_percent > 80
        for: 3m
        labels:
          severity: warning
          component: memory
        annotations:
          summary: "GPU {{ $labels.gpu_id }} VRAM utilization high"
          description: "VRAM utilization {{ $value }}% may impact performance"
          
      - alert: ROCmCollectionErrors
        expr: increase(rocm_monitor_collection_errors_total[10m]) > 5
        for: 2m
        labels:
          severity: warning
          component: monitoring
        annotations:
          summary: "ROCm data collection experiencing errors"
          description: "{{ $value }} collection errors in the last 10 minutes"
          
      - alert: MonitorPerformanceDegraded
        expr: rocm_monitor_collection_duration_ms > 1000
        for: 5m
        labels:
          severity: warning
          component: monitoring
        annotations:
          summary: "ROCm monitor collection performance degraded"
          description: "Collection duration {{ $value }}ms exceeds normal range"

  - name: rocm_info_alerts
    rules:
      - alert: ROCmTestFailure
        expr: rocm_test_suite_success == 0
        for: 0s
        labels:
          severity: info
          component: diagnostics
        annotations:
          summary: "ROCm diagnostic tests failed"
          description: "ROCm system validation detected issues requiring attention"
```

### Observability Best Practices

#### 1. Monitoring Strategy

**Real-time Monitoring:**
- Use 15-second intervals for production workloads
- Monitor temperature and power consumption continuously
- Track VRAM utilization during compute-intensive tasks
- Set up immediate alerts for critical thresholds

**Historical Analysis:**
- Retain 24-48 hours of high-resolution data
- Export daily summaries for long-term trending
- Analyze performance patterns and capacity planning
- Correlate GPU metrics with application performance

**Health Monitoring:**
- Monitor collector reliability and performance
- Track ROCm stack health with periodic diagnostic tests
- Validate driver stability and runtime functionality
- Monitor for hardware degradation over time

#### 2. Dashboard Design

**Executive Dashboard:**
- System overview with key health indicators
- Multi-GPU summary with aggregate metrics
- Alert status and system uptime tracking
- Performance trending and capacity utilization

**Operations Dashboard:**
- Detailed GPU metrics with drill-down capability
- Real-time troubleshooting and diagnostic tools
- Collection health and monitoring system status
- Historical data analysis and export capabilities

**Engineering Dashboard:**
- Clock frequency analysis and performance tuning
- Thermal management and power efficiency metrics
- VRAM allocation patterns and optimization opportunities
- ROCm stack validation and compatibility tracking

#### 3. Alert Management

**Alert Hierarchy:**
1. **Critical** - Immediate response required (temperature >85°C, VRAM >95%)
2. **Warning** - Action needed soon (temperature >75°C, VRAM >80%)
3. **Info** - Awareness alerts (test failures, performance changes)

**Escalation Procedures:**
- Critical alerts: Immediate notification and automated response
- Warning alerts: Team notification within 15 minutes
- Info alerts: Daily digest and trend analysis

**Alert Fatigue Prevention:**
- Use proper thresholds based on workload characteristics
- Implement alert suppression during maintenance windows
- Group related alerts to avoid notification spam
- Regular review and tuning of alert sensitivity

## License

[Your License Here]

## Acknowledgments

- ROCm team for the GPU driver and tools
- Chart.js for the visualization library