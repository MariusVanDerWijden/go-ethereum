package catalyst

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	txfuzz "github.com/MariusVanDerWijden/tx-fuzz"
	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
)

func weirdHash(data *engine.ExecutableData, hashes ...common.Hash) common.Hash {
	rnd := rand.Int() % 10
	if data == nil {
		rnd += 6
	}
	switch rnd {
	case 1:
		return data.BlockHash
	case 2:
		return data.ParentHash
	case 3:
		return data.StateRoot
	case 4:
		return data.ReceiptsRoot
	case 5:
		return data.Random
	case 6:
		return common.Hash{}
	case 7:
		return hashes[rand.Int31n(int32(len(hashes)))]
	default:
		hash := hashes[rand.Int31n(int32(len(hashes)))]
		newBytes := hash.Bytes()
		index := rand.Int31n(int32(len(newBytes)))
		i := rand.Int31n(8)
		newBytes[index] = newBytes[index] ^ 1<<i
		return common.BytesToHash(newBytes)
	}
}

func weirdNumber(data *engine.ExecutableData, number uint64) uint64 {
	rnd := rand.Int()
	switch rnd % 7 {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return rand.Uint64()
	case 3:
		return ^uint64(0)
	case 4:
		return number + 1
	case 5:
		return number - 1
	default:
		return number + uint64(rand.Int63n(100000))
	}
}

func weirdByteSlice(data []byte) []byte {
	rnd := rand.Int()
	switch rnd % 4 {
	case 0:
		return make([]byte, 0)
	case 1:
		return make([]byte, 257)
	case 2:
		return []byte{1, 2}
	case 3:
		slice := make([]byte, len(data))
		rand.Read(slice)
		return slice
	default:
		return data
	}
}

func (api *ConsensusAPI) mutate(envelope *engine.ExecutionPayloadEnvelope, beaconRoot *common.Hash) *engine.ExecutionPayloadEnvelope {
	// mutate payload
	envelope.ExecutionPayload = api.mutatePayload(envelope.ExecutionPayload, beaconRoot)
	// mutate blobs
	envelope = api.mutateBlobs(envelope)
	// mutate basic fields, shouldn't do anything anyway
	envelope.BlockValue = big.NewInt(int64(weirdNumber(envelope.ExecutionPayload, envelope.BlockValue.Uint64())))
	if rand.Int()%2 == 0 {
		envelope.Override = !envelope.Override
	}
	return envelope
}

var hashCache = make(map[common.Hash]struct{})

func (api *ConsensusAPI) mutatePayload(data *engine.ExecutableData, beaconRoot *common.Hash) *engine.ExecutableData {
	var withdrawalsHash *common.Hash
	if data.Withdrawals != nil {
		h := types.DeriveSha(types.Withdrawals(data.Withdrawals), trie.NewStackTrie(nil))
		withdrawalsHash = &h
	}
	hashes := []common.Hash{
		data.ReceiptsRoot,
		data.StateRoot,
		data.BlockHash,
		data.ParentHash,
		api.eth.BlockChain().GetCanonicalHash(0),
		api.eth.BlockChain().GetCanonicalHash(data.Number - 255),
		api.eth.BlockChain().GetCanonicalHash(data.Number - 256),
		api.eth.BlockChain().GetCanonicalHash(data.Number - 257),
		api.eth.BlockChain().GetCanonicalHash(data.Number - 1000),
		api.eth.BlockChain().GetCanonicalHash(data.Number - 90001),
	}
	if withdrawalsHash != nil {
		hashes = append(hashes, *withdrawalsHash)
	}
	requests, requestsHash := api.mutateRequests(data, hashes)
	hashes = append(hashes, requestsHash)
	// cache the hashes
	for _, hash := range hashes {
		hashCache[hash] = struct{}{}
	}
	// add some of the hashes from cache
	if len(hashCache) > 0 {
		for i := 0; i < 4; i++ {
			iter := rand.Intn(len(hashCache))
			for hash := range hashCache {
				if iter == 0 {
					hashes = append(hashes, hash)
					break
				}
				iter--
			}
		}
	}

	bloom := types.BytesToBloom(data.LogsBloom)
	rnd := rand.Int()
	switch rnd % 19 {
	case 1:
		data.BlockHash = weirdHash(data, hashes...)
	case 2:
		data.ParentHash = weirdHash(data, hashes...)
	case 3:
		data.FeeRecipient = common.Address{}
	case 4:
		data.StateRoot = weirdHash(data, data.StateRoot)
	case 5:
		data.ReceiptsRoot = weirdHash(data, data.ReceiptsRoot)
	case 6:
		slice := weirdByteSlice(data.LogsBloom)
		if len(slice) > len(bloom) {
			bloom.SetBytes(slice[:len(bloom)])
		} else {
			bloom.SetBytes(slice)
		}
	case 7:
		data.Random = weirdHash(data, data.Random)
	case 8:
		data.Number = weirdNumber(data, data.Number)
	case 9:
		data.GasLimit = weirdNumber(data, data.GasLimit)
	case 10:
		data.GasUsed = weirdNumber(data, data.GasUsed)
	case 11:
		data.Timestamp = weirdNumber(data, data.Timestamp)
	case 12:
		hash := weirdHash(data, common.Hash{})
		data.ExtraData = hash[:]
	case 13:
		data.BaseFeePerGas = big.NewInt(int64(weirdNumber(data, data.BaseFeePerGas.Uint64())))
	case 14:
		data.BlockHash = weirdHash(data, data.BlockHash)
	case 15:
		num := weirdNumber(data, *data.ExcessBlobGas)
		data.ExcessBlobGas = &num
	case 16:
		h := weirdHash(data, *beaconRoot)
		beaconRoot = &h
	case 17:
		if withdrawalsHash != nil {
			h := weirdHash(data, *withdrawalsHash)
			withdrawalsHash = &h
		}
	case 18:
		h := weirdHash(data, requestsHash)
		requestsHash = h
	}
	if rand.Int()%10 != 0 {
		// Set correct blockhash in 90% of cases
		txs, _ := decodeTx(data.Transactions)
		txs, txhash := api.mutateTransactions(txs)
		number := big.NewInt(0)
		number.SetUint64(data.Number)
		header := &types.Header{
			ParentHash:       data.ParentHash,
			UncleHash:        types.EmptyUncleHash,
			Coinbase:         data.FeeRecipient,
			Root:             data.StateRoot,
			TxHash:           txhash,
			ReceiptHash:      data.ReceiptsRoot,
			Bloom:            bloom,
			Difficulty:       common.Big0,
			Number:           number,
			GasLimit:         data.GasLimit,
			GasUsed:          data.GasUsed,
			Time:             data.Timestamp,
			BaseFee:          data.BaseFeePerGas,
			Extra:            data.ExtraData,
			MixDigest:        data.Random,
			ExcessBlobGas:    data.ExcessBlobGas,
			ParentBeaconRoot: beaconRoot,
			BlobGasUsed:      data.BlobGasUsed,
			WithdrawalsHash:  withdrawalsHash,
			RequestsHash:     &requestsHash,
		}
		body := types.Body{Transactions: txs, Withdrawals: data.Withdrawals, Uncles: nil, Requests: requests}
		block := types.NewBlockWithHeader(header).WithBody(body)
		data.BlockHash = block.Hash()
	}
	return data
}

func decodeTx(enc [][]byte) ([]*types.Transaction, error) {
	var txs = make([]*types.Transaction, len(enc))
	for i, encTx := range enc {
		var tx types.Transaction
		if err := tx.UnmarshalBinary(encTx); err != nil {
			return nil, fmt.Errorf("invalid transaction %d: %v", i, err)
		}
		txs[i] = &tx
	}
	return txs, nil
}

// Used in tests to add a the list of transactions from a block to the tx pool.
func (api *ConsensusAPI) insertTransactions(txs types.Transactions) error {
	for _, tx := range txs {
		api.eth.TxPool().Add([]*types.Transaction{tx}, true, true)
	}
	return nil
}

func (api *ConsensusAPI) mutateTransactions(txs []*types.Transaction) ([]*types.Transaction, common.Hash) {
	chainid := api.eth.APIBackend.ChainConfig().ChainID
	txhash := types.DeriveSha(types.Transactions(txs), trie.NewStackTrie(nil))
	rnd := rand.Int()
	add := 0
	// if no txs are available, don't duplicate/modify any
	if len(txs) == 0 {
		add = 3
	}
	switch rnd%10 + add {
	case 1:
		// duplicate a txs
		tx := txs[rand.Intn(len(txs))]
		txs = append(txs, tx)
	case 2:
		// replace a tx
		index := rand.Intn(len(txs))
		b := make([]byte, 200)
		rand.Read(b)
		tx, err := txfuzz.RandomTx(filler.NewFiller(b))
		if err != nil {
			fmt.Println(err)
		}
		if rand.Int()%2 == 0 {
			txs[index] = tx
		} else {
			key := "0xaf5ead4413ff4b78bc94191a2926ae9ccbec86ce099d65aaf469e9eb1a0fa87f"
			sk := crypto.ToECDSAUnsafe(common.FromHex(key))
			signedTx, err := types.SignTx(tx, types.NewLondonSigner(tx.ChainId()), sk)
			if err != nil {
				break
			}
			txs[index] = signedTx
		}
	case 3:
		// Add a huuuge transaction
		gasLimit := uint64(7_800_000)
		code := []byte{0x60, 0x00, 0x60, 0x00, 0x60, 0x00, 0xf3}
		bigSlice := make([]byte, randomSize())
		code = append(code, bigSlice...)
		nonce, err := api.eth.APIBackend.GetPoolNonce(context.Background(), common.HexToAddress("0xb02A2EdA1b317FBd16760128836B0Ac59B560e9D"))
		if err != nil {
			fmt.Println(err)
		}
		gasPrice, err := api.eth.APIBackend.SuggestGasTipCap(context.Background())
		if err != nil {
			fmt.Println(err)
		}
		tx := types.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, code)

		key := "0xcdfbe6f7602f67a97602e3e9fc24cde1cdffa88acd47745c0b84c5ff55891e1b"
		sk := crypto.ToECDSAUnsafe(common.FromHex(key))
		signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainid), sk)
		if err != nil {
			panic(err)
		}
		txs = append(txs, signedTx)
	case 4:
		// add lots and lots of transactions
		rounds := rand.Int31n(1000)
		for i := 0; i < int(rounds); i++ {
			b := make([]byte, 200)
			rand.Read(b)
			tx, err := txfuzz.RandomTx(filler.NewFiller(b))
			if err != nil {
				fmt.Println(err)
			}

			key := "0xaf5ead4413ff4b78bc94191a2926ae9ccbec86ce099d65aaf469e9eb1a0fa87f"
			sk := crypto.ToECDSAUnsafe(common.FromHex(key))
			signedTx, err := types.SignTx(tx, types.NewLondonSigner(tx.ChainId()), sk)
			if err != nil {
				fmt.Println(err)
			}
			txs = append(txs, signedTx)
		}
	}

	if rand.Int()%20 > 17 {
		// Recompute correct txhash in most cases
		txhash = types.DeriveSha(types.Transactions(txs), trie.NewStackTrie(nil))
	}
	return txs, txhash
}

func (api *ConsensusAPI) mutateRequests(data *engine.ExecutableData, hashes []common.Hash) (types.Requests, common.Hash) {
	var requestsHash common.Hash
	var requests types.Requests
	if data.Deposits != nil {
		requests = make(types.Requests, 0)
		for _, d := range data.Deposits {
			requests = append(requests, types.NewRequest(d))
		}
	}
	if data.WithdrawalRequests != nil {
		for _, w := range data.WithdrawalRequests {
			requests = append(requests, types.NewRequest(w))
		}
	}
	if data.ConsolidationRequests != nil {
		requests = append(requests, data.ConsolidationRequests.Requests()...)
	}
	if requests != nil {
		h := types.DeriveSha(requests, trie.NewStackTrie(nil))
		requestsHash = h
	}

	rnd := rand.Int()
	switch rnd % 5 {
	case 0:
		// drop a request
		if len(requests) > 0 {
			index := rand.Intn(len(requests))
			requests = append(requests[index-1:], requests[index:]...)
		}
	case 1:
		// drop a request
		if len(requests) > 0 {
			index := rand.Intn(len(requests))
			requests[index] = nil
		}
	case 2:
		// replace with empty
		if len(requests) > 0 {
			index := rand.Intn(len(requests))
			requests[index] = &types.Request{}
		}
	case 3:
		// duplicate
		if len(requests) > 0 {
			index := rand.Intn(len(requests))
			requests = append(requests, requests[index])
		}
	case 4:
		// random deposit
		if len(requests) > 0 {
			index := rand.Intn(len(requests))
			var pk [48]byte
			rand.Read(pk[:])
			var sig [96]byte
			rand.Read(sig[:])
			requests[index] = types.NewRequest(&types.Deposit{
				PublicKey:             pk,
				WithdrawalCredentials: weirdHash(data, hashes...),
				Amount:                rand.Uint64(),
				Signature:             sig,
				Index:                 rand.Uint64(),
			})
		}
	case 5:
		// random withdrawal
		if len(requests) > 0 {
			index := rand.Intn(len(requests))
			var pk [48]byte
			rand.Read(pk[:])
			requests[index] = types.NewRequest(&types.WithdrawalRequest{
				PublicKey: pk,
				Source:    common.BytesToAddress(weirdHash(data, hashes...).Bytes()),
				Amount:    rand.Uint64(),
			})
		}
	case 6:
		// random consolidation
		if len(requests) > 0 {
			index := rand.Intn(len(requests))
			var pk [48]byte
			rand.Read(pk[:])
			var target [48]byte
			rand.Read(target[:])
			requests[index] = types.NewRequest(&types.ConsolidationRequest{
				SourcePublicKey: pk,
				Source:          common.BytesToAddress(weirdHash(data, hashes...).Bytes()),
				TargetPublicKey: target,
			})
		}
	case 7:
		// append a nil request
		requests = append(requests, types.NewRequest(&types.ConsolidationRequest{}))
	}

	return requests, requestsHash
}

func randomSize() int {
	rnd := rand.Int31n(100)
	if rnd < 5 {
		return int(rand.Int31n(11 * 1024 * 1024))
	} else if rnd < 10 {
		return 128*1024 + 1
	} else if rnd < 20 {
		return int(rand.Int31n(128 * 1024))
	}
	return int(rand.Int31n(127 * 1024))
}

func (api *ConsensusAPI) mutateBlobs(envelope *engine.ExecutionPayloadEnvelope) *engine.ExecutionPayloadEnvelope {
	// flips a random byte
	flipByte := func(data []hexutil.Bytes) []hexutil.Bytes {
		if len(data) == 0 {
			return data
		}
		element := rand.Intn(len(data))
		index := rand.Intn(len(data[element]))
		data[element][index] ^= data[element][index]
		return data
	}

	add := 0
	// if no txs are available, don't duplicate/modify any
	if len(envelope.BlobsBundle.Blobs) == 0 {
		add = 4
	}
	bundle := envelope.BlobsBundle
	switch rand.Int()%18 + add {
	case 0:
		// duplicate blob
		b := bundle.Blobs[rand.Intn(len(bundle.Blobs))]
		bundle.Blobs = append(bundle.Blobs, b)
	case 1:
		// zero blob elem
		blobIndex := rand.Intn(len(bundle.Blobs))
		elemIndex := rand.Intn(params.BlobTxBytesPerFieldElement) - 32
		var elem [32]byte
		copy(bundle.Blobs[blobIndex][elemIndex:], elem[:])
	case 2:
		// random blob elem
		blobIndex := rand.Intn(len(bundle.Blobs))
		elemIndex := rand.Intn(params.BlobTxBytesPerFieldElement) - 32
		var elem [32]byte
		rand.Read(elem[:])
		copy(bundle.Blobs[blobIndex][elemIndex:], elem[:])
	case 3:
		// one blob elem
		blobIndex := rand.Intn(len(bundle.Blobs))
		elemIndex := rand.Intn(params.BlobTxBytesPerFieldElement) - 32
		var elem [32]byte
		elem[31] = 1
		copy(bundle.Blobs[blobIndex][elemIndex:], elem[:])
	case 4:
		// append empty blob
		b := kzg4844.Blob{}
		bundle.Blobs = append(bundle.Blobs, hexutil.Bytes(b[:]))
	case 5:
		// set blockhash
		envelope.ExecutionPayload.BlockHash = weirdHash(nil, envelope.ExecutionPayload.BlockHash)
	case 6:
		// drop blobs
		bundle.Blobs = make([]hexutil.Bytes, 0)
	case 7:
		// all empty blobs
		bundle.Blobs = make([]hexutil.Bytes, len(bundle.Blobs))
	case 8:
		// all empty commitments
		bundle.Commitments = make([]hexutil.Bytes, len(bundle.Commitments))
	case 9:
		// drop commitments
		bundle.Commitments = make([]hexutil.Bytes, 0)
	case 10:
		// duplicate commitment
		if len(bundle.Commitments) == 0 {
			break
		}
		b := bundle.Commitments[rand.Intn(len(bundle.Commitments))]
		bundle.Commitments = append(bundle.Commitments, b)
	case 11:
		// append empty commitment
		b := kzg4844.Commitment{}
		bundle.Commitments = append(bundle.Commitments, hexutil.Bytes(b[:]))
	case 12:
		// append random commitment
		var b [48]byte
		rand.Read(b[:])
		bundle.Commitments = append(bundle.Commitments, hexutil.Bytes(b[:]))
	case 13:
		// replace empty commitment
		if len(bundle.Commitments) == 0 {
			break
		}
		index := rand.Intn(len(bundle.Commitments))
		var b [48]byte
		bundle.Commitments[index] = hexutil.Bytes(b[:])
	case 14:
		// replace 1 commitment
		if len(bundle.Commitments) == 0 {
			break
		}
		var b [48]byte
		b[41] = 1
		index := rand.Intn(len(bundle.Commitments))
		bundle.Commitments[index] = hexutil.Bytes(b[:])
	case 15:
		bundle.Blobs = flipByte(bundle.Blobs)
	case 16:
		bundle.Commitments = flipByte(bundle.Commitments)
	case 17:
		bundle.Proofs = flipByte(bundle.Proofs)
	}

	return envelope
}
