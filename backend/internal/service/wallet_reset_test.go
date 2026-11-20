package service

import (
	"math"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestWalletResetCreditCalculation(t *testing.T) {
	require.Equal(t, "3.77", resetCardCredit(decimal.RequireFromString("9.425"), 40, 150).String())
	require.Equal(t, "3", resetCardCredit(decimal.NewFromInt(10), 40, 3).String())
	require.Equal(t, "0.00000001", resetCardCredit(decimal.RequireFromString("0.00000003"), 50, 10).String())
	require.True(t, resetCardCredit(decimal.Zero, 20, 100).IsZero())
	id := int64(5)
	require.NoError(t, validateWalletResetCard(&id, 30, 150, 7))
	require.Error(t, validateWalletResetCard(&id, 0, 150, 7))
	require.Error(t, validateWalletResetCard(&id, 101, 150, 7))
	require.Error(t, validateWalletResetCard(&id, math.NaN(), 150, 7))
	require.Error(t, validateWalletResetCard(&id, 30, math.Inf(1), 7))
	require.Error(t, validateWalletResetCard(&id, 30, 150, 366))
	require.Error(t, validateWalletResetCard(nil, 30, 150, 7))
}
