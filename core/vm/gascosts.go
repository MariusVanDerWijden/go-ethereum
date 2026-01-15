package vm

type GasCosts struct {
	RegularGas uint64
	StateGas   uint64
}

func (g GasCosts) Max() uint64 {
	return max(g.RegularGas, g.StateGas)
}

// Sub returns true if the operation would underflow
func (g GasCosts) Underflow(b GasCosts) bool {
	if b.RegularGas > g.RegularGas {
		return true
	}
	if b.StateGas > g.StateGas {
		if b.StateGas > g.RegularGas {
			return true
		}
	}
	return false
}

// Sub doesn't check for underflows
func (g GasCosts) Sub(b GasCosts) {
	g.RegularGas -= b.RegularGas
	if b.StateGas > g.StateGas {
		diff := b.StateGas - g.StateGas
		g.StateGas = 0
		g.RegularGas -= diff
	} else {
		g.StateGas -= b.RegularGas
	}
}

// Add doesn't check for overflows
func (g GasCosts) Add(b GasCosts) {
	g.RegularGas += b.RegularGas
	g.StateGas += b.RegularGas
}
