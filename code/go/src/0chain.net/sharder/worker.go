package sharder

import (
	"context"
	"log"
	"os"
	"time"

	"0chain.net/blockstore"

	"0chain.net/node"

	"0chain.net/common"
	"0chain.net/config"
	"0chain.net/datastore"
	. "0chain.net/logging"
	"0chain.net/round"
	metrics "github.com/rcrowley/go-metrics"
	"go.uber.org/zap"
)

/*SetupWorkers - setup the background workers */
func SetupWorkers() {
	ClearWorkerState()
	ctx := common.GetRootContext()
	go GetSharderChain().BlockWorker(ctx)
	go GetSharderChain().BlockStorageWorker(ctx)
	if config.Development() {
		go metrics.LogScaled(metrics.DefaultRegistry, 60*time.Second, time.Millisecond, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}
}

/*ClearWorkerState - clears the worker state */
func ClearWorkerState() {
}

/*BlockWorker - stores the blocks */
func (sc *Chain) BlockWorker(ctx context.Context) {
	for true {
		select {
		case <-ctx.Done():
			return
		case b := <-sc.GetBlockChannel():
			eb, err := sc.GetBlock(ctx, b.Hash)
			if eb != nil {
				if err == nil {
					Logger.Debug("block already received", zap.Any("round", b.Round), zap.Any("block", b.Hash))
					continue
				} else {
					Logger.Error("get block", zap.Any("block", b.Hash), zap.Error(err))
				}
			}
			if err := sc.VerifyNotarization(ctx, b, b.VerificationTickets); err != nil {
				Logger.Error("notarization verification failed", zap.Int64("round", b.Round), zap.String("block", b.Hash), zap.Error(err))
				continue
			}
			if err := b.Validate(ctx); err != nil {
				Logger.Error("block validation", zap.Any("round", b.Round), zap.Any("hash", b.Hash), zap.Error(err))
				continue
			}
			sc.AddBlock(b)
			er := sc.GetRound(b.Round)
			if er != nil {
				nb := er.GetNotarizedBlocks()
				if len(nb) > 0 {
					Logger.Error("*** different blocks for the same round ***", zap.Any("round", b.Round), zap.Any("block", b.Hash), zap.Any("existing_block", nb[0].Hash))
				}
			} else {
				er = datastore.GetEntityMetadata("round").Instance().(*round.Round)
				er.Number = b.Round
				er.RandomSeed = b.RoundRandomSeed
				sc.AddRound(er)
			}
			er.AddNotarizedBlock(b)
			pr := sc.GetRound(er.Number - 1)
			if pr != nil {
				go sc.FinalizeRound(ctx, pr, sc)
			}
		}
	}
}

func (sc *Chain) BlockStorageWorker(ctx context.Context) {
	for true {
		select {
		case <-ctx.Done():
			return
		case r := <-sc.GetRoundChannel():
			b, err := sc.GetBlockFromHash(ctx, r.BlockHash, r.Number)
			if err != nil {
				Logger.Error("failed to get block", zap.String("blockhash", r.BlockHash), zap.Error(err))
			}
			self := node.GetSelfNode(ctx)
			if sc.CanStoreBlock(r, self.Node) {
				if b != nil {
					var sTxns = make([]datastore.Entity, len(b.Txns))
					for idx, txn := range b.Txns {
						txnSummary := txn.GetSummary()
						txnSummary.BlockHash = b.Hash
						sTxns[idx] = txnSummary
					}
					err = sc.StoreTransactions(ctx, sTxns)
					if err != nil {
						Logger.Error("db error (save transaction)", zap.String("block", b.Hash), zap.Error(err))
					} else {
						err = sc.StoreBlock(ctx, b)
						if err != nil {
							Logger.Error("db error (save block)", zap.Any("round", b.Round), zap.String("block", b.Hash), zap.Error(err))
						}
					}
				}
			} else {
				err = blockstore.GetStore().DeleteBlock(b)
				if err != nil {
					Logger.Error("failed to delete block from file system", zap.String("blockhash", b.Hash), zap.Error(err))
				}
			}
		}
	}
}
