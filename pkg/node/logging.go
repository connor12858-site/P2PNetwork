package node

import "fmt"

// logf is a helper function to print messages while trying to keep the command prompt clean.
func (n *Node) logf(format string, args ...interface{}) {
	n.printMu.Lock()
	defer n.printMu.Unlock()

	fmt.Print("\r\033[K")

	fmt.Printf(format+"\n", args...)

	fmt.Print("> ")
}

// Logf exposes the node logger for app packages that share the same terminal.
func (n *Node) Logf(format string, args ...interface{}) {
	n.logf(format, args...)
}
