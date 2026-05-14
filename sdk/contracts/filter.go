package contracts

// IsCommandAllowed reports whether command passes the filter.
// If filter is nil, all commands are allowed.
func IsCommandAllowed(command byte, filter *MIDIEventFilter) bool {
	if filter == nil {
		return true
	}
	for _, allowed := range filter.Commands {
		if command == byte(allowed) {
			return true
		}
	}
	return false
}
