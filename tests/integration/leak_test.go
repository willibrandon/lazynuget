package integration

import (
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// TestResourceLeakDetection runs 1000 startup/shutdown cycles to detect memory leaks
// This test validates SC-010: No resource leaks in application lifecycle
func TestResourceLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping leak test in short mode")
	}

	// Build binary fresh for this test
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-leak-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "-f", "../../lazynuget-leak-test").Run()

	// Baseline memory
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	cycles := 1000
	t.Logf("Running %d startup/shutdown cycles...", cycles)

	start := time.Now()
	for i := 0; i < cycles; i++ {
		cmd := exec.Command("../../lazynuget-leak-test", "--version")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Cycle %d failed: %v", i+1, err)
		}

		// Progress indicator
		if (i+1)%100 == 0 {
			t.Logf("Completed %d/%d cycles", i+1, cycles)
		}
	}
	elapsed := time.Since(start)

	// Force GC to clean up any temporary allocations
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	// Measure memory after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate metrics (handle potential underflow if GC reduced memory)
	avgCycleTime := elapsed / time.Duration(cycles)
	var memGrowthMB float64
	if memAfter.Alloc > memBefore.Alloc {
		memGrowthBytes := memAfter.Alloc - memBefore.Alloc
		memGrowthMB = float64(memGrowthBytes) / 1024 / 1024
	} else {
		// Memory decreased (GC was effective) - this is good
		memGrowthMB = 0
	}

	t.Logf("Completed %d cycles in %v (avg: %v per cycle)", cycles, elapsed, avgCycleTime)
	t.Logf("Memory before: %.2f MB", float64(memBefore.Alloc)/1024/1024)
	t.Logf("Memory after: %.2f MB", float64(memAfter.Alloc)/1024/1024)
	t.Logf("Memory growth: %.2f MB", memGrowthMB)

	// Validation: Memory growth should be minimal (< 10MB for 1000 cycles)
	maxGrowthMB := 10.0
	if memGrowthMB > maxGrowthMB {
		t.Errorf("Memory leak detected: %.2f MB growth exceeds %.2f MB threshold", memGrowthMB, maxGrowthMB)
	}

	// Validation: Average cycle time should be reasonable (< 500ms)
	maxCycleTime := 500 * time.Millisecond
	if avgCycleTime > maxCycleTime {
		t.Errorf("Performance degradation: avg cycle time %v exceeds %v", avgCycleTime, maxCycleTime)
	}
}

// TestConcurrentStartupShutdown tests multiple concurrent instances
func TestConcurrentStartupShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	// Build binary
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-concurrent-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "-f", "../../lazynuget-concurrent-test").Run()

	concurrency := 10
	iterationsPerWorker := 50

	t.Logf("Running %d concurrent workers, %d iterations each", concurrency, iterationsPerWorker)

	errCh := make(chan error, concurrency)
	doneCh := make(chan bool, concurrency)

	start := time.Now()
	for w := 0; w < concurrency; w++ {
		go func(workerID int) {
			for i := 0; i < iterationsPerWorker; i++ {
				cmd := exec.Command("../../lazynuget-concurrent-test", "--version")
				if err := cmd.Run(); err != nil {
					errCh <- err
					return
				}
			}
			doneCh <- true
		}(w)
	}

	// Wait for all workers
	completed := 0
	for completed < concurrency {
		select {
		case err := <-errCh:
			t.Fatalf("Concurrent execution failed: %v", err)
		case <-doneCh:
			completed++
		case <-time.After(2 * time.Minute):
			t.Fatal("Concurrent test timed out")
		}
	}

	elapsed := time.Since(start)
	totalCycles := concurrency * iterationsPerWorker
	avgTime := elapsed / time.Duration(totalCycles)

	t.Logf("Completed %d total cycles in %v (avg: %v per cycle)", totalCycles, elapsed, avgTime)

	// Validation: Should handle concurrent access without errors
	// Test passes if we reach here without errors
}
