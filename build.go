package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var version = "1.1.1"
var outputFileName = "qws"

var validOS = map[string]bool{
	"linux":     true, //(Linux)
	"darwin":    true, //(macOS)
	"windows":   true, //(Windows)
	"freebsd":   true, //(FreeBSD)
	"netbsd":    true, //(NetBSD)
	"openbsd":   true, //(OpenBSD)
	"dragonfly": true, //(DragonFly BSD)
	"plan9":     true, //(Plan 9)
	"solaris":   true, //(Oracle Solaris)
	"aix":       true, //(IBM AIX)
}

var validArch = map[string]bool{
	"amd64":    true, //(x86-64)
	"386":      true, //(x86 32-bit)
	"arm":      true, //(ARM)
	"arm64":    true, //(ARM 64-bit)
	"ppc64":    true, //(PowerPC 64-bit)
	"ppc64le":  true, //(PowerPC 64-bit Little-Endian)
	"mips":     true, //(MIPS 32-bit)
	"mipsle":   true, //(MIPS 32-bit Little-Endian)
	"mips64":   true, //(MIPS 64-bit)
	"mips64le": true, //(MIPS 64-bit Little-Endian)
	"s390x":    true, //(IBM Z Systems)
	// Add more valid arch values here
}

var commonCombination = []struct {
	os   string
	arch string
}{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"linux", "arm"},
	{"darwin", "amd64"},
	{"windows", "amd64"},
	{"windows", "386"},
	{"freebsd", "amd64"},
	{"freebsd", "arm64"},
	{"freebsd", "arm"},
	{"netbsd", "amd64"},
	{"netbsd", "arm64"},
	{"netbsd", "arm"},
	{"openbsd", "amd64"},
	{"openbsd", "arm64"},
	{"openbsd", "arm"},
}

func main() {
	var targetOS string
	var targetArch string
	var commonFlag bool

	flag.BoolVar(&commonFlag, "all", false, "Build for all popular OS and arch combinations")
	flag.StringVar(&targetOS, "os", runtime.GOOS, "Target operating system")
	flag.StringVar(&targetArch, "arch", runtime.GOARCH, "Target architecture")
	flag.Parse()

	// Check if the provided OS and arch are valid
	if !validOS[targetOS] {
		fmt.Printf("Invalid target OS: %s\n", targetOS)
		return
	}
	if !validArch[targetArch] {
		fmt.Printf("Invalid target architecture: %s\n", targetArch)
		return
	}
	if commonFlag {
		for _, comb := range commonCombination {
			buildTarget(comb.os, comb.arch)
		}
		return
	} else {
		buildTarget(targetOS, targetArch)
	}
}

func buildTarget(targetOS string, targetArch string) {

	fmt.Printf("Building for %s %s...\n", targetOS, targetArch)

	os.Setenv("GOOS", targetOS)
	os.Setenv("GOARCH", targetArch)

	outputName := "./dist/" + version + "/" + outputFileName + "_" + version + "_" + targetOS + "_" + targetArch
	if targetOS == "windows" {
		outputName += ".exe"
	}

	buildCmd := exec.Command("go", "build", "-o", outputName, "main.go")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		fmt.Println("Build error:", err)
		return
	}

	fmt.Println("Build successful!")
}
