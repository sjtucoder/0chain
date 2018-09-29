package sharder

import (
	"context"
	"net/http"

	"0chain.net/block"
	"0chain.net/common"
	"0chain.net/datastore"
	"go.uber.org/zap"
	. "0chain.net/logging"
	"0chain.net/node"
	"0chain.net/persistencestore"
)

/*SetupM2SReceivers - setup handlers for all the messages received from the miner */
func SetupM2SReceivers() {
	http.HandleFunc("/v1/_m2s/block/finalized", node.ToN2NReceiveEntityHandler(persistencestore.WithConnectionEntityJSONHandler(FinalizedBlockHandler, datastore.GetEntityMetadata("block"))))
	http.HandleFunc("/v1/_m2s/block/notarized", node.ToN2NReceiveEntityHandler(persistencestore.WithConnectionEntityJSONHandler(NotarizedBlockHandler, datastore.GetEntityMetadata("block"))))
}

/*SetupM2SResponders - setup handlers for all the requests from the miner */
func SetupM2SResponders() {
	http.HandleFunc("/v1/_m2s/block/latest_finalized/get", node.ToN2NSendEntityHandler(LatestFinalizedBlockHandler))
}

/*FinalizedBlockHandler - handle the finalized block */
func FinalizedBlockHandler(ctx context.Context, entity datastore.Entity) (interface{}, error) {
	return NotarizedBlockHandler(ctx, entity)
}

/*NotarizedBlockHandler - handle the notarized block */
func NotarizedBlockHandler(ctx context.Context, entity datastore.Entity) (interface{}, error) {
	b, ok := entity.(*block.Block)
	if !ok {
		return nil, common.InvalidRequest("Invalid Entity")
	}
	sc := GetSharderChain()
	if b.Round < sc.LatestFinalizedBlock.Round {
		return true, nil
	}
	sc.GetBlockChannel() <- b
	return true, nil
}

/*LatestFinalizedBlockHandler - handle latest finalized block*/
func LatestFinalizedBlockHandler(ctx context.Context, r *http.Request) (interface{}, error) {
	nodeIndex := node.Self.SetIndex
	sc := GetSharderChain()
	lfb := sc.LatestFinalizedBlock
	if nodeIndex%3 == 0 {
		// send the right block
		return lfb, nil
	}

	if nodeIndex%3 == 1 {
		// send old block
		roundNumber := lfb.Round - 10
		r1, err := sc.GetRoundFromStore(ctx, roundNumber)
		if err == nil {
			return r1.Block, nil
		} else {
			Logger.Error("could not retrieve round from store", zap.Int64("round", roundNumber))
			return nil, err
		}
	}

	if nodeIndex%3 == 2 {
		//send corrupt block
		b := lfb
		b.CreationDate = common.Now()
		return b, nil
	}
	return lfb, nil
}
