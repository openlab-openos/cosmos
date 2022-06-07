package globalfee

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/v8/x/globalfee/types"
)

var _ types.QueryServer = &GrpcQuerier{}

type GrpcQuerier struct {
	paramSource paramSource
}

func NewGrpcQuerier(paramSource paramSource) GrpcQuerier {
	return GrpcQuerier{paramSource: paramSource}
}

// MinimumGasPrices return minimum gas prices
func (g GrpcQuerier) MinimumGasPrices(stdCtx context.Context, _ *types.QueryMinimumGasPricesRequest) (*types.QueryMinimumGasPricesResponse, error) {
	var minGasPrices sdk.DecCoins
	ctx := sdk.UnwrapSDKContext(stdCtx)
	if g.paramSource.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		g.paramSource.Get(ctx, types.ParamStoreKeyMinGasPrices, &minGasPrices)
	}
	return &types.QueryMinimumGasPricesResponse{
		MinimumGasPrices: minGasPrices,
	}, nil
}
