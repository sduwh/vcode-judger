package sandbox

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"sync"
	"time"
)

const (
	maxBinarySize    = 100 << 20
	runTimeout       = 5 * time.Second
	maxOutputSize    = 100 << 20
	memoryLimitBytes = 100 << 20
)

const containedStartMessage = "golang-gvisor-process-started\n"


var (
	container  = flag.String("untrusted-container", "gcr.io/golang-org/playground-sandbox-gvisor:latest", "container image name that hosts the untrusted binary under gvisor")
)
var (
	readyContainer chan *Container
	runSem         chan struct{}
)

var (
	wantedMu        sync.Mutex
	containerWanted = map[string]bool{}
)

// run code container
type Container struct {
	// container name
	name string

	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	// command to control the container
	cmd      *exec.Cmd
	waitOnce sync.Once
	waitVal  error
}

func (c *Container) Close() {
	setContainerWanted(c.name, false)
	_ = c.stdin.Close()
	_ = c.stdout.Close()
	_ = c.stderr.Close()
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.Wait()
	}
}

func (c *Container) Wait() error {
	c.waitOnce.Do(c.wait)
	return c.waitVal
}

func (c *Container) wait() {
	c.waitVal = c.cmd.Wait()
}

func setContainerWanted(name string, wanted bool) {
	// map is not tread safe
	wantedMu.Lock()
	defer wantedMu.Unlock()
	if wanted {
		containerWanted[name] = true
	} else {
		delete(containerWanted, name)
	}
}

func getContainer(ctx context.Context) (*Container, error) {
	select {
	case c := <-readyContainer:
		return c, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func startContainer(ctx context.Context) (c *Container, err error) {
	name := "run_" + randHex(8)
	setContainerWanted(name, true)
	var stdin io.WriteCloser
	var stdout io.ReadCloser
	var stderr io.ReadCloser
	defer func() {
		if err == nil {
			return
		}
		setContainerWanted(name, false)
		if stdin != nil {
			_ = stdin.Close()
		}
		if stdout != nil {
			_ = stdout.Close()
		}
		if stderr != nil {
			_ = stderr.Close()
		}
	}()
	// create the command
	cmd := exec.Command(
		"docker",
		"run",
		"--name="+name,
		"--rm",
		"--tmpfs=/tmpfs",
		"-i",
		"--runtime=runsc",
		"--network=none",
		"--memory="+fmt.Sprint(memoryLimitBytes),
		*container,
		"--mode=contained")
	stdin, err = cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// execute the command
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	errc := make(chan error, 1)
	go func() {
		buf := make([]byte, len(containedStartMessage))
		if _, err := io.ReadFull(stdout, buf); err != nil {
			errc <- fmt.Errorf("error reading header from sandbox container: %v", err)
			return
		}
		if string(buf) != containedStartMessage {
			errc <- fmt.Errorf("sandbox container sent wrong header %q; want %q", buf, containedStartMessage)
			return
		}
		errc <- nil
	}()
	select {
	case <-ctx.Done():
		logrus.Printf("timeout starting container")
		_ = cmd.Process.Kill()
		return nil, ctx.Err()
	case err := <-errc:
		if err != nil {
			logrus.Printf("error starting container: %v", err)
			return nil, err
		}
	}
	return &Container{
		name:   name,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		cmd:    cmd,
	}, nil
}

func randHex(n int) string {
	b := make([]byte, n/2)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
