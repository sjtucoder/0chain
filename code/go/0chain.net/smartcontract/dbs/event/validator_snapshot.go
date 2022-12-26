package event

import (
	"github.com/0chain/common/core/currency"
	"github.com/0chain/common/core/logging"
	"go.uber.org/zap"
)

// swagger:model ValidatorSnapshot
type ValidatorSnapshot struct {
	ValidatorID string `json:"id" gorm:"index"`

	UnstakeTotal  currency.Coin `json:"unstake_total"`
	TotalStake    currency.Coin `json:"total_stake"`
	ServiceCharge float64       `json:"service_charge"`
}

// nolint
func (edb *EventDb) getValidatorSnapshots(limit, offset int64) (map[string]ValidatorSnapshot, error) {
	var snapshots []ValidatorSnapshot
	result := edb.Store.Get().
		Raw("SELECT * FROM validator_snapshots WHERE validator_id in (select id from temp_ids ORDER BY ID limit ? offset ?)", limit, offset).
		Scan(&snapshots)
	if result.Error != nil {
		return nil, result.Error
	}

	var mapSnapshots = make(map[string]ValidatorSnapshot, len(snapshots))
	logging.Logger.Debug("get_validator_snapshot", zap.Int("snapshots selected", len(snapshots)))
	logging.Logger.Debug("get_validator_snapshot", zap.Int64("snapshots rows selected", result.RowsAffected))

	for _, snapshot := range snapshots {
		mapSnapshots[snapshot.ValidatorID] = snapshot
	}

	result = edb.Store.Get().Where("validator_id IN (select id from temp_ids ORDER BY ID limit ? offset ?)", limit, offset).Delete(&ValidatorSnapshot{})
	logging.Logger.Debug("get_validator_snapshot", zap.Int64("deleted rows", result.RowsAffected))
	return mapSnapshots, result.Error
}

// nolint
func (edb *EventDb) addValidatorSnapshot(validators []Validator) error {
	var snapshots []ValidatorSnapshot
	for _, validator := range validators {
		snapshots = append(snapshots, ValidatorSnapshot{
			ValidatorID:   validator.ID,
			UnstakeTotal:  validator.UnstakeTotal,
			TotalStake:    validator.TotalStake,
			ServiceCharge: validator.ServiceCharge,
		})
	}

	return edb.Store.Get().Create(&snapshots).Error
}
