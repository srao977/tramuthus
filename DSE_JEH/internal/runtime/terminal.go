package runtime

import (
	"fmt"
	"io"
	"sync"
	"time"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

type Terminal struct {
	mu     sync.Mutex
	output io.Writer
}

func NewTerminal(output io.Writer) *Terminal {
	return &Terminal{output: output}
}

func (terminal *Terminal) Event(category, format string, args ...any) {
	terminal.mu.Lock()
	defer terminal.mu.Unlock()
	fmt.Fprintf(terminal.output, "[%s] %-11s | %s\n", time.Now().Format("15:04:05.000"), category, fmt.Sprintf(format, args...))
}

func (terminal *Terminal) Status(state *State) string {
	snapshot := state.Snapshot()
	lastBar, lastBarUnixMs := state.LastBar()
	if lastBar == "" {
		lastBar = "NONE"
	}
	idle := "NONE"
	if lastBarUnixMs > 0 {
		idle = time.Since(time.UnixMilli(lastBarUnixMs)).Round(time.Millisecond).String()
	}
	activity := snapshot.GetActivity()
	line := fmt.Sprintf("%s | source=%s | bars=%d | admitted=%d | rejected=%d | initializing=%d | eligible=%d | last_bar=%s | idle=%s | health=%s",
		shortLifecycle(snapshot.GetLifecycleStatus()), shortSource(snapshot.GetSourceStatus().GetConnectionStatus()), activity.GetBarsReceived(), activity.GetBarsAdmitted(), activity.GetBarsRejected(), activity.GetBarsInitializing(), activity.GetBarsPhaseEligible(), lastBar, idle, shortHealth(snapshot.GetHealthStatus()))
	terminal.Event("STATUS", "%s", line)
	return line
}

func shortLifecycle(value dsejehv1.RuntimeLifecycleStatus) string {
	return trimPrefix(value.String(), "RUNTIME_LIFECYCLE_STATUS_")
}

func shortSource(value dsejehv1.SourceConnectionStatus) string {
	return trimPrefix(value.String(), "SOURCE_CONNECTION_STATUS_")
}

func shortHealth(value dsejehv1.RuntimeHealthStatus) string {
	return trimPrefix(value.String(), "RUNTIME_HEALTH_STATUS_")
}

func trimPrefix(value, prefix string) string {
	if len(value) >= len(prefix) && value[:len(prefix)] == prefix {
		return value[len(prefix):]
	}
	return value
}
