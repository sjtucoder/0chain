package storagesc

import (
	"fmt"
	"time"

	"0chain.net/core/config"
	"github.com/0chain/common/core/currency"

	chainState "0chain.net/chaincore/chain/state"
	"0chain.net/core/datastore"
	"0chain.net/core/encryption"
	"github.com/0chain/common/core/statecache"
	"github.com/0chain/common/core/util"
)

//go:generate msgp -io=false -tests=false -unexported=true -v

func scConfigKey(scKey string) datastore.Key {
	return scKey + encryption.Hash("storagesc_config")
}

type freeAllocationSettings struct {
	DataShards       int        `json:"data_shards"`
	ParityShards     int        `json:"parity_shards"`
	Size             int64      `json:"size"`
	ReadPriceRange   PriceRange `json:"read_price_range"`
	WritePriceRange  PriceRange `json:"write_price_range"`
	ReadPoolFraction float64    `json:"read_pool_fraction"`
}

type stakePoolConfig struct {
	MinLockPeriod time.Duration `json:"min_lock_period"`
	KillSlash     float64       `json:"kill_slash"`
}

type readPoolConfig struct {
	MinLock currency.Coin `json:"min_lock"`
}

type writePoolConfig struct {
	MinLock currency.Coin `json:"min_lock"`
}

type blockReward struct {
	BlockReward             currency.Coin    `json:"block_reward"`
	BlockRewardChangePeriod int64            `json:"block_reward_change_period"`
	BlockRewardChangeRatio  float64          `json:"block_reward_change_ratio"`
	QualifyingStake         currency.Coin    `json:"qualifying_stake"`
	TriggerPeriod           int64            `json:"trigger_period"`
	Gamma                   blockRewardGamma `json:"gamma"`
	Zeta                    blockRewardZeta  `json:"zeta"`
}

type blockRewardGamma struct {
	Alpha float64 `json:"alpha"`
	A     float64 `json:"a"`
	B     float64 `json:"b"`
}

type blockRewardZeta struct {
	I  float64 `json:"i"`
	K  float64 `json:"k"`
	Mu float64 `json:"mu"`
}

func newConfig() *Config {
	return &Config{
		ReadPool:               &readPoolConfig{},
		WritePool:              &writePoolConfig{},
		StakePool:              &stakePoolConfig{},
		FreeAllocationSettings: freeAllocationSettings{},
		BlockReward:            &blockReward{},
		Cost:                   make(map[string]int),
	}
}

// Config represents SC configurations ('storagesc:' from sc.yaml).
type Config struct {
	// TimeUnit is a duration used as divider for a write price. A write price
	// measured in tok / GB / time unit. Where the time unit is this
	// configuration.
	TimeUnit time.Duration `json:"time_unit"`
	// Minted tokens by entire SC.
	Minted currency.Coin `json:"minted"`
	// MinAllocSize is minimum possible size (bytes)
	// of an allocation the SC accept.
	MinAllocSize int64 `json:"min_alloc_size"`
	// MaxChallengeCompletionRounds is max time to complete a challenge.
	MaxChallengeCompletionRounds int64 `json:"max_challenge_completion_rounds"`
	// MinBlobberCapacity allowed to register in the SC.
	MinBlobberCapacity int64 `json:"min_blobber_capacity"`
	// ReadPool related configurations.
	ReadPool *readPoolConfig `json:"readpool"`
	// WritePool related configurations.
	WritePool *writePoolConfig `json:"write_pool"`
	// StakePool related configurations.
	StakePool *stakePoolConfig `json:"stakepool"`
	// ValidatorReward represents % (value in [0; 1] range) of blobbers' reward
	// goes to validators. Even if a blobber doesn't pass a challenge validators
	// receive this reward.
	ValidatorReward float64 `json:"validator_reward"`
	// BlobberSlash represents % (value in [0; 1] range) of blobbers' stake
	// tokens penalized on challenge not passed.
	BlobberSlash      float64       `json:"blobber_slash"`
	HealthCheckPeriod time.Duration `json:"health_check_period"`
	// MaxBlobbersPerAllocation maximum blobbers that can be sent per allocation
	MaxBlobbersPerAllocation int `json:"max_blobbers_per_allocation"`

	// price limits for blobbers

	// MaxReadPrice allowed for a blobber.
	MaxReadPrice currency.Coin `json:"max_read_price"`
	// MaxWritePrice
	MaxWritePrice currency.Coin `json:"max_write_price"`
	MinWritePrice currency.Coin `json:"min_write_price"`
	MaxFileSize   int64         `json:"max_file_size"`

	// allocation cancellation
	CancellationCharge float64 `json:"cancellation_charge"`
	// free allocations
	MaxTotalFreeAllocation      currency.Coin          `json:"max_total_free_allocation"`
	MaxIndividualFreeAllocation currency.Coin          `json:"max_individual_free_allocation"`
	FreeAllocationSettings      freeAllocationSettings `json:"free_allocation_settings"`

	// challenges generating

	// ChallengeEnabled is challenges generating pin.
	ChallengeEnabled       bool  `json:"challenge_enabled"`
	ChallengeGenerationGap int64 `json:"challenge_generation_gap"`
	// ValidatorsPerChallenge is the number of validators to select per
	// challenges.
	ValidatorsPerChallenge       int `json:"validators_per_challenge"`
	NumValidatorsRewarded        int `json:"num_validators_rewarded"`
	MaxBlobberSelectForChallenge int `json:"max_blobber_select_for_challenge"`

	// MinStake allowed by a blobber/validator (entire SC boundary).
	MinStake currency.Coin `json:"min_stake"`
	// MaxStake allowed by a blobber/validator (entire SC boundary).
	MaxStake            currency.Coin `json:"max_stake"`
	MinStakePerDelegate currency.Coin `json:"min_stake_per_delegate"`

	// MaxDelegates per stake pool
	MaxDelegates int `json:"max_delegates"`

	// MaxCharge that blobber gets from rewards to its delegate_wallet.
	MaxCharge float64 `json:"max_charge"`

	BlockReward *blockReward `json:"block_reward"`

	OwnerId string         `json:"owner_id"`
	Cost    map[string]int `json:"cost"`
}

func (conf *Config) validate() (err error) {
	if conf.TimeUnit <= 1*time.Second {
		return fmt.Errorf("time_unit less than 1s: %v", conf.TimeUnit)
	}
	if conf.ValidatorReward < 0.0 || 1.0 < conf.ValidatorReward {
		return fmt.Errorf("validator_reward not in [0; 1] range: %v",
			conf.ValidatorReward)
	}
	if conf.BlobberSlash < 0.0 || 1.0 < conf.BlobberSlash {
		return fmt.Errorf("blobber_slash not in [0; 1] range: %v",
			conf.BlobberSlash)
	}
	if conf.CancellationCharge < 0.0 || 1.0 < conf.CancellationCharge {
		return fmt.Errorf("cancellation_charge not in [0, 1] range: %v",
			conf.CancellationCharge)
	}
	if conf.MaxBlobbersPerAllocation <= 0 {
		return fmt.Errorf("invalid max_blobber_per_allocation <= 0: %v",
			conf.MaxBlobbersPerAllocation)
	}
	if conf.MinBlobberCapacity < 0 {
		return fmt.Errorf("negative min_blobber_capacity: %v",
			conf.MinBlobberCapacity)
	}
	if conf.MaxChallengeCompletionRounds < 0 {
		return fmt.Errorf("negative max_challenge_completion_rounds: %v",
			conf.MaxChallengeCompletionRounds)
	}
	if conf.HealthCheckPeriod <= 0 {
		return fmt.Errorf("non-positive health check period: %v", conf.HealthCheckPeriod)
	}
	if conf.MinAllocSize < 0 {
		return fmt.Errorf("negative min_alloc_size: %v", conf.MinAllocSize)
	}

	if conf.MaxWritePrice < conf.MinWritePrice {
		return fmt.Errorf("max wirte price %v must be more than min_write_price: %v",
			conf.MaxWritePrice, conf.MinWritePrice)
	}
	if conf.StakePool.KillSlash < 0 || conf.StakePool.KillSlash > 1 {
		return fmt.Errorf("stakepool.kill_slash, %v must be in interval [0.1]", conf.StakePool.KillSlash)
	}

	if conf.FreeAllocationSettings.DataShards < 0 {
		return fmt.Errorf("negative free_allocation_settings.data_shards: %v",
			conf.FreeAllocationSettings.DataShards)
	}
	if conf.FreeAllocationSettings.ParityShards < 0 {
		return fmt.Errorf("negative free_allocation_settings.parity_shards: %v",
			conf.FreeAllocationSettings.ParityShards)
	}
	if conf.FreeAllocationSettings.Size < 0 {
		return fmt.Errorf("negative free_allocation_settings.size: %v",
			conf.FreeAllocationSettings.Size)
	}
	if !conf.FreeAllocationSettings.ReadPriceRange.isValid() {
		return fmt.Errorf("invalid free_allocation_settings.read_price_range: %v",
			conf.FreeAllocationSettings.ReadPriceRange)
	}
	if !conf.FreeAllocationSettings.WritePriceRange.isValid() {
		return fmt.Errorf("invalid free_allocation_settings.write_price_range: %v",
			conf.FreeAllocationSettings.WritePriceRange)
	}
	if conf.FreeAllocationSettings.ReadPoolFraction < 0 || 1 < conf.FreeAllocationSettings.ReadPoolFraction {
		return fmt.Errorf("free_allocation_settings.free_read_pool must be in [0,1]: %v",
			conf.FreeAllocationSettings.ReadPoolFraction)
	}
	if conf.ValidatorsPerChallenge <= 0 {
		return fmt.Errorf("invalid validators_per_challenge <= 0: %v",
			conf.ValidatorsPerChallenge)
	}
	if conf.NumValidatorsRewarded <= 0 {
		return fmt.Errorf("invalid num_validators_rewarded <= 0: %v",
			conf.NumValidatorsRewarded)
	}
	if conf.MaxBlobberSelectForChallenge <= 0 {
		return fmt.Errorf("invalid max_blobber_select_for_challenge <= 0: %v",
			conf.MaxBlobberSelectForChallenge)
	}
	if conf.MaxStake < conf.MinStake {
		return fmt.Errorf("max_stake less than min_stake: %v < %v", conf.MinStake,
			conf.MaxStake)
	}
	if conf.MaxDelegates < 1 {
		return fmt.Errorf("max_delegates is too small %v", conf.MaxDelegates)
	}
	if conf.MaxCharge < 0 {
		return fmt.Errorf("negative max_charge: %v", conf.MaxCharge)
	}
	if conf.MaxCharge > 1.0 {
		return fmt.Errorf("max_change >= 1.0 (> 100%%, invalid): %v",
			conf.MaxCharge)
	}

	if len(conf.OwnerId) == 0 {
		return fmt.Errorf("owner_id does not set or empty")
	}

	if conf.BlockReward.Gamma.A <= 0 {
		return fmt.Errorf("invalid block_reward.gamma.a <= 0: %v", conf.BlockReward.Gamma.A)
	}
	if conf.BlockReward.Gamma.B <= 0 {
		return fmt.Errorf("invalid block_reward.gamma.b <= 0: %v", conf.BlockReward.Gamma.B)
	}
	if conf.BlockReward.Gamma.Alpha <= 0 {
		return fmt.Errorf("invalid block_reward.gamma.alpha <= 0: %v", conf.BlockReward.Gamma.Alpha)
	}
	if conf.BlockReward.Zeta.Mu <= 0 {
		return fmt.Errorf("invalid block_reward.zeta.mu <= 0: %v", conf.BlockReward.Zeta.Mu)
	}
	if conf.BlockReward.Zeta.I <= 0 {
		return fmt.Errorf("invalid block_reward.zeta.i <= 0: %v", conf.BlockReward.Zeta.I)
	}
	if conf.BlockReward.Zeta.K <= 0 {
		return fmt.Errorf("invalid block_reward.zeta.k <=0: %v", conf.BlockReward.Zeta.K)
	}

	return
}

func (conf *Config) validateStakeRange(min, max currency.Coin) (err error) {
	if min < conf.MinStake {
		return fmt.Errorf("min_stake is less than allowed by SC: %v < %v", min,
			conf.MinStake)
	}
	if max > conf.MaxStake {
		return fmt.Errorf("max_stake is greater than allowed by SC: %v > %v",
			max, conf.MaxStake)
	}
	if max < min {
		return fmt.Errorf("max_stake less than min_stake: %v < %v", max, min)
	}
	return
}

func (conf *Config) ValidateStakeRange(min, max currency.Coin) (err error) {
	return conf.validateStakeRange(min, max)
}

func (conf *Config) Clone() statecache.Value {
	// cc := *conf
	// if conf.ReadPool != nil {
	// 	cc.ReadPool = &readPoolConfig{}
	// 	*cc.ReadPool = *conf.ReadPool
	// }

	// if conf.WritePool != nil {
	// 	cc.WritePool = &writePoolConfig{}
	// 	*cc.WritePool = *conf.WritePool
	// }

	// if conf.StakePool != nil {
	// 	cc.StakePool = &stakePoolConfig{}
	// 	*cc.StakePool = *conf.StakePool
	// }

	// if conf.BlockReward != nil {
	// 	cc.BlockReward = &blockReward{}
	// 	*cc.BlockReward = *conf.BlockReward
	// }

	// cc.Cost = make(map[string]int, len(conf.Cost))
	// for k, v := range conf.Cost {
	// 	cc.Cost[k] = v
	// }

	v, err := conf.MarshalMsg(nil)
	if err != nil {
		panic(fmt.Sprintf("could not marshal config: %v", err))
	}

	cc := newConfig()
	_, err = cc.UnmarshalMsg(v)
	if err != nil {
		panic(fmt.Sprintf("could not unmarshal config: %v", err))
	}

	return cc
}

func (conf *Config) CopyFrom(v interface{}) bool {
	cc, ok := v.(*Config)
	if !ok {
		return false
	}

	cv := cc.Clone().(*Config)
	*conf = *cv
	return true
}

// configs from sc.yaml
func getConfiguredConfig() (conf *Config, err error) {
	const pfx = "smart_contracts.storagesc."

	conf = newConfig()
	var scc = config.SmartContractConfig
	// sc
	conf.TimeUnit = scc.GetDuration(pfx + "time_unit")
	conf.MinStake, err = currency.ParseZCN(scc.GetFloat64(pfx + "min_stake"))
	if err != nil {
		return nil, err
	}
	conf.MinStakePerDelegate, err = currency.ParseZCN(scc.GetFloat64(pfx + "min_stake_per_delegate"))
	if err != nil {
		return nil, err
	}
	conf.MaxStake, err = currency.ParseZCN(scc.GetFloat64(pfx + "max_stake"))
	if err != nil {
		return nil, err
	}
	conf.MinAllocSize = scc.GetInt64(pfx + "min_alloc_size")
	conf.HealthCheckPeriod = scc.GetDuration(pfx + "health_check_period")
	conf.MaxChallengeCompletionRounds = scc.GetInt64(pfx + "max_challenge_completion_rounds")
	conf.MinBlobberCapacity = scc.GetInt64(pfx + "min_blobber_capacity")
	conf.ValidatorReward = scc.GetFloat64(pfx + "validator_reward")
	conf.BlobberSlash = scc.GetFloat64(pfx + "blobber_slash")
	conf.CancellationCharge = scc.GetFloat64(pfx + "cancellation_charge")
	conf.MaxBlobbersPerAllocation = scc.GetInt(pfx + "max_blobbers_per_allocation")
	conf.MaxReadPrice, err = currency.ParseZCN(scc.GetFloat64(pfx + "max_read_price"))
	if err != nil {
		return nil, err
	}
	conf.MinWritePrice, err = currency.ParseZCN(scc.GetFloat64(pfx + "min_write_price"))
	if err != nil {
		return nil, err
	}
	conf.MaxWritePrice, err = currency.ParseZCN(scc.GetFloat64(pfx + "max_write_price"))
	if err != nil {
		return nil, err
	}
	conf.MaxFileSize = scc.GetInt64(pfx + "max_file_size")
	// read pool
	conf.ReadPool = new(readPoolConfig)
	conf.ReadPool.MinLock, err = currency.ParseZCN(scc.GetFloat64(pfx + "readpool.min_lock"))
	if err != nil {
		return nil, err
	}

	// write pool
	conf.WritePool = new(writePoolConfig)
	conf.WritePool.MinLock, err = currency.ParseZCN(scc.GetFloat64(pfx + "writepool.min_lock"))
	if err != nil {
		return nil, err
	}
	// stake pool
	conf.StakePool = new(stakePoolConfig)
	conf.StakePool.MinLockPeriod = scc.GetDuration(pfx + "stakepool.min_lock_period")
	conf.StakePool.KillSlash = scc.GetFloat64(pfx + "stakepool.kill_slash")

	conf.MaxTotalFreeAllocation, err = currency.MultFloat64(1e10, scc.GetFloat64(pfx+"max_total_free_allocation"))
	if err != nil {
		return nil, err
	}

	conf.MaxIndividualFreeAllocation, err = currency.MultFloat64(1e10, scc.GetFloat64(pfx+"max_individual_free_allocation"))
	if err != nil {
		return nil, err
	}

	fas := pfx + "free_allocation_settings."
	conf.FreeAllocationSettings.DataShards = int(scc.GetFloat64(fas + "data_shards"))
	conf.FreeAllocationSettings.ParityShards = int(scc.GetFloat64(fas + "parity_shards"))
	conf.FreeAllocationSettings.Size = int64(scc.GetFloat64(fas + "size"))

	readPriceRangeMin, err := currency.MultFloat64(1e10, scc.GetFloat64(fas+"read_price_range.min"))
	if err != nil {
		return nil, err
	}

	readPriceRangeMax, err := currency.MultFloat64(1e10, scc.GetFloat64(fas+"read_price_range.max"))
	if err != nil {
		return nil, err
	}

	conf.FreeAllocationSettings.ReadPriceRange = PriceRange{
		Min: readPriceRangeMin,
		Max: readPriceRangeMax,
	}

	writePriceRangeMin, err := currency.MultFloat64(1e10, scc.GetFloat64(fas+"write_price_range.min"))
	if err != nil {
		return nil, err
	}

	writePriceRangeMax, err := currency.MultFloat64(1e10, scc.GetFloat64(fas+"write_price_range.max"))
	if err != nil {
		return nil, err
	}

	conf.FreeAllocationSettings.WritePriceRange = PriceRange{
		Min: writePriceRangeMin,
		Max: writePriceRangeMax,
	}
	conf.FreeAllocationSettings.ReadPoolFraction = scc.GetFloat64(fas + "read_pool_fraction")

	// challenges generating
	conf.ChallengeEnabled = scc.GetBool(pfx + "challenge_enabled")
	conf.ChallengeGenerationGap = scc.GetInt64(pfx + "challenge_generation_gap")
	conf.ValidatorsPerChallenge = scc.GetInt(pfx + "validators_per_challenge")
	conf.NumValidatorsRewarded = scc.GetInt(pfx + "num_validators_rewarded")
	conf.MaxBlobberSelectForChallenge = scc.GetInt(pfx + "max_blobber_select_for_challenge")
	conf.MaxDelegates = scc.GetInt(pfx + "max_delegates")
	conf.MaxCharge = scc.GetFloat64(pfx + "max_charge")

	conf.BlockReward = new(blockReward)
	conf.BlockReward.BlockReward, err = currency.ParseZCN(scc.GetFloat64(pfx + "block_reward.block_reward"))
	if err != nil {
		return nil, err
	}
	conf.BlockReward.BlockRewardChangePeriod = scc.GetInt64(pfx + "block_reward.block_reward_change_period")
	conf.BlockReward.BlockRewardChangeRatio = scc.GetFloat64(pfx + "block_reward.block_reward_change_ratio")
	conf.BlockReward.QualifyingStake, err = currency.ParseZCN(scc.GetFloat64(pfx + "block_reward.qualifying_stake"))
	if err != nil {
		return nil, err
	}
	conf.BlockReward.TriggerPeriod = scc.GetInt64(pfx + "block_reward.trigger_period")
	conf.BlockReward.Gamma.Alpha = scc.GetFloat64(pfx + "block_reward.gamma.alpha")
	conf.BlockReward.Gamma.A = scc.GetFloat64(pfx + "block_reward.gamma.a")
	conf.BlockReward.Gamma.B = scc.GetFloat64(pfx + "block_reward.gamma.b")
	conf.BlockReward.Zeta.I = scc.GetFloat64(pfx + "block_reward.zeta.i")
	conf.BlockReward.Zeta.K = scc.GetFloat64(pfx + "block_reward.zeta.k")
	conf.BlockReward.Zeta.Mu = scc.GetFloat64(pfx + "block_reward.zeta.mu")

	conf.OwnerId = scc.GetString(pfx + "owner_id")
	conf.Cost = scc.GetStringMapInt(pfx + "cost")

	err = conf.validate()
	return
}

func InitConfig(balances chainState.StateContextI) error {
	err := balances.GetTrieNode(scConfigKey(ADDRESS), &Config{})
	if err == util.ErrValueNotPresent {
		conf, err := getConfiguredConfig()
		if err != nil {
			return err
		}
		_, err = balances.InsertTrieNode(scConfigKey(ADDRESS), conf)
		return err
	}
	return err
}

// getConfig
func (ssc *StorageSmartContract) getConfig(
	balances chainState.StateContextI, _ bool) (
	conf *Config, err error) {

	conf = newConfig()
	err = balances.GetTrieNode(scConfigKey(ADDRESS), conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func (ssc *StorageSmartContract) saveConfig(balances chainState.StateContextI, conf *Config) error {
	_, err := balances.InsertTrieNode(scConfigKey(ADDRESS), conf)
	return err
}

// getReadPoolConfig
func (ssc *StorageSmartContract) getReadPoolConfig(
	balances chainState.StateContextI, setup bool) (
	conf *readPoolConfig, err error) {

	var scconf *Config
	if scconf, err = ssc.getConfig(balances, setup); err != nil {
		return
	}
	return scconf.ReadPool, nil
}
