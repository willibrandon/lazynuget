package platform

import (
	"fmt"
	"strings"
)

// ValidateDotnetCLI checks if the dotnet CLI is available and functional.
// Returns an error with helpful installation instructions if dotnet is not found or not working.
// See: T091, FR-031
func ValidateDotnetCLI() error {
	spawner := NewProcessSpawner()

	// Try to run dotnet --version
	result, err := spawner.Run("dotnet", []string{"--version"}, "", nil)
	if err != nil {
		// dotnet not found or failed to execute
		return fmt.Errorf("dotnet CLI not found in PATH\n\n"+
			"LazyNuGet requires the .NET SDK to manage NuGet packages.\n\n"+
			"Installation instructions:\n"+
			"  • Windows: https://dotnet.microsoft.com/download\n"+
			"  • macOS: brew install dotnet-sdk\n"+
			"  • Linux: https://docs.microsoft.com/dotnet/core/install/linux\n\n"+
			"Error: %w", err)
	}

	// Check exit code
	if result.ExitCode != 0 {
		return fmt.Errorf("dotnet CLI failed to execute (exit code %d)\n"+
			"Stdout: %s\n"+
			"Stderr: %s\n\n"+
			"Try reinstalling the .NET SDK",
			result.ExitCode,
			result.Stdout,
			result.Stderr)
	}

	// Verify we got a version string
	version := strings.TrimSpace(result.Stdout)
	if version == "" {
		return fmt.Errorf("dotnet CLI returned empty version\n\n" +
			"Try reinstalling the .NET SDK")
	}

	// Success - dotnet is available and working
	return nil
}
