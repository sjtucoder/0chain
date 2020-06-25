package storagesc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	chainState "0chain.net/chaincore/chain/state"
	"0chain.net/chaincore/config"
	"0chain.net/chaincore/state"
	"0chain.net/chaincore/transaction"
	"0chain.net/core/common"
	"0chain.net/core/datastore"
	"0chain.net/core/util"
)

func scConfigKey(scKey string) datastore.Key {
	return datastore.Key(scKey + ":configurations")
}

// stake pool configs

type stakePoolConfig struct {
	MinLock int64 `json:"min_lock"`
	// Interest rate of the stake pool
	InterestRate     float64       `json:"interest_rate"`
	InterestInterval time.Duration `json:"interest_interval"`
}

// read pool configs

type readPoolConfig struct {
	MinLock       int64         `json:"min_lock"`
	MinLockPeriod time.Duration `json:"min_lock_period"`
	MaxLockPeriod time.Duration `json:"max_lock_period"`
}

// write pool configurations

type writePoolConfig struct {
	MinLock       int64         `json:"min_lock"`
	MinLockPeriod time.Duration `json:"min_lock_period"`
	MaxLockPeriod time.Duration `json:"max_lock_period"`
}

// scConfig represents SC configurations ('storagesc:' from sc.yaml).
type scConfig struct {
	// MaxMint is max minting.
	MaxMint state.Balance `json:"max_mint"`
	// Minted tokens by entire SC.
	Minted state.Balance `json:"minted"`
	// MinAllocSize is minimum possible size (bytes)
	// of an allocation the SC accept.
	MinAllocSize int64 `json:"min_alloc_size"`
	// MinAllocDuration is minimum possible duration of an
	// allocation allowed by the SC.
	MinAllocDuration time.Duration `json:"min_alloc_duration"`
	// MaxChallengeCompletionTime is max time to complete a challenge.
	MaxChallengeCompletionTime time.Duration `json:"max_challenge_completion_time"`
	// MinOfferDuration represents lower boundary of blobber's MaxOfferDuration.
	MinOfferDuration time.Duration `json:"min_offer_duration"`
	// MinBlobberCapacity allowed to register in the SC.
	MinBlobberCapacity int64 `json:"min_blobber_capacity"`
	// ReadPool related configurations.
	ReadPool *readPoolConfig `json:"readpool"`
	// WritePool related configurations.
	WritePool *writePoolConfig `json:"writepool"`
	// StakePool related configurations.
	StakePool *stakePoolConfig `json:"stakepool"`
	// ValidatorReward represents % (value in [0; 1] range) of blobbers' reward
	// goes to validators. Even if a blobber doesn't pass a challenge validators
	// receive this reward.
	ValidatorReward float64 `json:"validator_reward"`
	// BlobberSlash represents % (value in [0; 1] range) of blobbers' stake
	// tokens penalized on challenge not passed.
	BlobberSlash float64 `json:"blobber_slash"`

	// price limits for blobbers

	// MaxReadPrice allowed for a blobber.
	MaxReadPrice state.Balance `json:"max_read_price"`
	// MaxWrtiePrice
	MaxWritePrice state.Balance `json:"max_write_price"`

	// allocation cancellation

	// FailedChallengesToCancel is number of failed challenges of an allocation
	// to be able to cancel an allocation.
	FailedChallengesToCancel int `json:"failed_challenges_to_cancel"`
	// FailedChallengesToRevokeMinLock is number of failed challenges of a
	// blobber to revoke its min_lock demand back to user; only part not
	// paid yet can go back.
	FailedChallengesToRevokeMinLock int `json:"failed_challenges_to_revoke_min_lock"`

	// challenges generating

	// ChallengeEnabled is challenges generating pin.
	ChallengeEnabled bool `json:"challenge_enabled"`
	// MaxChallengesPerGeneration is max number of challenges can be generated
	// at once for a blobber-allocation pair with size difference for the
	// moment of the generation.
	MaxChallengesPerGeneration int `json:"max_challenges_per_generation"`
	// ChallengeGenerationRate is number of challenges generated for a MB/min.
	ChallengeGenerationRate float64 `json:"challenge_rate_per_mb_min"`

	// MinStake allowed by a blobber/validator (entire SC boundary).
	MinStake state.Balance `json:"min_stake"`
	// MaxStake allowed by a blobber/validator (entire SC boundary).
	MaxStake state.Balance `json:"max_stake"`

	// max delegates per stake pool
	MaxDelegates int `json:"max_delegates"`
}

func (sc *scConfig) validate() (err error) {
	if sc.ValidatorReward < 0.0 || 1.0 < sc.ValidatorReward {
		return fmt.Errorf("validator_reward not in [0; 1] range: %v",
			sc.ValidatorReward)
	}
	if sc.BlobberSlash < 0.0 || 1.0 < sc.BlobberSlash {
		return fmt.Errorf("blobber_slash not in [0; 1] range: %v",
			sc.BlobberSlash)
	}
	if sc.MinBlobberCapacity < 0 {
		return fmt.Errorf("negative min_blobber_capacity: %v",
			sc.MinBlobberCapacity)
	}
	if sc.MinOfferDuration < 0 {
		return fmt.Errorf("negative min_offer_duration: %v",
			sc.MinOfferDuration)
	}
	if sc.MaxChallengeCompletionTime < 0 {
		return fmt.Errorf("negative max_challenge_completion_time: %v",
			sc.MaxChallengeCompletionTime)
	}
	if sc.MinAllocDuration < 0 {
		return fmt.Errorf("negative min_alloc_duration: %v",
			sc.MinAllocDuration)
	}
	if sc.MaxMint < 0 {
		return fmt.Errorf("negative max_mint: %v", sc.MaxMint)
	}
	if sc.MinAllocSize < 0 {
		return fmt.Errorf("negative min_alloc_size: %v", sc.MinAllocSize)
	}
	if sc.MaxReadPrice < 0 {
		return fmt.Errorf("negative max_read_price: %v", sc.MaxReadPrice)
	}
	if sc.MaxWritePrice < 0 {
		return fmt.Errorf("negative max_write_price: %v", sc.MaxWritePrice)
	}
	if sc.StakePool.MinLock <= 1 {
		return fmt.Errorf("invalid stakepool.min_lock: %v <= 1",
			sc.StakePool.MinLock)
	}
	if sc.StakePool.InterestRate < 0 {
		return fmt.Errorf("negative stakepool.interest_rate: %v",
			sc.StakePool.InterestRate)
	}
	if sc.StakePool.InterestInterval <= 0 {
		return fmt.Errorf("invalid stakepool.interest_interval <= 0: %v",
			sc.StakePool.InterestInterval)
	}
	if sc.FailedChallengesToCancel < 0 {
		return fmt.Errorf("negative failed_challenges_to_cancel: %v",
			sc.FailedChallengesToCancel)
	}
	if sc.FailedChallengesToRevokeMinLock < 0 {
		return fmt.Errorf("negative failed_challenges_to_revoke_min_lock: %v",
			sc.FailedChallengesToRevokeMinLock)
	}
	if sc.MaxChallengesPerGeneration <= 0 {
		return fmt.Errorf("invalid max_challenges_per_generation <= 0: %v",
			sc.MaxChallengesPerGeneration)
	}
	if sc.ChallengeGenerationRate < 0 {
		return fmt.Errorf("negative challenge_rate_per_mb_min: %v",
			sc.ChallengeGenerationRate)
	}
	if sc.MinStake < 0 {
		return fmt.Errorf("negative min_stake: %v", sc.MinStake)
	}
	if sc.MaxStake < sc.MinStake {
		return fmt.Errorf("max_stake less than min_stake: %v < %v", sc.MinStake,
			sc.MaxStake)
	}
	if sc.MaxDelegates < 1 {
		return fmt.Errorf("max_delegates is too small %v", sc.MaxDelegates)
	}
	return
}

func (conf *scConfig) canMint() bool {
	return conf.Minted < conf.MaxMint
}

func (conf *scConfig) validateStakeRange(min, max state.Balance) (err error) {
	if min < conf.MinStake {
		return fmt.Errorf("min_stake is less than allowed by SC: %v < %v", min,
			conf.MinStake)
	}
	if max > conf.MaxStake {
		return fmt.Errorf("max_stake is greater than allowed by SC: %v < %v",
			max, conf.MaxStake)
	}
	if max < min {
		return fmt.Errorf("max_stake less than min_stake: %v < %v", min, max)
	}
	return
}

func (conf *scConfig) Encode() (b []byte) {
	var err error
	if b, err = json.Marshal(conf); err != nil {
		panic(err) // must not happens
	}
	return
}

func (conf *scConfig) Decode(b []byte) error {
	return json.Unmarshal(b, conf)
}

//
// rest handler and update function
//

// getConfigBytes returns encoded configurations or an error.
func (ssc *StorageSmartContract) getConfigBytes(
	balances chainState.StateContextI) (b []byte, err error) {

	var val util.Serializable
	val, err = balances.GetTrieNode(scConfigKey(ssc.ID))
	if err != nil {
		return
	}
	return val.Encode(), nil
}

// configs from sc.yaml
func getConfiguredConfig() (conf *scConfig, err error) {

	const pfx = "smart_contracts.storagesc."

	conf = new(scConfig)
	var scc = config.SmartContractConfig
	// sc
	conf.MaxMint = state.Balance(scc.GetFloat64(pfx+"max_mint") * 1e10)
	conf.MinStake = state.Balance(scc.GetFloat64(pfx+"min_stake") * 1e10)
	conf.MaxStake = state.Balance(scc.GetFloat64(pfx+"max_stake") * 1e10)
	conf.MinAllocSize = scc.GetInt64(pfx + "min_alloc_size")
	conf.MinAllocDuration = scc.GetDuration(pfx + "min_alloc_duration")
	conf.MaxChallengeCompletionTime = scc.GetDuration(pfx + "max_challenge_completion_time")
	conf.MinOfferDuration = scc.GetDuration(pfx + "min_offer_duration")
	conf.MinBlobberCapacity = scc.GetInt64(pfx + "min_blobber_capacity")
	conf.ValidatorReward = scc.GetFloat64(pfx + "validator_reward")
	conf.BlobberSlash = scc.GetFloat64(pfx + "blobber_slash")
	conf.MaxReadPrice = state.Balance(
		scc.GetFloat64(pfx+"max_read_price") * 1e10)
	conf.MaxWritePrice = state.Balance(
		scc.GetFloat64(pfx+"max_write_price") * 1e10)
	// read pool
	conf.ReadPool = new(readPoolConfig)
	conf.ReadPool.MinLock = scc.GetInt64(pfx + "readpool.min_lock")
	conf.ReadPool.MinLockPeriod = scc.GetDuration(
		pfx + "readpool.min_lock_period")
	conf.ReadPool.MaxLockPeriod = scc.GetDuration(
		pfx + "readpool.max_lock_period")
	// write pool
	conf.WritePool = new(writePoolConfig)
	conf.WritePool.MinLock = scc.GetInt64(pfx + "writepool.min_lock")
	conf.WritePool.MinLockPeriod = scc.GetDuration(
		pfx + "writepool.min_lock_period")
	conf.WritePool.MaxLockPeriod = scc.GetDuration(
		pfx + "writepool.max_lock_period")
	// stake pool
	conf.StakePool = new(stakePoolConfig)
	conf.StakePool.MinLock = scc.GetInt64(pfx + "stakepool.min_lock")
	conf.StakePool.InterestRate = scc.GetFloat64(
		pfx + "stakepool.interest_rate")
	conf.StakePool.InterestInterval = scc.GetDuration(
		pfx + "stakepool.interest_interval")
	// allocation cancellation
	conf.FailedChallengesToCancel = scc.GetInt(
		pfx + "failed_challenges_to_cancel")
	conf.FailedChallengesToRevokeMinLock = scc.GetInt(
		pfx + "failed_challenges_to_revoke_min_lock")
	// challenges generating
	conf.ChallengeEnabled = scc.GetBool(pfx + "challenge_enabled")
	conf.MaxChallengesPerGeneration = scc.GetInt(
		pfx + "max_challenges_per_generation")
	conf.ChallengeGenerationRate = scc.GetFloat64(
		pfx + "challenge_rate_per_mb_min")

	conf.MaxDelegates = scc.GetInt(pfx + "max_delegates")

	err = conf.validate()
	return
}

func (ssc *StorageSmartContract) setupConfig(
	balances chainState.StateContextI) (conf *scConfig, err error) {

	if conf, err = getConfiguredConfig(); err != nil {
		return
	}
	_, err = balances.InsertTrieNode(scConfigKey(ssc.ID), conf)
	if err != nil {
		return nil, err
	}
	return
}

// getConfig
func (ssc *StorageSmartContract) getConfig(
	balances chainState.StateContextI, setup bool) (
	conf *scConfig, err error) {

	var confb []byte
	confb, err = ssc.getConfigBytes(balances)
	if err != nil && err != util.ErrValueNotPresent {
		return
	}

	conf = new(scConfig)

	if err == util.ErrValueNotPresent {
		if !setup {
			return // value not present
		}
		return ssc.setupConfig(balances)
	}

	if err = conf.Decode(confb); err != nil {
		return nil, err
	}
	return
}

func (ssc *StorageSmartContract) getConfigHandler(ctx context.Context,
	params url.Values, balances chainState.StateContextI) (
	resp interface{}, err error) {

	var conf *scConfig
	conf, err = ssc.getConfig(balances, false)

	if err != nil && err != util.ErrValueNotPresent {
		return // unexpected error
	}

	// return configurations from sc.yaml not saving them
	if err == util.ErrValueNotPresent {
		return getConfiguredConfig()
	}

	return conf, nil // actual value
}

// updateConfig is SC function used by SC owner
// to update storage SC configurations
func (ssc *StorageSmartContract) updateConfig(t *transaction.Transaction,
	input []byte, balances chainState.StateContextI) (resp string, err error) {

	if t.ClientID != owner {
		return "", common.NewError("update_config",
			"unauthorized access - only the owner can update the variables")
	}

	var conf *scConfig
	if conf, err = ssc.getConfig(balances, true); err != nil {
		return "", common.NewError("update_config",
			"can't get config: "+err.Error())
	}

	var update scConfig
	if err = update.Decode(input); err != nil {
		return "", common.NewError("update_config", err.Error())
	}

	if err = update.validate(); err != nil {
		return
	}

	update.Minted = conf.Minted

	_, err = balances.InsertTrieNode(scConfigKey(ssc.ID), &update)
	if err != nil {
		return "", common.NewError("update_config", err.Error())
	}

	return string(update.Encode()), nil
}

// getWritePoolConfig
func (ssc *StorageSmartContract) getWritePoolConfig(
	balances chainState.StateContextI, setup bool) (
	conf *writePoolConfig, err error) {

	var scconf *scConfig
	if scconf, err = ssc.getConfig(balances, setup); err != nil {
		return
	}
	return scconf.WritePool, nil
}

// getReadPoolConfig
func (ssc *StorageSmartContract) getReadPoolConfig(
	balances chainState.StateContextI, setup bool) (
	conf *readPoolConfig, err error) {

	var scconf *scConfig
	if scconf, err = ssc.getConfig(balances, setup); err != nil {
		return
	}
	return scconf.ReadPool, nil
}
