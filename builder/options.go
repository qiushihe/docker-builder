package builder

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type builderOptions struct {
	dryRun    bool
	srcPath   string
	libPaths  []string
	buildOpts []string
}

type stringsFlag []string

func (v *stringsFlag) String() string {
	if v != nil {
		return strings.Join(*v, ", ")
	}
	return ""
}

func (v *stringsFlag) Set(value string) error {
	*v = append(*v, value)
	return nil
}

func helpMessage() string {
	return strings.Join([]string{
		"Usage:",
		"  $ docker-builder -src [SOURCE PATH] (-lib [LIB PATH])* (-dry)? -- [BUILD OPTIONS]",
		"Example:",
		"  $ docker-builder \\",
		"    -src path/to/image/dir \\",
		"    -lib path/to/lib \\",
		"    -lib path/to/another/lib \\",
		"    -- -t my/image:tag",
	}, "\n")
}

func parseOptions(args []string) (*builderOptions, error) {
	var dryRunFlag bool
	var srcPathFlag string
	var libPathFlags stringsFlag

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.BoolVar(&dryRunFlag, "dry", false, "dry run")
	flags.StringVar(&srcPathFlag, "src", "", "source directory")
	flags.Var(&libPathFlags, "lib", "lib directory")

	err := flags.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	libPaths := make([]string, 0)
	for _, libPathFlag := range libPathFlags {
		libPaths = append(libPaths, trimString(libPathFlag))
	}

	buildOpts := make([]string, 0)
	for _, buildArg := range flags.Args() {
		if strings.Contains(buildArg, " ") {
			buildOpts = append(buildOpts, fmt.Sprintf("\"%+v\"", buildArg))
		} else {
			buildOpts = append(buildOpts, buildArg)
		}
	}

	return &builderOptions{
		dryRun:    dryRunFlag,
		srcPath:   trimString(srcPathFlag),
		libPaths:  libPaths,
		buildOpts: buildOpts,
	}, nil
}

func validateOptions(opts *builderOptions) error {
	if len(opts.srcPath) <= 0 {
		return errors.New("missing source directory")
	}

	if len(opts.buildOpts) <= 0 {
		return errors.New("missing build options")
	}

	return nil
}
