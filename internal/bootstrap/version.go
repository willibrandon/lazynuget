package bootstrap

import "fmt"

// ShowVersion formats and prints version information.
func ShowVersion(version VersionInfo) {
	fmt.Println(version.String())
}
