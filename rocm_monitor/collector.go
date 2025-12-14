package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

// Collector manages the data collection process
type Collector struct {
	parser        *Parser
	dataMutex     sync.RWMutex
	history       []RocmData
	maxHistory    int
	interval      time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	errorCallback func(error)
}

// CollectorConfig holds configuration for the collector
type CollectorConfig struct {
	MaxHistory    int
	Interval      time.Duration
	ErrorCallback func(error)
}

// NewCollector creates a new collector instance
func NewCollector(config CollectorConfig) *Collector {
	if config.MaxHistory <= 0 {
		config.MaxHistory = 1000
	}
	if config.Interval <= 0 {
		config.Interval = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	return &Collector{
		parser:        NewParser(),
		history:       make([]RocmData, 0, config.MaxHistory),
		maxHistory:    config.MaxHistory,
		interval:      config.Interval,
		ctx:           ctx,
		cancel:        cancel,
		errorCallback: config.ErrorCallback,
	}
}

// Start begins the collection process
func (c *Collector) Start() {
	go c.collectLoop()
}

// Stop halts the collection process
func (c *Collector) Stop() {
	c.cancel()
}

// collectLoop runs the collection process
func (c *Collector) collectLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect initial data
	c.collect()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

// collect executes rocm-smi and stores the data
func (c *Collector) collect() {
	// Create context with timeout for command execution
	ctx, cancel := context.WithTimeout(c.ctx, 3*time.Second)
	defer cancel()

	// Execute rocm-smi with timeout protection
	cmd := exec.CommandContext(ctx, "rocm-smi")
	output, err := cmd.Output()
	
	if err != nil {
		if c.errorCallback != nil {
			c.errorCallback(fmt.Errorf("rocm-smi execution failed: %w", err))
		}
		return
	}

	// Also get detailed VRAM information
	cmdVRAM := exec.CommandContext(ctx, "rocm-smi", "--showmeminfo", "vram")
	vramOutput, vramErr := cmdVRAM.Output()
	
	// Get clock frequencies
	cmdClock := exec.CommandContext(ctx, "rocm-smi", "-c")
	clockOutput, clockErr := cmdClock.Output()
	
	// Combine outputs for parsing
	combinedOutput := string(output)
	if vramErr == nil {
		combinedOutput += "\n" + string(vramOutput)
	}
	if clockErr == nil {
		combinedOutput += "\n" + string(clockOutput)
	}
	// Parse the combined output
	data, err := c.parser.ParseRocmSMIOutput(combinedOutput)
	if err != nil {
		if c.errorCallback != nil {
			c.errorCallback(fmt.Errorf("parsing failed: %w", err))
		}
		return
	}

	// Get CPU usage
	cpuUsage, err := GetCPUUsage()
	if err != nil {
		if c.errorCallback != nil {
			c.errorCallback(fmt.Errorf("CPU usage collection failed: %w", err))
		}
		// Continue without CPU data
		cpuUsage = 0
	}
	data.CPUUsage = cpuUsage

	// Validate the data
	if err := data.Validate(); err != nil {
		if c.errorCallback != nil {
			c.errorCallback(fmt.Errorf("validation failed: %w", err))
		}
		return
	}

	// Store the data
	c.dataMutex.Lock()
	c.history = append(c.history, *data)
	
	// Maintain history size limit
	if len(c.history) > c.maxHistory {
		// Keep only the most recent data
		c.history = c.history[len(c.history)-c.maxHistory:]
	}
	c.dataMutex.Unlock()

	log.Printf("Collected data for %d GPUs at %s (SCLK: %.0f, MCLK: %.0f)", len(data.GPUs), data.Timestamp.Format(time.RFC3339), 
		func() float64 { if len(data.GPUs) > 0 { return data.GPUs[0].SCLKFreq } else { return 0 } }(),
		func() float64 { if len(data.GPUs) > 0 { return data.GPUs[0].MCLKFreq } else { return 0 } }())
}

// GetHistory returns a copy of the collected data history
func (c *Collector) GetHistory() []RocmData {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()
	
	// Return a copy to prevent external modification
	historyCopy := make([]RocmData, len(c.history))
	copy(historyCopy, c.history)
	
	return historyCopy
}

// GetLatest returns the most recent data point
func (c *Collector) GetLatest() (*RocmData, error) {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()
	
	if len(c.history) == 0 {
		return nil, fmt.Errorf("no data available")
	}
	
	latest := c.history[len(c.history)-1]
	return &latest, nil
}

// SetInterval updates the collection interval
func (c *Collector) SetInterval(interval time.Duration) {
	if interval <= 0 {
		return
	}
	
	// Stop current collection
	c.Stop()
	
	// Update interval
	c.interval = interval
	
	// Create new context and restart
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.Start()
}

// ClearHistory removes all collected data
func (c *Collector) ClearHistory() {
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()
	
	c.history = c.history[:0]
}

// GetStats returns statistics about the collected data
func (c *Collector) GetStats() map[string]interface{} {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()
	
	stats := make(map[string]interface{})
	stats["history_size"] = len(c.history)
	stats["max_history"] = c.maxHistory
	stats["interval_seconds"] = c.interval.Seconds()
	
	if len(c.history) > 0 {
		stats["oldest_timestamp"] = c.history[0].Timestamp
		stats["newest_timestamp"] = c.history[len(c.history)-1].Timestamp
		
		// Calculate average values across all GPUs and time
		var totalTemp, totalPower, totalGPU, totalVRAM float64
		var count int
		
		for _, data := range c.history {
			for _, gpu := range data.GPUs {
				totalTemp += gpu.Temperature
				totalPower += gpu.Power
				totalGPU += gpu.GPUUsage
				totalVRAM += gpu.VRAMUsage
				count++
			}
		}
		
		if count > 0 {
			stats["avg_temperature"] = totalTemp / float64(count)
			stats["avg_power"] = totalPower / float64(count)
			stats["avg_gpu_usage"] = totalGPU / float64(count)
			stats["avg_vram_usage"] = totalVRAM / float64(count)
		}
	}
	
	return stats
}