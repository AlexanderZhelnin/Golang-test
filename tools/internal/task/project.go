package task

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const ServerAddress = "127.0.0.1:4000"

const PGOProfileEnv = "API_TEST_PGO_PROFILE"

const projectMarker = "primitives.json"

type Variant struct {
	Name         string
	GOEXPERIMENT string
	PGO          bool
}

var (
	Pure    = Variant{Name: "pure", GOEXPERIMENT: "", PGO: false}
	SIMD    = Variant{Name: "simd", GOEXPERIMENT: "simd", PGO: false}
	PurePGO = Variant{Name: "pgo", GOEXPERIMENT: "", PGO: true}
	SIMDPGO = Variant{Name: "simd-pgo", GOEXPERIMENT: "simd", PGO: true}

	AllVariants = []Variant{Pure, SIMD, PurePGO, SIMDPGO}
)

func ParseVariant(name string) (Variant, error) {
	for _, candidate := range AllVariants {
		if candidate.Name == name {
			return candidate, nil
		}
	}
	names := make([]string, 0, len(AllVariants))
	for _, candidate := range AllVariants {
		names = append(names, candidate.Name)
	}
	return Variant{}, fmt.Errorf("unknown variant %q: expected one of %s", name, strings.Join(names, ", "))
}

func (variant Variant) profileKey() string {
	if variant.GOEXPERIMENT == "" {
		return "pure"
	}
	return "simd"
}

type Env struct {
	Root    string
	GOAMD64 string
}

func NewEnv(requestedGOAMD64 string) (*Env, error) {
	root, err := findRoot()
	if err != nil {
		return nil, err
	}

	level := requestedGOAMD64
	if level == "" || level == "auto" {
		detected, err := detectGOAMD64()
		if err != nil {
			return nil, err
		}
		level = detected
		Step("GOAMD64=%s detected for this CPU", level)
	} else {
		Step("GOAMD64=%s set explicitly", level)
	}
	return &Env{Root: root, GOAMD64: level}, nil
}

func findRoot() (string, error) {
	working, err := os.Getwd()
	if err != nil {
		return "", err
	}
	directory := working
	for {
		_, markerErr := os.Stat(filepath.Join(directory, projectMarker))
		_, modErr := os.Stat(filepath.Join(directory, "go.mod"))
		if markerErr == nil && modErr == nil {
			return directory, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("project directory not found (%s next to go.mod), searched upwards from %s", projectMarker, working)
		}
		directory = parent
	}
}

func detectGOAMD64() (string, error) {
	if runtime.GOARCH != "amd64" {
		return "", fmt.Errorf("GOAMD64 detection only makes sense on amd64, not on %s", runtime.GOARCH)
	}

	directory, err := os.MkdirTemp("", "goamd64-probe")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(directory)

	if err := os.WriteFile(filepath.Join(directory, "go.mod"), []byte("module goamd64probe\n\ngo 1.21\n"), 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(directory, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		return "", err
	}

	for _, level := range []string{"v4", "v3", "v2"} {
		binary := filepath.Join(directory, "probe-"+level+exeSuffix())
		build := exec.Command("go", "build", "-o", binary, ".")
		build.Dir = directory
		build.Env = append(os.Environ(), "GOEXPERIMENT=", "GOAMD64="+level, "GOFLAGS=")
		if output, err := build.CombinedOutput(); err != nil {
			return "", fmt.Errorf("building GOAMD64=%s probe: %w: %s", level, err, output)
		}

		probe := exec.Command(binary)
		probe.Dir = directory
		if err := probe.Run(); err == nil {
			return level, nil
		}
	}
	return "v1", nil
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func (env *Env) BinDir() string {
	return filepath.Join(env.Root, "bin")
}

func (env *Env) ProfileDir() string {
	return filepath.Join(env.BinDir(), "profiles")
}

func (env *Env) ProfilePath(variant Variant) string {
	return filepath.Join(env.ProfileDir(), variant.profileKey()+".pprof")
}

func (env *Env) ServerBinary(variant Variant) string {
	return filepath.Join(env.BinDir(), "api-"+variant.Name+exeSuffix())
}

func (env *Env) TestBinary(variant Variant) string {
	return filepath.Join(env.BinDir(), variant.Name+".test"+exeSuffix())
}

func (env *Env) Relative(path string) string {
	relative, err := filepath.Rel(env.Root, path)
	if err != nil {
		return path
	}
	return relative
}

func (env *Env) goEnvironment(variant Variant) []string {
	return append(os.Environ(),
		"GOEXPERIMENT="+variant.GOEXPERIMENT,
		"GOAMD64="+env.GOAMD64,
	)
}

func (env *Env) RunGo(variant Variant, args ...string) error {
	command := exec.Command("go", args...)
	command.Dir = env.Root
	command.Env = env.goEnvironment(variant)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("go %s (%s): %w", strings.Join(args, " "), variant.Name, err)
	}
	return nil
}

func Step(format string, arguments ...any) {
	fmt.Printf("==> "+format+"\n", arguments...)
}
