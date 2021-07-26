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
	"strings"
	"sync"

	"github.com/alessio/shellescape"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

const (
	goCmd = "go"
)

var (
	goToolDistListArgs = []string{"tool", "dist", "list", "--json"}
)

type Platform struct {
	GoOS   string `json:"GOOS"`
	GoArch string `json:"GOARCH"`
}

func getSuffixForGoOSAndGOArch(goOS, goArch string) string {
	if goOS == "windows" {
		return ".exe"
	}

	if goOS == "js" && goArch == "wasm" {
		return ".wasm"
	}

	return ""
}

func getUnameConventionForGoArch(goArch string) string {
	switch goArch {
	case "386":
		return "i686"
	case "amd64":
		return "x86_64"
	case "arm":
		return "armv7l" // Best guess, could also be `armv6l` etc. depending on `GOARCH`
	case "arm64":
		return "aarch64"
	default:
		return goArch
	}
}

func main() {
	// Define usage
	pflag.Usage = func() {
		fmt.Printf(`Build for all Go-supported platforms by default, disable those which you don't want.

Example usage: %s -n myexample -j $(nproc) -d out -b '(android/arm$|ios/*|openbsd/mips64)' main.go

See https://github.com/pojntfx/bagop for more information.

Usage: %s [OPTION...] <INPUT>
`, os.Args[0], os.Args[0])

		pflag.PrintDefaults()
	}

	// Parse flags
	nameFlag := pflag.StringP("name", "n", "example", "Prefix of resulting binary")
	distFlag := pflag.StringP("dist", "d", "out", "Directory to build to")
	blockFlag := pflag.StringP("block", "b", "", "Regex of platforms not to build for, i.e. (windows/386|linux/mips64)")
	extraArgs := pflag.StringP("extraArgs", "e", "", "Extra arguments to pass to the Go compiler")
	jobFlag := pflag.Int64P("jobs", "j", 1, "Maximum amount of parallel jobs")
	goarchConventions := pflag.BoolP("goarchConventions", "g", false, "Use GOARCH conventions in resulting file names (i.e. amd64). If not specified, uname's conventions (i.e. x86_64) will be used.")

	pflag.Parse()

	// Validate arguments
	if pflag.NArg() == 0 {
		help := `command needs an argument: INPUT`

		fmt.Println(help)

		pflag.Usage()

		fmt.Println(help)

		os.Exit(2)
	}

	// Interpret arguments
	input := strings.Join(pflag.Args(), " ")

	// Construct the platform query command
	cmd := exec.Command(goCmd, goToolDistListArgs...)

	// Capture stdout and stderr
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	// Get supported platforms
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not query supported platforms: err=%v, stdout=%v, stderr=%v", err, outb.String(), errb.String())
	}

	parsedPlatforms := []Platform{}
	if err := json.Unmarshal(outb.Bytes(), &parsedPlatforms); err != nil { // It is safe to read outb.Bytes() here as we would only read it in `cmd.Run()` above if we were to error, in which case we would exit first
		log.Fatal("Could not parse supported platforms:", err)
	}

	// See https://play.golang.org/p/othihEtsOBZ
	var wg = sync.WaitGroup{}
	guard := make(chan struct{}, *jobFlag)

	for _, oplatform := range parsedPlatforms {
		guard <- struct{}{}
		wg.Add(1)

		go func(platform Platform) {
			defer func() {
				wg.Done()

				<-guard
			}()

			// Process suffixes
			suffix := ""

			// Add `.exe` suffix on Windows
			if platform.GoOS == "windows" {
				suffix = ".exe"
			}

			// Add `.wasm` suffix on WASM
			if platform.GoOS == "js" && platform.GoArch == "wasm" {
				suffix = ".wasm"
			}

			// Construct the binary name
			resultingName := filepath.Join(*distFlag, *nameFlag+"."+platform.GoOS+"-")
			resultingIdentifier := getUnameConventionForGoArch(platform.GoArch) + suffix
			if *goarchConventions {
				resultingIdentifier = platform.GoArch + suffix
			}
			resultingName += resultingIdentifier
			resultingName += getSuffixForGoOSAndGOArch(platform.GoOS, platform.GoArch)

			// Check if current platform matches the regex
			matched, err := regexp.MatchString(*blockFlag, platform.GoOS+"/"+platform.GoArch)
			if err != nil {
				log.Fatal("Could not match check if platform should be blocked based on regex:", err)
			}

			// Skip the platform if it matches the block regex
			if matched {
				log.Printf("%v %v/%v (platform matched the provided regex)", color.New(color.FgYellow).SprintFunc()("skipping"), color.New(color.FgCyan).SprintFunc()(platform.GoOS), color.New(color.FgMagenta).SprintFunc()(platform.GoArch))

				return
			}

			// Continue if it does not match the regex
			log.Printf("%v %v/%v (%v)", color.New(color.FgGreen).SprintFunc()("building"), color.New(color.FgCyan).SprintFunc()(platform.GoOS), color.New(color.FgMagenta).SprintFunc()(platform.GoArch), resultingName)

			// Construct the build command
			args := append(append([]string{"build", "-o"}, []string{shellescape.Quote(resultingName)}...), shellescape.Quote(input))

			// Add the extra args if they are defined
			if *extraArgs != "" {
				args = append([]string{*extraArgs}, args...)
			}

			cmd := exec.Command(goCmd, args...)

			// Set `GOOS` and `GOARCH` env vars
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GOOS="+shellescape.Quote(platform.GoOS), "GOARCH="+shellescape.Quote(platform.GoArch))

			// Capture stdout and stderr
			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb

			// Start the build
			if err := cmd.Run(); err != nil {
				log.Fatalf("Could not build for platform %v/%v: err=%v, stdout=%v, stderr=%v", platform.GoOS, platform.GoArch, err, outb.String(), errb.String())
			}

		}(oplatform)
	}

	wg.Wait()
}
