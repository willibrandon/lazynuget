package platform

import (
	"fmt"
	"os/exec"
)

// ValidateDotnetCLI checks if the dotnet CLI is available and functional.
// Returns an error with helpful installation instructions if dotnet is not found or not working.
func ValidateDotnetCLI() error {
	// Try to find dotnet in PATH
	dotnetPath, err := exec.LookPath("dotnet")
	if err != nil {
		return fmt.Errorf("dotnet CLI not found in PATH\n\n" +
			"LazyNuGet requires the .NET SDK to manage NuGet packages.\n\n" +
			"Installation instructions:\n" +
			"  • Windows: https://dotnet.microsoft.com/download\n" +
			"  • macOS: brew install dotnet-sdk\n" +
			"  • Linux: https://docs.microsoft.com/dotnet/core/install/linux")
	}

	// Verify dotnet works by running --version
	// #nosec G204 - dotnetPath comes from exec.LookPath which validates it's a real executable in PATH
	cmd := exec.Command(dotnetPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dotnet CLI found at %s but failed to execute: %w\n"+
			"Output: %s\n\n"+
			"Try reinstalling the .NET SDK", dotnetPath, err, string(output))
	}

	// Success - dotnet is available and working
	return nil
}
