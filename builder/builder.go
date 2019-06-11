package builder

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/otiai10/copy"
)

type dockerBuilder struct {
	options   *builderOptions
	srcPath   string
	buildPath string
	libPaths  []string
}

func newDockerBuilder(options *builderOptions) *dockerBuilder {
	return &dockerBuilder{options: options}
}

func (b *dockerBuilder) build() error {
	err := b.ensurePaths()
	if err != nil {
		return err
	}

	err = b.ensureBuildDirectory()
	if err != nil {
		return err
	}

	err = b.prepareBuildDirectory()
	if err != nil {
		return err
	}

	buildTemplate, err := b.validateBuildDirectory()
	if err != nil {
		return err
	}

	if buildTemplate {
		err = b.buildTemplate()
		if err != nil {
			return err
		}
	}

	err = b.buildImage()
	if err != nil {
		return err
	}

	return nil
}

func (b *dockerBuilder) ensurePaths() error {
	absSrcPath, err := filepath.Abs(b.options.srcPath)
	if err != nil {
		return fmt.Errorf("unable to determine absolute path for: %+v", b.options.srcPath)
	}

	b.srcPath = absSrcPath
	b.buildPath = filepath.Join(b.srcPath, "_docker-build")

	b.libPaths = make([]string, 0)
	for _, libPath := range b.options.libPaths {
		absLibPath, err := filepath.Abs(libPath)
		if err != nil {
			return fmt.Errorf("unable to determine absolute path for: %+v", libPath)
		}
		b.libPaths = append(b.libPaths, absLibPath)
	}

	return nil
}

func (b *dockerBuilder) ensureBuildDirectory() error {
	err := os.RemoveAll(b.buildPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(b.buildPath, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (b *dockerBuilder) prepareBuildDirectory() error {
	if _, err := os.Stat(filepath.Join(b.srcPath, "Dockerfile.template")); err == nil {
		// If Dockerfile.template exists, then copy it over.
		if err = copy.Copy(filepath.Join(b.srcPath, "Dockerfile.template"), filepath.Join(b.buildPath, "Dockerfile.template")); err != nil {
			return err
		}
	} else {
		// If Dockerfile.template doesn't exist ...
		if _, err := os.Stat(filepath.Join(b.srcPath, "Dockerfile")); err == nil {
			// ... but Dockerfile exists, then copy Dockerfile over ...
			if err = copy.Copy(filepath.Join(b.srcPath, "Dockerfile"), filepath.Join(b.buildPath, "Dockerfile")); err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(filepath.Join(b.srcPath, "src")); err == nil {
		if err = copy.Copy(filepath.Join(b.srcPath, "src"), filepath.Join(b.buildPath, "src")); err != nil {
			return err
		}
	}

	return nil
}

func (b *dockerBuilder) validateBuildDirectory() (bool, error) {
	hasDockerfileTemplate := false
	hasDockerfile := false

	if _, err := os.Stat(filepath.Join(b.srcPath, "Dockerfile.template")); err == nil {
		hasDockerfileTemplate = true
	}

	if _, err := os.Stat(filepath.Join(b.srcPath, "Dockerfile")); err == nil {
		hasDockerfile = true
	}

	if hasDockerfileTemplate {
		return true, nil
	} else if hasDockerfile {
		return false, nil
	} else {
		return false, errors.New("either Dockerfile.template or Dockerfile is required")
	}
}

func (b *dockerBuilder) buildTemplate() error {
	dockerfileTemplate, err := readFileToString(filepath.Join(b.buildPath, "Dockerfile.template"))
	if err != nil {
		return err
	}

	for _, fileMatch := range regexp.MustCompile(`%%(_COMMON\/([^%]+))%%`).FindAllStringSubmatch(dockerfileTemplate, -1) {
		templateString := trimString(fileMatch[0])
		referenceString := trimString(fileMatch[1])
		pathString := trimString(fileMatch[2])

		foundFile := false
		for _, libPath := range b.libPaths {
			libFilePath := filepath.Join(libPath, pathString)
			if _, err := os.Stat(libFilePath); err == nil {
				err = copy.Copy(libFilePath, filepath.Join(b.buildPath, pathString))
				if err != nil {
					return err
				}

				dockerfileTemplate = strings.ReplaceAll(dockerfileTemplate, templateString, pathString)
				foundFile = true
				break
			}
		}

		if !foundFile {
			return fmt.Errorf("unable to locate file for: %+v", referenceString)
		}
	}

	allVariables, err := b.readAllVariables()
	if err != nil {
		return err
	}

	for _, variableMatch := range regexp.MustCompile(`%%([^%]+)%%`).FindAllStringSubmatch(dockerfileTemplate, -1) {
		templateString := trimString(variableMatch[0])
		referenceString := trimString(variableMatch[1])

		valueString := trimString(allVariables[referenceString])

		if valueString == "" {
			return fmt.Errorf("unable to locate value for: %+v", referenceString)
		}

		dockerfileTemplate = strings.ReplaceAll(dockerfileTemplate, templateString, valueString)
	}

	return writeStringToFile(filepath.Join(b.buildPath, "Dockerfile"), dockerfileTemplate)
}

func (b *dockerBuilder) readAllVariables() (map[string]string, error) {
	variables := map[string]string{}

	fileVariables, err := b.readFileVariables()
	if err != nil {
		return nil, err
	}

	for key, value := range fileVariables {
		variables[key] = value
	}

	fragmentVariables, err := b.readFragmentVariables()
	if err != nil {
		return nil, err
	}

	for key, value := range fragmentVariables {
		variables[key] = value
	}

	return variables, nil
}

func (b *dockerBuilder) readFileVariables() (map[string]string, error) {
	variables := map[string]string{}
	filePaths := make([]string, 0)

	srcVariableFilePath := filepath.Join(b.srcPath, "Dockerfile.variables")
	if _, err := os.Stat(srcVariableFilePath); err == nil {
		filePaths = append(filePaths, srcVariableFilePath)
	}

	for _, libPath := range b.libPaths {
		libVariableFilePath := filepath.Join(libPath, "Dockerfile.variables")
		if _, err := os.Stat(libVariableFilePath); err == nil {
			filePaths = append(filePaths, libVariableFilePath)
		}
	}

	for _, filePath := range filePaths {
		fileContent, err := readFileToString(filePath)
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(trimString(fileContent), "\n") {
			parts := strings.SplitN(trimString(line), ":", 2)
			if len(parts) == 2 {
				variables[trimString(parts[0])] = trimString(parts[1])
			}
		}
	}

	return variables, nil
}

func (b *dockerBuilder) readFragmentVariables() (map[string]string, error) {
	variables := map[string]string{}
	fragmentsPaths := make([]string, 0)

	srcFragmentsPath := filepath.Join(b.srcPath, "Dockerfile.variables.d")
	if _, err := os.Stat(srcFragmentsPath); err == nil {
		fragmentsPaths = append(fragmentsPaths, srcFragmentsPath)
	}

	for _, libPath := range b.libPaths {
		libFragmentsPath := filepath.Join(libPath, "Dockerfile.variables.d")
		if _, err := os.Stat(libFragmentsPath); err == nil {
			fragmentsPaths = append(fragmentsPaths, libFragmentsPath)
		}
	}

	for _, fragmentsPath := range fragmentsPaths {
		files, err := ioutil.ReadDir(fragmentsPath)
		if err != nil {
			return nil, err
		}

		for _, fragment := range files {
			fragmentName := trimString(fragment.Name())
			fragmentContent, err := readFileToString(filepath.Join(fragmentsPath, fragmentName))
			if err != nil {
				return nil, err
			}
			variables[fragmentName] = trimString(fragmentContent)
		}
	}

	return variables, nil
}

func (b *dockerBuilder) buildImage() error {
	err := os.RemoveAll(filepath.Join(b.buildPath, "Dockerfile.template"))
	if err != nil {
		return err
	}

	if b.options.dryRun {
		buildCommand := fmt.Sprintf("docker build %+v %+v", strings.Join(b.options.buildOpts, " "), b.buildPath)

		dockerFile, err := readFileToString(filepath.Join(b.buildPath, "Dockerfile"))
		if err != nil {
			return err
		}

		fmt.Printf("*** DRY RUN > Build Command:\n%+v\n\n", buildCommand)
		fmt.Printf("*** DRY RUN > Dockerfile:\n")
		fmt.Printf("%+v\n", trimString(dockerFile))
	} else {
		dockerOptions := []string{"build"}
		dockerOptions = append(dockerOptions, b.options.buildOpts...)
		dockerOptions = append(dockerOptions, b.buildPath)

		cmd := exec.Command("docker", dockerOptions...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

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

	builder := newDockerBuilder(opts)
	err = builder.build()
	if err != nil {
		fmt.Printf("Error building Dockerfile: %+v\n", err)
		return 1
	}

	return 0
}
