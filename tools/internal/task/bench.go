package task

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
)

var benchmarkLine = regexp.MustCompile(`^(Benchmark\S+?)(?:-\d+)?\s+\d+\s+([\d.]+)\s+ns/op`)

type BenchOptions struct {
	Pattern      string
	Benchtime    string
	Count        int
	Passes       int
	AffinityMask uint64
	Build        BuildOptions
}

type benchmarkSample struct {
	variant   string
	benchmark string
	nanos     float64
}

func (env *Env) Bench(options BenchOptions) error {
	profiles, err := env.EnsureProfiles(AllVariants, options.Build)
	if err != nil {
		return err
	}

	binaries := make([]string, len(AllVariants))
	for index, variant := range AllVariants {
		profile := ""
		if variant.PGO {
			profile = profiles[variant.profileKey()]
		}
		binary, err := env.BuildTestBinary(variant, profile)
		if err != nil {
			return err
		}
		binaries[index] = binary
	}

	if options.AffinityMask == 0 {
		Step("measuring: %d passes, no core pinning — spread will be higher", options.Passes)
	} else {
		Step("measuring: %d passes, affinity 0x%X", options.Passes, options.AffinityMask)
	}

	var samples []benchmarkSample
	for pass := 1; pass <= options.Passes; pass++ {
		for index, variant := range AllVariants {
			collected, err := env.runBenchmark(binaries[index], variant, options)
			if err != nil {
				return err
			}
			samples = append(samples, collected...)
		}
		fmt.Printf("    pass %d/%d\n", pass, options.Passes)
	}

	reportBenchmarks(samples)
	return nil
}

func (env *Env) runBenchmark(binary string, variant Variant, options BenchOptions) ([]benchmarkSample, error) {
	command := exec.Command(binary,
		"-test.run=^$",
		"-test.bench="+options.Pattern,
		"-test.benchtime="+options.Benchtime,
		"-test.count="+strconv.Itoa(options.Count),
		"-test.cpu=1",
	)
	command.Dir = env.Root
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = os.Stderr

	if err := command.Start(); err != nil {
		return nil, err
	}
	if options.AffinityMask != 0 {
		if err := tuneBenchmarkProcess(command.Process, options.AffinityMask); err != nil {
			fmt.Fprintln(os.Stderr, "could not pin the process:", err)
		}
	}
	if err := command.Wait(); err != nil {
		return nil, fmt.Errorf("benchmark %s: %w", variant.Name, err)
	}

	var samples []benchmarkSample
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		groups := benchmarkLine.FindStringSubmatch(scanner.Text())
		if groups == nil {
			continue
		}
		nanos, err := strconv.ParseFloat(groups[2], 64)
		if err != nil {
			continue
		}
		samples = append(samples, benchmarkSample{variant: variant.Name, benchmark: groups[1], nanos: nanos})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("benchmark %s produced no ns/op lines", variant.Name)
	}
	return samples, nil
}

func reportBenchmarks(samples []benchmarkSample) {
	grouped := make(map[string]map[string][]float64)
	var order []string
	for _, sample := range samples {
		if grouped[sample.benchmark] == nil {
			grouped[sample.benchmark] = make(map[string][]float64)
			order = append(order, sample.benchmark)
		}
		grouped[sample.benchmark][sample.variant] = append(grouped[sample.benchmark][sample.variant], sample.nanos)
	}
	sort.Strings(order)

	for _, benchmark := range order {
		fmt.Printf("\n== %s\n", benchmark)
		var baseline float64
		for _, variant := range AllVariants {
			values := grouped[benchmark][variant.Name]
			if len(values) == 0 {
				continue
			}
			sort.Float64s(values)
			median := values[len(values)/2]
			if baseline == 0 {
				baseline = median
			}
			spread := 100 * (values[len(values)-1] - values[0]) / median
			delta := 100 * (median - baseline) / baseline
			fmt.Printf("  %-10s %9.1f us   n=%2d  spread %5.1f%%   vs %s: %+7.2f%%\n",
				variant.Name, median/1000, len(values), spread, AllVariants[0].Name, delta)
		}
	}
}
