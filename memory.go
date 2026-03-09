package unrevealed

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/process"
)

// MemoryStats holds aggregated memory information for the browser process tree.
type MemoryStats struct {
	RSS uint64 // Resident Set Size in bytes (sum of entire process tree).
	VMS uint64 // Virtual Memory Size in bytes (sum of entire process tree).
}

// BrowserMemory returns the aggregated RSS and VMS for the browser's
// entire process tree (root process + all descendants).
// Returns an error if the browser process is not running.
func (b *Browser) BrowserMemory() (MemoryStats, error) {
	b.mu.Lock()
	cmd := b.cmd
	closed := b.closed
	b.mu.Unlock()

	if closed || cmd == nil || cmd.Process == nil {
		return MemoryStats{}, fmt.Errorf("browser process is not running")
	}

	root, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return MemoryStats{}, fmt.Errorf("open browser process: %w", err)
	}

	return sumTree(root)
}

// sumTree recursively sums memory stats for a process and all its children.
func sumTree(p *process.Process) (MemoryStats, error) {
	var stats MemoryStats

	mem, err := p.MemoryInfo()
	if err != nil {
		return stats, fmt.Errorf("memory info for pid %d: %w", p.Pid, err)
	}
	stats.RSS = mem.RSS
	stats.VMS = mem.VMS

	children, _ := p.Children() // ignore error no children are fine
	for _, child := range children {
		cs, err := sumTree(child)
		if err != nil {
			continue // child may have exited between listing and reading
		}
		stats.RSS += cs.RSS
		stats.VMS += cs.VMS
	}
	return stats, nil
}
