package v16

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/gaia/v16/app/upgrades"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v16"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			feemarkettypes.ModuleName,
		},
	},
}
