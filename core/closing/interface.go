package closing

// Closer closes all stuff released by an object
type Closer interface {
	Close() error
}

// IntRandomizer interface provides functionality over generating integer numbers.
//
// ISSUE-045 closure: signature is `Intn(n int) (int, error)` so callers
// can react to entropy-source failures (e.g. /dev/urandom unavailable
// in containers, FIPS-mode kernel restrictions) without crashing the
// process. The previous `Intn(n int) int` shape forced implementations
// to either panic or silently return a deterministic value — both bad
// outcomes for the consensus / heartbeat / shutdown-delay code paths
// that consume this interface.
type IntRandomizer interface {
	Intn(n int) (int, error)
	IsInterfaceNil() bool
}
