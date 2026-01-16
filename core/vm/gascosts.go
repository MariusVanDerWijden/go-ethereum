package vm

type GasCosts struct {
	RegularGas uint64
	StateGas   uint64
}

func (g GasCosts) Max() uint64 {
	return max(g.RegularGas, g.StateGas)
}

func (g GasCosts) Underflow(b GasCosts) bool {
	stateOverflow := uint64(0)
	if b.StateGas > g.StateGas {
		stateOverflow = b.StateGas - g.StateGas
	}
	return b.RegularGas+stateOverflow > g.RegularGas
}

func (g *GasCosts) Sub(b GasCosts) {
	g.RegularGas -= b.RegularGas
	if b.StateGas > g.StateGas {
		diff := b.StateGas - g.StateGas
		g.StateGas = 0
		g.RegularGas -= diff
	} else {
		g.StateGas -= b.StateGas
	}
}

func (g *GasCosts) Add(b GasCosts) {
	g.RegularGas += b.RegularGas
	g.StateGas += b.StateGas
}
