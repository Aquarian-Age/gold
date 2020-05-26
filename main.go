package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go101.org/gold/server"
	"go101.org/gold/util"
)

func init() {
	// Leave some power for others.
	var numProcessors = runtime.NumCPU() * 3 / 4
	if numProcessors < 1 {
		numProcessors = 1
	}
	runtime.GOMAXPROCS(numProcessors)
}

func main() {
	// This is used for updating Gold. It is invisible to users.
	var roughBuildTimeFlag = flag.Bool("rough-build-time", false, "show rough build time")

	flag.Parse()
	if *hFlag || *helpFlag {
		printUsage(os.Stdout)
		return
	}

	if *roughBuildTimeFlag {
		fmt.Print(RoughBuildTime)
		return
	}
	var getRoughBuildTime = func() time.Time {
		output, err := util.RunShellCommand(time.Second*3, "", os.Args[0], "-rough-build-time")
		if err != nil {
			log.Printf("Run: %s -rough-build-time error: %s", os.Args[0], err)
			return time.Now()
		}

		t, err := time.Parse("2006-01-02", string(output))
		if err != nil {
			log.Printf("! parse rough build time (%s) error: %s", output, err)
			return time.Now()
		}

		return t
	}

	// ...
	log.SetFlags(log.Lshortfile)

	var validateDiir = func(dir string) string {
		if dir == "" {
			dir = "."
		} else {
			dir = strings.TrimRight(dir, "\\/")
			dir = strings.Replace(dir, "/", string(filepath.Separator), -1)
			dir = strings.Replace(dir, "\\", string(filepath.Separator), -1)
		}
		return dir
	}

	silentMode := *silentFlag || *sFlag

	if gen := *genFlag; gen {
		var viewDocsCommand = func(docsDir string) string {
			return os.Args[0] + " -dir=" + docsDir
		}
		server.Gen(validateDiir(*dirFlag), flag.Args(), silentMode, printUsage, getRoughBuildTime, viewDocsCommand)
		return
	}

	if dir := *dirFlag; dir != "" {
		log.Printf("a%sb", dir)
		util.ServeFiles(validateDiir(dir), *portFlag, silentMode)
		return
	}

	server.Run(*portFlag, flag.Args(), silentMode, printUsage, getRoughBuildTime)
}

var hFlag = flag.Bool("h", false, "show help")
var helpFlag = flag.Bool("help", false, "show help")
var genFlag = flag.Bool("gen", false, "HTML generation mode")
var dirFlag = flag.String("dir", "", "directory for file serving or HTML generation")
var portFlag = flag.String("port", "56789", "preferred server port")
var sFlag = flag.Bool("s", false, "not open a browser automatically")
var silentFlag = flag.Bool("silent", false, "not open a browser automatically")

// var versionFlag = flag.String("version", "", "show version info")

func printUsage(out io.Writer) {
	fmt.Fprintf(out, `Usage:
	%[1]v [options] [arguments]

Options:
	-h/-help
		Show help information.
		When the flags present, others will be ignored.
	-gen=OutputFolder
		Generate all doc pages in the specified folder.
		This flag will surpress "dir" and "port" flags.
	-dir
		Directory serving mode (instead of docs server mode).
		The first argument will be viewed as the served directory.
		Current directory will be used if no arguments specified.
	-port=ServicePort
		Service port, default to 56789.
		If the specified or default port is not
		availabe, a random port will be used.
	-s/-silent
		Don't open a browser automatically or don't show HTML
		file generation logs in docs generation mode.

Examples:
	%[1]v std
		Show docs of standard packages.
	%[1]v x.y.z/myapp
		Show docs of package x.y.z/myapp.
	%[1]v
		Show docs of the package in the current directory.
	%[1]v .
		Show docs of the package in the current directory.
	%[1]v ./...
		Show docs of the package and sub-packages in the
		current directory.
	%[1]v -gen=./generated ./...
		Generate HTML docs pages for the package and
		sub-packages in the current directory.
	%[1]v -dir -s
		Serving the files in the current directory
		without opening browser automatically.
`,
		os.Args[0])
}
