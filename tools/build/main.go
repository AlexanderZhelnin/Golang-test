package main

import (
	"flag"
	"fmt"
	"go_test/tools/internal/task"
	"os"
)

func main() {
	flags := flag.NewFlagSet("build", flag.ExitOnError)
	flags.Usage = usage
	goamd64 := flags.String("goamd64", "auto", "GOAMD64 level: auto, v1..v4")
	strip := flags.Bool("strip", true, `pass -ldflags="-s -w"`)
	workers := flags.Int("workers", 100, "load workers used while collecting the profile")
	warmup := flags.Int("warmup-seconds", 1, "warmup per endpoint while collecting the profile")
	training := flags.Int("training-seconds", 5, "measured time per endpoint while collecting the profile")

	// Флаги идут после позиционного варианта, поэтому разбор в два приёма.
	arguments := os.Args[1:]
	var variantName string
	if len(arguments) > 0 && !isFlag(arguments[0]) {
		variantName, arguments = arguments[0], arguments[1:]
	}
	if err := flags.Parse(arguments); err != nil {
		os.Exit(1)
	}

	variants := task.AllVariants
	if variantName != "" {
		variant, err := task.ParseVariant(variantName)
		if err != nil {
			fail(err)
		}
		variants = []task.Variant{variant}
	}

	env, err := task.NewEnv(*goamd64)
	if err != nil {
		fail(err)
	}
	if err := env.Build(variants, task.BuildOptions{
		Strip:           *strip,
		Workers:         *workers,
		WarmupSeconds:   *warmup,
		TrainingSeconds: *training,
	}); err != nil {
		fail(err)
	}
}

func isFlag(argument string) bool {
	return len(argument) > 0 && argument[0] == '-'
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "build:", err)
	os.Exit(1)
}

func usage() {
	fmt.Print(`Build the project server binaries

USAGE
    go tool build [variant] [flags]

VARIANTS
    (no argument)  build all four
    pure           no SIMD, no PGO
    simd           GOEXPERIMENT=simd
    pgo            with PGO
    simd-pgo       GOEXPERIMENT=simd and with PGO

FLAGS
    --goamd64 auto|v1..v4   GOAMD64 level; auto probes this CPU
    --strip                 -ldflags="-s -w" (on by default)
    --workers N             load workers while collecting the profile (100)
    --warmup-seconds N      warmup per endpoint (1)
    --training-seconds N    measured time per endpoint (5)

Binaries go to bin/api-<variant>, profiles to bin/profiles/.
`)
}
