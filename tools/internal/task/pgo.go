package task

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

var projectFiles = []string{
	"primitives.json",
}

var trainingWorkloads = []struct {
	name     string
	requests []string
}{
	{
		name: "/map",
		requests: []string{
			"/map?x=0&y=0",
		},
	},
	{
		name: "/mapBlazing",
		requests: []string{
			"/mapBlazing?x=0&y=0",
		},
	},
	{
		name: "/mapJSON",
		requests: []string{
			"/mapJSON?x=0&y=0",
		},
	},
	{
		name: "/mapJSONBlazing",
		requests: []string{
			"/mapJSONBlazing?x=0&y=0",
		},
	},
	{
		name: "/naturalsort",
		requests: []string{
			"/naturalsort",
		},
	},
	{
		name: "/naturalsortBlazing",
		requests: []string{
			"/naturalsortBlazing",
		},
	},
}

func copyFile(src, dst string) error {
	// Открываем исходный файл для чтения
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Создаём (или усекаем) файл назначения
	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	// Копируем содержимое
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Синхронизируем метаданные на диск
	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

func (env *Env) collectProfile(variant Variant, options BuildOptions) (string, error) {
	if err := ensurePortFree(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(env.ProfileDir(), 0o755); err != nil {
		return "", err
	}
	profile := env.ProfilePath(variant)

	training := filepath.Join(env.BinDir(), "training-"+variant.profileKey()+exeSuffix())

	for i := range projectFiles {
		copyFile(filepath.Join(env.Root, projectFiles[i]), filepath.Join(env.BinDir(), projectFiles[i]))
	}

	trainingVariant := Variant{
		Name:         variant.profileKey() + "/training",
		GOEXPERIMENT: variant.GOEXPERIMENT,
	}

	if _, err := env.buildServer(trainingVariant, "", false, training); err != nil {
		return "", err
	}

	defer os.Remove(training)

	Step("profiling %s: %d workers, %s warmup, %s measured per endpoint",
		variant.profileKey(), options.Workers, options.warmup(), options.training())

	server := exec.Command(training)
	server.Dir = env.Root
	server.Env = append(os.Environ(), PGOProfileEnv+"="+profile)
	server.Stdout = io.Discard
	server.Stderr = os.Stderr
	stdin, err := server.StdinPipe()
	if err != nil {
		return "", err
	}
	if err := server.Start(); err != nil {
		return "", fmt.Errorf("starting training server: %w", err)
	}

	stopped := false
	defer func() {
		if !stopped {
			_ = server.Process.Kill()
			_ = server.Wait()
		}
	}()

	if err := waitUntilReady(20 * time.Second); err != nil {
		return "", err
	}

	client := newLoadClient(options.Workers)
	for _, workload := range trainingWorkloads {
		if options.warmup() > 0 {
			driveLoad(client, workload.requests, options.Workers, options.warmup())
		}
		completed, failures := driveLoad(client, workload.requests, options.Workers, options.training())
		if failures > 0 {
			return "", fmt.Errorf("workload %s: %d failed requests", workload.name, failures)
		}
		Step("  %-13s %d requests", workload.name, completed)
	}
	client.CloseIdleConnections()

	stopped = true
	_, _ = io.WriteString(stdin, "\n")
	_ = stdin.Close()
	if err := server.Wait(); err != nil {
		return "", fmt.Errorf("stopping training server: %w", err)
	}

	info, err := os.Stat(profile)
	if err != nil {
		return "", fmt.Errorf("server did not write the profile %s: %w", profile, err)
	}
	if info.Size() == 0 {
		return "", fmt.Errorf("profile %s is empty", profile)
	}
	Step("profile %s ready (%d KiB)", env.Relative(profile), info.Size()/1024)
	return profile, nil
}

func newLoadClient(workers int) *http.Client {
	return &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			MaxIdleConns:        workers * 2,
			MaxIdleConnsPerHost: workers * 2,
			MaxConnsPerHost:     workers * 2,
			IdleConnTimeout:     time.Minute,
			DisableCompression:  true,
		},
	}
}

func driveLoad(client *http.Client, requests []string, workers int, duration time.Duration) (int64, int64) {
	var completed, failures atomic.Int64
	deadline := time.Now().Add(duration)

	var group sync.WaitGroup
	for worker := range workers {
		group.Add(1)
		go func(worker int) {
			defer group.Done()
			for index := 0; time.Now().Before(deadline); index++ {
				target := requests[(worker+index)%len(requests)]
				if err := fetchDiscard(client, target); err != nil {
					failures.Add(1)
					continue
				}
				completed.Add(1)
			}
		}(worker)
	}
	group.Wait()
	return completed.Load(), failures.Load()
}

func fetchDiscard(client *http.Client, path string) error {
	response, err := client.Get("http://" + ServerAddress + path)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: status %d", path, response.StatusCode)
	}
	return nil
}

func ensurePortFree() error {
	listener, err := net.Listen("tcp", ServerAddress)
	if err != nil {
		return fmt.Errorf("port %s is busy: stop the running server", ServerAddress)
	}
	return listener.Close()
}

func waitUntilReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		connection, err := net.DialTimeout("tcp", ServerAddress, 200*time.Millisecond)
		if err == nil {
			return connection.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("server did not come up on %s within %s", ServerAddress, timeout)
}
