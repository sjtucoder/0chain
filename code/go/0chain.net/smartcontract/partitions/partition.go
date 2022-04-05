package partitions

import (
	"0chain.net/core/common"
	"math/rand"

	"0chain.net/core/datastore"

	"0chain.net/core/util"

	"0chain.net/chaincore/chain/state"
)

type PartitionItem interface {
	util.Serializable
	Name() string
	Data() string
}

type PartitionItemList interface {
	util.Serializable
	add(it PartitionItem) error
	remove(item PartitionItem) error
	cutTail() PartitionItem
	changed() bool
	length() int
	itemRange(start, end int) []PartitionItem
	save(balances state.StateContextI) error
	get(key datastore.Key, balances state.StateContextI) error
	update(it PartitionItem) error
}

type ChangePartitionCallback = func(PartitionItem, int, int, state.StateContextI) error

type Partition interface {
	util.Serializable
	Add(PartitionItem, state.StateContextI) (int, error)
	Remove(PartitionItem, int, state.StateContextI) error

	SetCallback(ChangePartitionCallback)
	Size(state.StateContextI) (int, error)
	Save(state.StateContextI) error
	UpdateItem(partIndex int, it PartitionItem, balances state.StateContextI) error
	GetItem(partIndex int, itemName string, balances state.StateContextI) (PartitionItem, error)
}

type RandPartition interface {
	Partition
	AddRand(PartitionItem, *rand.Rand, state.StateContextI) (int, error)
	GetRandomSlice(*rand.Rand, state.StateContextI) ([]PartitionItem, error)
}

//go:generate msgp -io=false -tests=false -unexported=true -v

type PartitionLocation struct {
	Location  int
	Timestamp common.Timestamp
}

func NewPartitionLocation(location int, timestamp common.Timestamp) *PartitionLocation {
	pl := new(PartitionLocation)
	pl.Location = location
	pl.Timestamp = timestamp

	return pl
}
