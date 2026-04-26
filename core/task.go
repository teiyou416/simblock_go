package core

// SimTime is the simulation clock unit used by the engine.
// The unit is intentionally abstract (e.g., milliseconds).
type SimTime int64

// Task is the minimal executable event in the simulator.
//
// Interval returns when this task should run relative to the current
// simulation time. Run executes task behavior at the scheduled time.
type Task interface {
	Interval() SimTime
	Run()
}

// IdentifiedTask is an optional extension for tasks that need a stable ID
// for tracing, logging, or debugging.
//
// It is intentionally not part of Task to keep the core engine decoupled.
type IdentifiedTask interface {
	TaskID() uint64
}
