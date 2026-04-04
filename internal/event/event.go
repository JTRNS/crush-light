package event

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/charmbracelet/crush/internal/version"
)

const (
	nonInteractiveAttrName      = "NonInteractive"
	continueSessionByIDAttrName = "ContinueSessionByID"
	continueLastSessionAttrName = "ContinueLastSession"
)

type Properties map[string]any

var (
	basePropsMu sync.RWMutex

	baseProps = Properties{
		"GOOS":                 runtime.GOOS,
		"GOARCH":               runtime.GOARCH,
		"TERM":                 os.Getenv("TERM"),
		"SHELL":                filepath.Base(os.Getenv("SHELL")),
		"Version":              version.Version,
		"GoVersion":            runtime.Version(),
		nonInteractiveAttrName: false,
	}
)

func SetNonInteractive(nonInteractive bool) {
	setBaseProp(nonInteractiveAttrName, nonInteractive)
}

func SetContinueBySessionID(continueBySessionID bool) {
	setBaseProp(continueSessionByIDAttrName, continueBySessionID)
}

func SetContinueLastSession(continueLastSession bool) {
	setBaseProp(continueLastSessionAttrName, continueLastSession)
}

// Init is a no-op lifecycle hook kept for API compatibility after telemetry
// removal.
func Init() {}

// GetID intentionally returns an empty string because the local event system
// no longer tracks device or user identifiers.
func GetID() string { return "" }

// Alias preserves the legacy event alias hook without sending remote data.
func Alias(userID string) {
	if userID == "" {
		return
	}
}

// send records an internal event for future local handling.
func send(event string, props ...any) {
	if event == "" {
		return
	}
	validateProps(props...)
}

// Error records an internal error event for future local handling.
func Error(errToLog any, props ...any) {
	if errToLog == nil {
		return
	}
	validateProps(props...)
}

// Flush preserves the legacy event flush hook without remote telemetry.
func Flush() {}

func pairsToProps(props ...any) Properties {
	p := Properties{}

	if !isEven(len(props)) {
		slog.Error("Event properties must be provided as key-value pairs", "props", props)
		return p
	}

	for i := 0; i < len(props); i += 2 {
		key, ok := props[i].(string)
		if !ok {
			slog.Error("Event property key must be a string", "key", props[i], "index", i)
			continue
		}
		value := props[i+1]
		p[key] = value
	}
	return p
}

func (p Properties) Merge(other Properties) Properties {
	out := make(Properties, len(p)+len(other))
	for key, value := range other {
		out[key] = value
	}
	for key, value := range p {
		out[key] = value
	}
	return out
}

func cloneBaseProps() Properties {
	basePropsMu.RLock()
	defer basePropsMu.RUnlock()

	out := make(Properties, len(baseProps))
	for key, value := range baseProps {
		out[key] = value
	}
	return out
}

func setBaseProp(key string, value any) {
	basePropsMu.Lock()
	defer basePropsMu.Unlock()
	baseProps[key] = value
}

func validateProps(props ...any) {
	_ = pairsToProps(props...).Merge(cloneBaseProps())
}

func isEven(n int) bool {
	return n%2 == 0
}
