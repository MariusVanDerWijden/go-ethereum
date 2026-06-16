// Copyright 2026 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package engineapi

import (
	"net/http"

	sszt "github.com/ethereum/go-ethereum/beacon/engine/ssz"
	"github.com/ethereum/go-ethereum/params/forks"
	"github.com/karalabe/ssz"
)

// resolveFork maps the URL fork onto the ssz.Fork the codec multiplexes on,
// enforcing that it is at least min. On failure it writes the appropriate
// problem response and returns ok=false; the caller should return immediately.
func resolveFork(w http.ResponseWriter, fork, min forks.Fork) (ssz.Fork, bool) {
	if fork < min {
		writeProblem(w, http.StatusBadRequest, ErrUnsupportedFork, "")
		return ssz.ForkUnknown, false
	}
	sf, ok := sszt.ForkFor(fork)
	if !ok {
		writeProblem(w, http.StatusBadRequest, ErrUnsupportedFork, "")
		return ssz.ForkUnknown, false
	}
	return sf, true
}

// baseEngineFork collapses Blob-Parameter-Only forks (BPO1..BPO5) onto their
// base protocol fork (Osaka), which is the one with a fork-scoped engine URL.
func baseEngineFork(f forks.Fork) forks.Fork {
	switch f {
	case forks.BPO1, forks.BPO2, forks.BPO3, forks.BPO4, forks.BPO5:
		return forks.Osaka
	default:
		return f
	}
}

// eraForks returns the fork plus the BPO forks layered on it, for validating a
// built payload's timestamp-derived fork (which LatestFork may report as a BPO).
func eraForks(fork forks.Fork) []forks.Fork {
	if fork == forks.Osaka {
		return []forks.Fork{forks.Osaka, forks.BPO1, forks.BPO2, forks.BPO3, forks.BPO4, forks.BPO5}
	}
	return []forks.Fork{fork}
}
