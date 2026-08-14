package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"test/tools/internal/task"
)

func main() {
	flags := flag.NewFlagSet("bench", flag.ExitOnError)
	flags.Usage = usage
	goamd64 := flags.String("goamd64", "auto", "GOAMD64 level: auto, v1..v4")
	pattern := flags.String("bench", "BenchmarkMapBuild|BenchmarkMapJSONBuildAndEncode", "regexp of benchmark names")
	benchtime := flags.String("benchtime", "1s", "value for -test.benchtime")
	count := flags.Int("count", 2, "runs inside a single invocation")
	passes := flags.Int("passes", 8, "how many times to walk all variants")
	affinity := flags.String("affinity", "0xfff", "affinity mask (0 disables pinning); default is the P-cores of an i5-13500")
	workers := flags.Int("workers", 100, "load workers used while collecting the profile")
	warmup := flags.Int("warmup-seconds", 1, "warmup per endpoint while collecting the profile")
	training := flags.Int("training-seconds", 5, "measured time per endpoint while collecting the profile")
	if err := flags.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	if *passes <= 0 || *count <= 0 {
		fail(fmt.Errorf("--passes and --count must be greater than zero"))
	}
	mask, err := strconv.ParseUint(strings.TrimSpace(*affinity), 0, 64)
	if err != nil {
		fail(fmt.Errorf("could not parse --affinity %q: %w", *affinity, err))
	}

	env, err := task.NewEnv(*goamd64)
	if err != nil {
		fail(err)
	}
	if err := env.Bench(task.BenchOptions{
		Pattern:      *pattern,
		Benchtime:    *benchtime,
		Count:        *count,
		Passes:       *passes,
		AffinityMask: mask,
		Build: task.BuildOptions{
			Strip:           false,
			Workers:         *workers,
			WarmupSeconds:   *warmup,
			TrainingSeconds: *training,
		},
	}); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "bench:", err)
	os.Exit(1)
}

func usage() {
	fmt.Print(`Measure every build variant

USAGE
    go tool bench [flags]

FLAGS
    --goamd64 auto|v1..v4   GOAMD64 level; auto probes this CPU
    --bench <regexp>        which benchmarks to run
    --benchtime <d>         -test.benchtime (1s)
    --count N               runs inside a single invocation (2)
    --passes N              passes over all variants (8)
    --affinity <mask>       core pinning, 0 disables it (0xfff)
    --workers N             load workers while collecting the profile (100)
    --warmup-seconds N      warmup per endpoint (1)
    --training-seconds N    measured time per endpoint (5)
`)
}
