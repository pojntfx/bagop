package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"github.com/alessio/shellescape"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

const (
	goCmd = "go"
)

var (
	listSupportedTargets = []string{"tool", "dist", "list", "--json"}
)

type Platform struct {
	GoOS   string `json:"GOOS"`
	GoArch string `json:"GOARCH"`
}

func getBinIdentifier(goOS, goArch string) string {
	if goOS == "windows" {
		return ".exe"
	}

	if goOS == "js" && goArch == "wasm" {
		return ".wasm"
	}

	return ""
}

func getArchIdentifier(goArch string) string {
	switch goArch {
	case "386":
		return "i686"
	case "amd64":
		return "x86_64"
	case "arm":
		return "armv7l" // Best effort, could also be `armv6l` etc. depending on `GOARCH`
	case "arm64":
		return "aarch64"
	default:
		return goArch
	}
}

func getSystemShell() []string {
	// For Windows, use PowerShell
	if runtime.GOOS == "windows" {
		return []string{"powershell", "-command"}
	}

	// Prefer Bash
	bash, err := exec.LookPath("bash")
	if err == nil {
		return []string{bash, "-c"}
	}

	// Fall back to POSIX shell
	return []string{"sh", "-c"}
}

func main() {
	// Define usage
	pflag.Usage = func() {
		fmt.Printf(`Build for all Go-supported platforms by default, disable those which you don't want.

Example usage: %s -b mybin -x '(android/arm$|ios/*|openbsd/mips64)' -j $(nproc) 'main.go'
Example usage (with plain flag): %s -b mybin -x '(android/arm$|ios/*|openbsd/mips64)' -j $(nproc) -p 'go build -o $DST main.go'

See https://github.com/pojntfx/bagop for more information.

Usage: %s [OPTION...] '<INPUT>'
`, os.Args[0], os.Args[0], os.Args[0])

		pflag.PrintDefaults()
	}

	// Parse flags
	binFlag := pflag.StringP("bin", "b", "mybin", "Prefix of resulting binary")
	distFlag := pflag.StringP("dist", "d", "out", "Directory build into")
	excludeFlag := pflag.StringP("exclude", "x", "", "Regex of platforms not to build for, i.e. (windows/386|linux/mips64)")
	extraArgs := pflag.StringP("extra-args", "e", "", "Extra arguments to pass to the Go compiler")
	jobsFlag := pflag.Int64P("jobs", "j", 1, "Maximum amount of parallel jobs")
	goismsFlag := pflag.BoolP("goisms", "g", false, "Use Go's conventions (i.e. amd64) instead of uname's conventions (i.e. x86_64)")
	plainFlag := pflag.BoolP("plain", "p", false, "Sets GOARCH, GOARCH and DST and leaves the rest up to you (see example usage)")

	pflag.Parse()

	// Validate arguments
	if pflag.NArg() == 0 {
		help := `command needs an argument: 'INPUT'`

		fmt.Println(help)

		pflag.Usage()

		fmt.Println(help)

		os.Exit(2)
	}

	// Interpret arguments
	input := pflag.Args()[0]

	// Construct platform query command
	queryCmd := exec.Command(goCmd, listSupportedTargets...)

	// Capture stdout and stderr
	var queryStdout, queryStderr bytes.Buffer
	queryCmd.Stdout = &queryStdout
	queryCmd.Stderr = &queryStderr

	// Get supported platforms
	if err := queryCmd.Run(); err != nil {
		log.Fatalf("could not query supported platforms: err=%v, stdout=%v, stderr=%v", err, queryStdout.String(), queryStderr.String())
	}

	// Parse platforms
	parsedPlatforms := []Platform{}
	if err := json.Unmarshal(queryStdout.Bytes(), &parsedPlatforms); err != nil { // It is safe to read outb.Bytes() here as we would only read it in `cmd.Run()` above if we were to error, in which case we would exit first
		log.Fatal("could not parse supported platforms:", err)
	}

	// Limits the max. amount of concurrent builds
	// See https://play.golang.org/p/othihEtsOBZ
	var wg = sync.WaitGroup{}
	guard := make(chan struct{}, *jobsFlag)

	for _, lplatform := range parsedPlatforms {
		guard <- struct{}{}
		wg.Add(1)

		go func(platform Platform) {
			defer func() {
				wg.Done()

				<-guard
			}()

			// Construct the filename
			output := filepath.Join(*distFlag, *binFlag+"."+platform.GoOS+"-")

			// Add the arch identifier
			archIdentifier := getArchIdentifier(platform.GoArch)
			if *goismsFlag {
				archIdentifier = platform.GoArch
			}
			output += archIdentifier

			// Add the binary identifier
			output += getBinIdentifier(platform.GoOS, platform.GoArch)

			// Check if current platform should be skipped
			skip, err := regexp.MatchString(*excludeFlag, platform.GoOS+"/"+platform.GoArch)
			if err != nil {
				log.Fatal("could not match check if platform should be blocked based on regex:", err)
			}

			// Skip the platform if it matches the exclude regex
			if skip {
				log.Printf("%v %v/%v (platform matched the provided regex)", color.New(color.FgYellow).SprintFunc()("skipping"), color.New(color.FgCyan).SprintFunc()(platform.GoOS), color.New(color.FgMagenta).SprintFunc()(platform.GoArch))

				return
			}

			// Continue if platform is enabled
			log.Printf("%v %v/%v (%v)", color.New(color.FgGreen).SprintFunc()("building"), color.New(color.FgCyan).SprintFunc()(platform.GoOS), color.New(color.FgMagenta).SprintFunc()(platform.GoArch), output)

			// Construct build args
			buildArgs := append([]string{"build", "-o"}, []string{shellescape.Quote(output)}...)

			// Add the extra args if they are defined
			if *extraArgs != "" {
				buildArgs = append(buildArgs, *extraArgs)
			}

			// Add the input
			buildArgs = append(buildArgs, shellescape.Quote(input))

			// Construct build command
			buildCmd := exec.Command(goCmd, buildArgs...)

			// If the plain flag is set, use the custom command
			if *plainFlag {
				buildCmd = exec.Command(getSystemShell()[0], []string{getSystemShell()[1], input}...)
			}

			// Set `GOOS` and `GOARCH` env vars
			buildCmd.Env = os.Environ()
			buildCmd.Env = append(buildCmd.Env, "GOOS="+shellescape.Quote(platform.GoOS), "GOARCH="+shellescape.Quote(platform.GoArch))

			// If the plain flag is set, also set DST
			if *plainFlag {
				buildCmd.Env = append(buildCmd.Env, "DST="+shellescape.Quote(output))
			}

			// Capture stdout and stderr
			var buildStdout, buildStderr bytes.Buffer
			buildCmd.Stdout = &buildStdout
			buildCmd.Stderr = &buildStderr

			// Start the build
			if err := buildCmd.Run(); err != nil {
				log.Fatalf("could not build for platform %v/%v: err=%v, stdout=%v, stderr=%v", platform.GoOS, platform.GoArch, err, buildStdout.String(), buildStderr.String())
			}

		}(lplatform)
	}

	wg.Wait()
}
