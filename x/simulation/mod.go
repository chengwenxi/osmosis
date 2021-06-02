package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/osmosis-labs/osmosis/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func GenAndDeliverTxWithRandFees(
	r *rand.Rand,
	app *baseapp.BaseApp,
	txGen client.TxConfig,
	msg sdk.Msg,
	coinsSpentInMsg sdk.Coins,
	ctx sdk.Context,
	simAccount simtypes.Account,
	ak stakingTypes.AccountKeeper,
	bk stakingTypes.BankKeeper,
	moduleName string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := ak.GetAccount(ctx, simAccount.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(coinsSpentInMsg)
	if hasNeg {
		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "message doesn't leave room for fees"), nil, err
	}

	fees, err = simtypes.RandomFees(r, ctx, coins)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(app, txGen, msg, fees, ctx, simAccount, ak, moduleName)
}

func GenAndDeliverTx(
	app *baseapp.BaseApp,
	txGen client.TxConfig,
	msg sdk.Msg,
	fees sdk.Coins,
	ctx sdk.Context,
	simAccount simtypes.Account,
	ak stakingTypes.AccountKeeper,
	moduleName string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := ak.GetAccount(ctx, simAccount.Address)
	tx, err := helpers.GenTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		simAccount.PrivKey,
	)

	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
	}

	_, _, err = app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(msg, true, ""), nil, nil

}