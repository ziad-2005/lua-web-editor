package execution

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

type Result struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"` 
	DurationMs int64  `json:"duration_ms"`
	Status     string `json:"status"`
}


func RunLua(code string, timeout time.Duration) (Result, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/usr/bin/lua", "-e", code)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	if ctx.Err() == context.DeadlineExceeded {
		return  Result{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			ExitCode: -1,
			DurationMs: duration,
			Status: "time_out",
		}, nil
	}

	exitCode := 0
	status := "completed"
	if err != nil {
		status = "failed"
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return  Result{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		ExitCode:   exitCode,
		DurationMs: duration,
		Status:     status,
	}, nil
}
