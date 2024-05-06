package live

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
)

func init() {
	tracers.LiveDirectory.Register("address", newAddressTracer)
}

// addressTracer is a live tracer that tracks all transactions that an address was involved in.
type addressTracer struct {
	addresses []common.Address
	writer    map[common.Address]*csv.Writer

	currentHash   common.Hash
	currentSender common.Address
}

type addressTracerConfig struct {
	Path      string   `json:"path"`      // Path to the directory where the tracer logs will be stored
	Addresses []string `json:"addresses"` // Addresses to be watched
}

func newAddressTracer(cfg json.RawMessage) (*tracing.Hooks, error) {
	var config addressTracerConfig
	if cfg != nil {
		if err := json.Unmarshal(cfg, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config: %v", err)
		}
		if config.Path == "" {
			return nil, fmt.Errorf("no path provided")
		}
	}

	var (
		addresses []common.Address
		writer    = make(map[common.Address]*csv.Writer)
	)
	for _, addr := range config.Addresses {
		a := common.HexToAddress(addr)
		addresses = append(addresses, a)
		file, err := os.Create(fmt.Sprintf("%v/%v.csv", config.Path, addr))
		if err != nil {
			return nil, err
		}
		writer[a] = csv.NewWriter(file)
		writer[a].Write([]string{
			"TxHash",
			"Sender",
			"Currency",
			"Previous",
			"New",
			"Reason",
		})
		writer[a].Flush()
	}

	t := &addressTracer{
		addresses: addresses,
		writer:    writer,
	}
	return &tracing.Hooks{
		OnTxStart:       t.OnTxStart,
		OnBalanceChange: t.OnBalanceChange,
		OnStorageChange: t.OnStorageChange,
	}, nil
}

func (t *addressTracer) OnTxStart(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
	t.currentHash = tx.Hash()
	t.currentSender = from
}

func (t *addressTracer) OnBalanceChange(a common.Address, prev, new *big.Int, reason tracing.BalanceChangeReason) {
	for _, addr := range t.addresses {
		if addr == a {
			t.writeRecord(addr, "ETH", prev.String(), new.String(), fmt.Sprint(reason))
			break
		}
	}
}

var (
	transferTopic   = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	transferType, _ = abi.NewType("tuple(address,address,uint256)", "", nil)
)

func (t *addressTracer) OnLog(l *types.Log) {
	for _, topic := range l.Topics {
		if topic == transferTopic {
			unpacked, err := (abi.Arguments{{Type: transferType}}).Unpack(l.Data[4:])
			if err != nil {
				continue
			}
			from, ok := unpacked[0].(common.Address)
			if !ok {
				continue
			}
			to, ok := unpacked[1].(common.Address)
			if !ok {
				continue
			}
			tokens, ok := unpacked[2].(*big.Int)
			if !ok {
				continue
			}
			for _, addr := range t.addresses {
				if addr == from || addr == to {
					t.writeRecord(addr, l.Address.Hex(), "", tokens.String(), "token transfer")
					break
				}
			}
		}
	}
}

func (t *addressTracer) OnStorageChange(a common.Address, k, prev, new common.Hash) {
	slot := common.BytesToAddress(k.Bytes()[12:])
	for _, addr := range t.addresses {
		if addr == slot {
			t.writeRecord(addr, a.Hex(), prev.String(), new.String(), "token transfer")
			break
		}
	}
}

func (t *addressTracer) writeRecord(addr common.Address, currency, previous, new, reason string) {
	t.writer[addr].Write([]string{
		t.currentHash.Hex(),
		t.currentSender.Hex(),
		currency,
		previous,
		new,
		reason,
	})
	t.writer[addr].Flush()
}
