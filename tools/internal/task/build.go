package task

import (
	"os"
	"path/filepath"
	"time"
)

type BuildOptions struct {
	Strip           bool
	Workers         int
	WarmupSeconds   int
	TrainingSeconds int
}

func (env *Env) Build(variants []Variant, options BuildOptions) error {
	if len(variants) == len(AllVariants) {
		if _, err := os.Stat(env.BinDir()); err == nil {
			Step("cleaning %s", env.Relative(env.BinDir()))
			if err := os.RemoveAll(env.BinDir()); err != nil {
				return err
			}
		}
	}

	profiles, err := env.EnsureProfiles(variants, options)
	if err != nil {
		return err
	}

	for _, variant := range variants {
		profile := ""
		if variant.PGO {
			profile = profiles[variant.profileKey()]
		}
		if _, err := env.buildServer(variant, profile, options.Strip, ""); err != nil {
			return err
		}
	}
	return nil
}

func (env *Env) buildServer(variant Variant, profile string, strip bool, output string) (string, error) {
	if output == "" {
		output = env.ServerBinary(variant)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return "", err
	}
	if err := os.Remove(output); err != nil && !os.IsNotExist(err) {
		return "", err
	}

	arguments := []string{"build", "-o", output}
	if strip {
		arguments = append(arguments, "-ldflags=-s -w")
	}
	if profile == "" {
		arguments = append(arguments, "-pgo=off")
	} else {
		arguments = append(arguments, "-pgo="+profile)
	}
	arguments = append(arguments, ".")

	Step("building %s -> %s", variant.Name, env.Relative(output))
	if err := env.RunGo(variant, arguments...); err != nil {
		return "", err
	}
	return output, nil
}

func (env *Env) BuildTestBinary(variant Variant, profile string) (string, error) {
	output := env.TestBinary(variant)
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return "", err
	}
	if err := os.Remove(output); err != nil && !os.IsNotExist(err) {
		return "", err
	}

	arguments := []string{"test", "-c", "-o", output}
	if profile == "" {
		arguments = append(arguments, "-pgo=off")
	} else {
		arguments = append(arguments, "-pgo="+profile)
	}
	arguments = append(arguments, ".")

	Step("building benchmark %s", env.Relative(output))
	if err := env.RunGo(variant, arguments...); err != nil {
		return "", err
	}
	return output, nil
}

func (env *Env) EnsureProfiles(variants []Variant, options BuildOptions) (map[string]string, error) {
	collected := make(map[string]string)
	for _, variant := range variants {
		if !variant.PGO {
			continue
		}
		key := variant.profileKey()
		if _, done := collected[key]; done {
			continue
		}
		profile, err := env.collectProfile(variant, options)
		if err != nil {
			return nil, err
		}
		collected[key] = profile
	}
	return collected, nil
}

func (options BuildOptions) warmup() time.Duration {
	return time.Duration(options.WarmupSeconds) * time.Second
}

func (options BuildOptions) training() time.Duration {
	return time.Duration(options.TrainingSeconds) * time.Second
}
