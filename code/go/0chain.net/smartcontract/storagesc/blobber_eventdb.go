package storagesc

import (
	cstate "0chain.net/chaincore/chain/state"
	"0chain.net/smartcontract/dbs"
	"0chain.net/smartcontract/dbs/event"
)

func emitUpdateBlobber(sn *StorageNode, sp *stakePool, balances cstate.StateContextI) error {
	staked, err := sp.stake()
	if err != nil {
		return err
	}
	b := sn.mustBase()
	data := &event.Blobber{
		BaseURL:    b.BaseURL,
		ReadPrice:  b.Terms.ReadPrice,
		WritePrice: b.Terms.WritePrice,

		Capacity:     b.Capacity,
		Allocated:    b.Allocated,
		SavedData:    b.SavedData,
		NotAvailable: b.NotAvailable,
		// IsRestricted: *sn.IsRestricted,
		Provider: event.Provider{
			ID:              b.ID,
			DelegateWallet:  b.StakePoolSettings.DelegateWallet,
			NumDelegates:    b.StakePoolSettings.MaxNumDelegates,
			ServiceCharge:   b.StakePoolSettings.ServiceChargeRatio,
			LastHealthCheck: b.LastHealthCheck,
			TotalStake:      staked,
		},
		OffersTotal: sp.TotalOffers,
	}

	if v2, ok := sn.Entity().(*storageNodeV2); ok && v2.IsRestricted != nil {
		data.IsRestricted = *v2.IsRestricted
	}

	balances.EmitEvent(event.TypeStats, event.TagUpdateBlobber, b.ID, data)
	return nil
}

func emitAddBlobber(sn *StorageNode, sp *stakePool, balances cstate.StateContextI) error {
	staked, err := sp.stake()
	if err != nil {
		return err
	}
	b := sn.mustBase()

	data := &event.Blobber{
		BaseURL:    b.BaseURL,
		ReadPrice:  b.Terms.ReadPrice,
		WritePrice: b.Terms.WritePrice,

		Capacity:     b.Capacity,
		Allocated:    b.Allocated,
		SavedData:    b.SavedData,
		NotAvailable: false,
		Provider: event.Provider{
			ID:              b.ID,
			DelegateWallet:  b.StakePoolSettings.DelegateWallet,
			NumDelegates:    b.StakePoolSettings.MaxNumDelegates,
			ServiceCharge:   b.StakePoolSettings.ServiceChargeRatio,
			LastHealthCheck: b.LastHealthCheck,
			TotalStake:      staked,
			Rewards: event.ProviderRewards{
				ProviderID:   b.ID,
				Rewards:      sp.Reward,
				TotalRewards: sp.Reward,
			},
		},

		OffersTotal: sp.TotalOffers,

		CreationRound: balances.GetBlock().Round,
	}

	if v2, ok := sn.Entity().(*storageNodeV2); ok {
		if v2.IsRestricted != nil {
			data.IsRestricted = *v2.IsRestricted
		}
	}

	balances.EmitEvent(event.TypeStats, event.TagAddBlobber, b.ID, data)
	return nil
}

func emitUpdateBlobberAllocatedSavedHealth(sn *StorageNode, balances cstate.StateContextI) {
	b := sn.mustBase()
	balances.EmitEvent(event.TypeStats, event.TagUpdateBlobberAllocatedSavedHealth, b.ID, event.Blobber{
		Provider: event.Provider{
			ID:              b.ID,
			LastHealthCheck: b.LastHealthCheck,
		},
		Allocated: b.Allocated,
		SavedData: b.SavedData,
	})
}

func emitBlobberHealthCheck(sn *StorageNode, downtime uint64, balances cstate.StateContextI) {
	b := sn.mustBase()
	data := dbs.DbHealthCheck{
		ID:              b.ID,
		LastHealthCheck: b.LastHealthCheck,
		Downtime:        downtime,
	}

	balances.EmitEvent(event.TypeStats, event.TagBlobberHealthCheck, b.ID, data)
}
