package builder

import (
	"fmt"
)

func Start(args []string) int {
	opts, err := parseOptions(args)
	if err != nil {
		fmt.Printf("%+v\n", err)
		fmt.Printf("%+v\n", helpMessage())
		return 1
	}

	err = validateOptions(opts)
	if err != nil {
		fmt.Printf("%+v\n", err)
		fmt.Printf("%+v\n", helpMessage())
		return 1
	}

	fmt.Printf("* Source path: %+v\n", opts.srcPath)
	for _, libPath := range opts.libPaths {
		fmt.Printf("* Lib path: %+v\n", libPath)
	}
	fmt.Printf("* Build options: %+v\n", opts.buildOpts)

	return 0
}
