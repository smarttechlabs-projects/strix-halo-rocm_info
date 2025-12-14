package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ROCmTestResult represents the result of a single ROCm test
type ROCmTestResult struct {
	Command     string    `json:"command"`
	Success     bool      `json:"success"`
	Output      string    `json:"output"`
	ErrorOutput string    `json:"error_output,omitempty"`
	Duration    int64     `json:"duration_ms"`
	Timestamp   time.Time `json:"timestamp"`
	Issues      []string  `json:"issues,omitempty"`
	Summary     string    `json:"summary"`
}

// ROCmTestSuite represents the complete test results
type ROCmTestSuite struct {
	OverallSuccess bool             `json:"overall_success"`
	TestResults    []ROCmTestResult `json:"test_results"`
	Summary        string           `json:"summary"`
	Timestamp      time.Time        `json:"timestamp"`
	Duration       int64            `json:"total_duration_ms"`
}

// ROCmTester handles running ROCm diagnostic tests
type ROCmTester struct {
	timeout time.Duration
}

// NewROCmTester creates a new ROCm tester
func NewROCmTester() *ROCmTester {
	return &ROCmTester{
		timeout: 30 * time.Second,
	}
}

// RunTests executes all ROCm diagnostic tests
func (rt *ROCmTester) RunTests() *ROCmTestSuite {
	suite := &ROCmTestSuite{
		Timestamp:      time.Now(),
		OverallSuccess: true,
		TestResults:    []ROCmTestResult{},
	}

	startTime := time.Now()

	// Define test commands
	tests := []struct {
		name        string
		command     string
		args        []string
		description string
	}{
		{
			name:        "ROCm Info",
			command:     "rocminfo",
			args:        []string{},
			description: "Basic ROCm system information",
		},
		{
			name:        "ROCm SMI",
			command:     "rocm-smi",
			args:        []string{},
			description: "ROCm System Management Interface",
		},
		{
			name:        "ROCm SMI Detailed",
			command:     "rocm-smi",
			args:        []string{"-a"},
			description: "ROCm SMI detailed information",
		},
		{
			name:        "ROCm SMI GPU List",
			command:     "rocm-smi",
			args:        []string{"-l"},
			description: "List available GPU devices",
		},
		{
			name:        "ROCm SMI Temperature",
			command:     "rocm-smi",
			args:        []string{"-t"},
			description: "GPU temperature information",
		},
		{
			name:        "ROCm SMI Power",
			command:     "rocm-smi",
			args:        []string{"-p"},
			description: "GPU power consumption",
		},
		{
			name:        "ROCm SMI Clock Frequencies",
			command:     "rocm-smi",
			args:        []string{"-c"},
			description: "GPU clock frequencies",
		},
		{
			name:        "ROCm SMI Memory Info",
			command:     "rocm-smi",
			args:        []string{"-u"},
			description: "GPU memory utilization",
		},
		{
			name:        "Hip Version",
			command:     "hipconfig",
			args:        []string{"--version"},
			description: "HIP runtime version",
		},
		{
			name:        "Hip Platform",
			command:     "hipconfig",
			args:        []string{"--platform"},
			description: "HIP platform information",
		},
		{
			name:        "Device Query",
			command:     "rocminfo",
			args:        []string{},
			description: "Detailed device information",
		},
	}

	// Run each test
	for _, test := range tests {
		result := rt.runSingleTest(test.name, test.command, test.args, test.description)
		suite.TestResults = append(suite.TestResults, result)
		
		if !result.Success {
			suite.OverallSuccess = false
		}
	}

	suite.Duration = time.Since(startTime).Milliseconds()
	suite.Summary = rt.generateSummary(suite)

	return suite
}

// runSingleTest executes a single test command
func (rt *ROCmTester) runSingleTest(name, command string, args []string, description string) ROCmTestResult {
	result := ROCmTestResult{
		Command:   fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
		Timestamp: time.Now(),
		Issues:    []string{},
	}

	startTime := time.Now()
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), rt.timeout)
	defer cancel()

	// Execute command
	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()
	
	result.Duration = time.Since(startTime).Milliseconds()
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.ErrorOutput = err.Error()
		
		// Check for specific error types
		if ctx.Err() == context.DeadlineExceeded {
			result.Issues = append(result.Issues, "Command timeout - may indicate system issues")
			result.Summary = fmt.Sprintf("❌ %s - Timeout after %d seconds", name, int(rt.timeout.Seconds()))
		} else if strings.Contains(err.Error(), "no such file") {
			result.Issues = append(result.Issues, fmt.Sprintf("Command '%s' not found - ROCm may not be installed", command))
			result.Summary = fmt.Sprintf("❌ %s - Command not found", name)
		} else {
			result.Issues = append(result.Issues, fmt.Sprintf("Command failed: %s", err.Error()))
			result.Summary = fmt.Sprintf("❌ %s - Failed", name)
		}
	} else {
		result.Success = true
		result.Summary = fmt.Sprintf("✅ %s - Success", name)
		
		// Analyze output for potential issues
		issues := rt.analyzeOutput(command, result.Output)
		result.Issues = append(result.Issues, issues...)
	}

	return result
}

// analyzeOutput checks command output for known issues and patterns
func (rt *ROCmTester) analyzeOutput(command, output string) []string {
	issues := []string{}
	outputLower := strings.ToLower(output)

	switch command {
	case "rocminfo":
		// Check for HSA agents section
		if !strings.Contains(outputLower, "hsa agents") {
			issues = append(issues, "No HSA Agents section found")
		}
		
		// Check for at least one agent
		if !strings.Contains(outputLower, "agent 1") && !strings.Contains(outputLower, "agent:") {
			issues = append(issues, "No agents detected")
		}
		
		// Check for GPU device type
		if !strings.Contains(outputLower, "device type:             gpu") {
			issues = append(issues, "No GPU device type detected")
		}
		
		// Check for HSA runtime version
		if !strings.Contains(outputLower, "runtime version:") {
			issues = append(issues, "HSA runtime version not found")
		}
		
		// Check for ROCk module
		if !strings.Contains(outputLower, "rock module is loaded") {
			issues = append(issues, "ROCk module not loaded")
		}

	case "rocm-smi":
		// Check for GPU devices - more flexible detection
		hasDevice := strings.Contains(outputLower, "device") && 
		           (strings.Contains(outputLower, "gpu[") || strings.Contains(outputLower, "radeon") || strings.Contains(outputLower, "amd"))
		
		if !hasDevice {
			issues = append(issues, "No AMD/ROCm GPU devices detected")
		}
		
		// Only flag critical errors, not normal "Not supported" messages
		if strings.Contains(outputLower, "permission denied") {
			issues = append(issues, "Permission denied - add user to render group: sudo usermod -a -G render $USER")
		}
		
		if strings.Contains(outputLower, "no devices found") {
			issues = append(issues, "No ROCm devices found - check ROCm installation")
		}
		
		// Check for critical failures only - exclude known informational messages
		if strings.Contains(outputLower, "failed to initialize") && !strings.Contains(outputLower, "gpu metrics") {
			issues = append(issues, "Critical initialization failure detected")
		}
		
		// Only flag truly critical driver issues, not informational messages about unsupported features
		if strings.Contains(outputLower, "driver") && strings.Contains(outputLower, "crashed") {
			issues = append(issues, "GPU driver crash detected")
		}
		
		// Note: "Likely driver error!" messages for empty clock frequencies are normal on APUs
		// Note: "Failed to retrieve GPU metrics" is normal when metric version not supported
		// Note: "Not supported on the given system" is normal for many features on APUs
		// These are informational, not errors requiring action

	case "hipconfig":
		// Check HIP installation
		if strings.Contains(outputLower, "not found") {
			issues = append(issues, "HIP not properly installed")
		}
		
		// Check for valid output based on the specific hipconfig command
		if strings.Contains(command, "--version") {
			// For --version, expect version number format
			if !regexp.MustCompile(`\d+\.\d+`).MatchString(output) {
				issues = append(issues, "Invalid HIP version output - no version number found")
			}
		} else if strings.Contains(command, "--platform") {
			// For --platform, expect platform name (amd, nvidia, etc.)
			if !regexp.MustCompile(`(?i)(amd|nvidia|rocm|cuda)`).MatchString(output) {
				issues = append(issues, "Invalid HIP platform output - no platform detected")
			}
		}
	}

	// Only check for truly critical error patterns, exclude known false positives
	criticalPatterns := []struct {
		pattern   string
		message   string
		excludes  []string // Exclude if any of these patterns are also present
	}{
		{"fatal error", "Fatal error detected", []string{}},
		{"segmentation fault", "Segmentation fault detected", []string{}},
		{"permission denied", "Permission denied - check user groups", []string{}},
		{"command not found", "Command not found - ROCm may not be installed", []string{}},
		{"timeout", "Timeout errors detected", []string{}},
		// Only flag driver errors that are truly critical, not informational messages
		{"driver initialization failed", "Critical driver initialization failure", []string{}},
		{"driver load failed", "Critical driver load failure", []string{}},
		{"cannot access device", "Device access error - check permissions or drivers", []string{}},
	}

	for _, ep := range criticalPatterns {
		if strings.Contains(outputLower, ep.pattern) {
			// Check if any exclude patterns are present
			shouldExclude := false
			for _, exclude := range ep.excludes {
				if exclude != "" && strings.Contains(outputLower, exclude) {
					shouldExclude = true
					break
				}
			}
			
			if !shouldExclude {
				issues = append(issues, ep.message)
			}
		}
	}

	return issues
}

// generateSummary creates an overall summary of the test results
func (rt *ROCmTester) generateSummary(suite *ROCmTestSuite) string {
	passed := 0
	failed := 0
	warnings := 0

	for _, result := range suite.TestResults {
		if result.Success {
			passed++
			if len(result.Issues) > 0 {
				warnings++
			}
		} else {
			failed++
		}
	}

	total := len(suite.TestResults)
	
	summary := fmt.Sprintf("Tests: %d total, %d passed, %d failed", total, passed, failed)
	
	if warnings > 0 {
		summary += fmt.Sprintf(", %d with warnings", warnings)
	}

	if suite.OverallSuccess {
		if warnings > 0 {
			summary = "⚠️  ROCm tests passed with warnings - " + summary
		} else {
			summary = "✅ All ROCm tests passed - " + summary
		}
	} else {
		summary = "❌ ROCm tests failed - " + summary
	}

	// Add duration
	summary += fmt.Sprintf(" (completed in %dms)", suite.Duration)

	return summary
}

// rocmTestHandler handles the /api/rocm-test endpoint
func rocmTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tester := NewROCmTester()
	results := tester.RunTests()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode test results", http.StatusInternalServerError)
		return
	}
}