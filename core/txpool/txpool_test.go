package txpool_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/consensus/misc/eip4844"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/txpool/blobpool"
	"github.com/ethereum/go-ethereum/core/txpool/legacypool"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var testChainConfig *params.ChainConfig

func init() {
	testChainConfig = new(params.ChainConfig)
	*testChainConfig = *params.MainnetChainConfig

	testChainConfig.CancunTime = new(uint64)
	*testChainConfig.CancunTime = uint64(time.Now().Unix())
}

// testBlockChain is a mock of the live chain for testing the pool.
type testBlockChain struct {
	config        *params.ChainConfig
	basefee       *uint256.Int
	blobfee       *uint256.Int
	statedb       *state.StateDB
	chainHeadFeed *event.Feed
}

func (bc *testBlockChain) Config() *params.ChainConfig {
	return bc.config
}

func (bc *testBlockChain) CurrentBlock() *types.Header {
	// Yolo, life is too short to invert mist.CalcBaseFee and misc.CalcBlobFee,
	// just binary search it them.

	// The base fee at 5714 ETH translates into the 21000 base gas higher than
	// mainnet ether existence, use that as a cap for the tests.
	var (
		blockNumber = new(big.Int).Add(bc.config.LondonBlock, big.NewInt(1))
		blockTime   = *bc.config.CancunTime + 1
		gasLimit    = uint64(30_000_000)
	)
	lo := new(big.Int)
	hi := new(big.Int).Mul(big.NewInt(5714), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	for new(big.Int).Add(lo, big.NewInt(1)).Cmp(hi) != 0 {
		mid := new(big.Int).Add(lo, hi)
		mid.Div(mid, big.NewInt(2))

		if eip1559.CalcBaseFee(bc.config, &types.Header{
			Number:   blockNumber,
			GasLimit: gasLimit,
			GasUsed:  0,
			BaseFee:  mid,
		}).Cmp(bc.basefee.ToBig()) > 0 {
			hi = mid
		} else {
			lo = mid
		}
	}
	baseFee := lo

	// The excess blob gas at 2^27 translates into a blob fee higher than mainnet
	// ether existence, use that as a cap for the tests.
	lo = new(big.Int)
	hi = new(big.Int).Exp(big.NewInt(2), big.NewInt(27), nil)

	for new(big.Int).Add(lo, big.NewInt(1)).Cmp(hi) != 0 {
		mid := new(big.Int).Add(lo, hi)
		mid.Div(mid, big.NewInt(2))

		if eip4844.CalcBlobFee(mid.Uint64()).Cmp(bc.blobfee.ToBig()) > 0 {
			hi = mid
		} else {
			lo = mid
		}
	}
	excessBlobGas := lo.Uint64()

	return &types.Header{
		Number:        blockNumber,
		Time:          blockTime,
		GasLimit:      gasLimit,
		BaseFee:       baseFee,
		ExcessBlobGas: &excessBlobGas,
	}
}

func (bc *testBlockChain) StateAt(common.Hash) (*state.StateDB, error) {
	return bc.statedb, nil
}

func (bc *testBlockChain) CurrentFinalBlock() *types.Header {
	return &types.Header{
		Number: big.NewInt(0),
	}
}

func (bc *testBlockChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return nil
}

func (bc *testBlockChain) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return bc.chainHeadFeed.Subscribe(ch)
}

func BenchmarkGetNonce(b *testing.B) {

	statedb, _ := state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewDatabase(memorydb.New())), nil)
	chain := &testBlockChain{
		config:        testChainConfig,
		basefee:       uint256.NewInt(0),
		blobfee:       uint256.NewInt(0),
		statedb:       statedb,
		chainHeadFeed: new(event.Feed),
	}

	blobPool := blobpool.New(blobpool.DefaultConfig, chain)
	legacyPool := legacypool.New(legacypool.DefaultConfig, chain)
	pool, err := txpool.New(0, chain, []txpool.SubPool{legacyPool, blobPool})
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		pool.Nonce(common.Address{0x01})
	}
}
