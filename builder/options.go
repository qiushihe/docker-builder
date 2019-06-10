package builder

import (
	"errors"
	"flag"
	"strings"
)

type builderOptions struct {
	srcPath   string
	libPaths  []string
	buildOpts string
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

func trimString(val string) string {
	return strings.Trim(val, " \r\n")
}

func helpMessage() string {
	return strings.Join([]string{
		"Usage:",
		"  $ docker-builder -src [SOURCE PATH] (-lib [LIB PATH])* -- [BUILD OPTIONS]",
		"Example:",
		"  $ docker-builder \\",
		"    -src path/to/image/dir \\",
		"    -lib path/to/lib \\",
		"    -lib path/to/another/lib \\",
		"    -- -t my/image:tag",
	}, "\n")
}

func parseOptions(args []string) (*builderOptions, error) {
	var srcPathFlag string
	var libPathFlags stringsFlag

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
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

	return &builderOptions{
		srcPath:   trimString(srcPathFlag),
		libPaths:  libPaths,
		buildOpts: trimString(strings.Join(flags.Args(), " ")),
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
