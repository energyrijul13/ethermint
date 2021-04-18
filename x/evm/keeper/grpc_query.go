package keeper

import (
	"context"
	"math/big"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/types"

	ethcmn "github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/ethermint/metrics"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

var _ types.QueryServer = Keeper{}

// Account implements the Query/Account gRPC method
func (k Keeper) Account(c context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsZeroAddress(req.Address) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	so := k.GetOrNewStateObject(ctx, ethcmn.HexToAddress(req.Address))
	balance, err := ethermint.MarshalBigInt(so.Balance())
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, err
	}

	return &types.QueryAccountResponse{
		Balance:  balance,
		CodeHash: so.CodeHash(),
		Nonce:    so.Nonce(),
	}, nil
}

func (k Keeper) CosmosAccount(c context.Context, req *types.QueryCosmosAccountRequest) (*types.QueryCosmosAccountResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsZeroAddress(req.Address) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	ethStr := req.Address
	ethAddr := ethcmn.FromHex(ethStr)

	ethToCosmosAddr := sdk.AccAddress(ethAddr[:]).String()
	cosmosToEthAddr, _ := sdk.AccAddressFromBech32(ethToCosmosAddr)

	acc := k.accountKeeper.GetAccount(ctx, cosmosToEthAddr)
	res := types.QueryCosmosAccountResponse{
		CosmosAddress: cosmosToEthAddr.String(),
	}
	if acc != nil {
		res.Sequence = acc.GetSequence()
		res.AccountNumber = acc.GetAccountNumber()
	}
	return &res, nil
}

// Balance implements the Query/Balance gRPC method
func (k Keeper) Balance(c context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsZeroAddress(req.Address) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	balanceInt := k.GetBalance(ctx, ethcmn.HexToAddress(req.Address))
	balance, err := ethermint.MarshalBigInt(balanceInt)
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.Internal,
			"failed to marshal big.Int to string",
		)
	}

	return &types.QueryBalanceResponse{
		Balance: balance,
	}, nil
}

// Storage implements the Query/Storage gRPC method
func (k Keeper) Storage(c context.Context, req *types.QueryStorageRequest) (*types.QueryStorageResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsZeroAddress(req.Address) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	address := ethcmn.HexToAddress(req.Address)
	key := ethcmn.HexToHash(req.Key)

	state := k.GetState(ctx, address, key)

	return &types.QueryStorageResponse{
		Value: state.String(),
	}, nil
}

// Code implements the Query/Code gRPC method
func (k Keeper) Code(c context.Context, req *types.QueryCodeRequest) (*types.QueryCodeResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsZeroAddress(req.Address) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	address := ethcmn.HexToAddress(req.Address)
	code := k.GetCode(ctx, address)

	return &types.QueryCodeResponse{
		Code: code,
	}, nil
}

// TxLogs implements the Query/TxLogs gRPC method
func (k Keeper) TxLogs(c context.Context, req *types.QueryTxLogsRequest) (*types.QueryTxLogsResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsEmptyHash(req.Hash) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrEmptyHash.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	hash := ethcmn.HexToHash(req.Hash)
	logs, err := k.GetLogs(ctx, hash)
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &types.QueryTxLogsResponse{
		Logs: types.NewTransactionLogsFromEth(hash, logs).Logs,
	}, nil
}

// TxReceipt implements the Query/TxReceipt gRPC method
func (k Keeper) TxReceipt(c context.Context, req *types.QueryTxReceiptRequest) (*types.QueryTxReceiptResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsEmptyHash(req.Hash) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrEmptyHash.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	hash := ethcmn.HexToHash(req.Hash)
	receipt, found := k.GetTxReceiptFromHash(ctx, hash)
	if !found {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.NotFound, types.ErrTxReceiptNotFound.Error(),
		)
	}

	return &types.QueryTxReceiptResponse{
		Receipt: receipt,
	}, nil
}

// TxReceiptsByBlockHeight implements the Query/TxReceiptsByBlockHeight gRPC method
func (k Keeper) TxReceiptsByBlockHeight(c context.Context, req *types.QueryTxReceiptsByBlockHeightRequest) (*types.QueryTxReceiptsByBlockHeightResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	receipts := k.GetTxReceiptsByBlockHeight(ctx, req.Height)
	return &types.QueryTxReceiptsByBlockHeightResponse{
		Receipts: receipts,
	}, nil
}

// TxReceiptsByBlockHash implements the Query/TxReceiptsByBlockHash gRPC method
func (k Keeper) TxReceiptsByBlockHash(c context.Context, req *types.QueryTxReceiptsByBlockHashRequest) (*types.QueryTxReceiptsByBlockHashResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsEmptyHash(req.Hash) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrEmptyHash.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	hash := ethcmn.HexToHash(req.Hash)
	receipts := k.GetTxReceiptsByBlockHash(ctx, hash)

	return &types.QueryTxReceiptsByBlockHashResponse{
		Receipts: receipts,
	}, nil
}

// BlockLogs implements the Query/BlockLogs gRPC method
func (k Keeper) BlockLogs(c context.Context, req *types.QueryBlockLogsRequest) (*types.QueryBlockLogsResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if types.IsEmptyHash(req.Hash) {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrEmptyHash.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	txLogs := k.GetAllTxLogs(ctx)

	return &types.QueryBlockLogsResponse{
		TxLogs: txLogs,
	}, nil
}

// BlockBloom implements the Query/BlockBloom gRPC method
func (k Keeper) BlockBloom(c context.Context, req *types.QueryBlockBloomRequest) (*types.QueryBlockBloomResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	height := ctx.BlockHeight()
	if setHeight := req.Height; setHeight > 0 {
		height = setHeight
	}

	bloom, found := k.GetBlockBloom(ctx, height)
	if !found {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(
			codes.NotFound, types.ErrBloomNotFound.Error(),
		)
	}

	return &types.QueryBlockBloomResponse{
		Bloom: bloom.Bytes(),
	}, nil
}

// Params implements the Query/Params gRPC method
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	if req == nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// StaticCall implements Query/StaticCall gRPCP method
func (k Keeper) StaticCall(c context.Context, req *types.QueryStaticCallRequest) (*types.QueryStaticCallResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// parse the chainID from a string to a base-10 integer
	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := ethcmn.BytesToHash(txHash)

	var recipient *ethcmn.Address
	if len(req.Address) > 0 {
		addr := ethcmn.HexToAddress(req.Address)
		recipient = &addr
	}

	so := k.GetOrNewStateObject(ctx, *recipient)
	sender := ethcmn.HexToAddress("0xaDd00275E3d9d213654Ce5223f0FADE8b106b707")

	st := &types.StateTransition{
		AccountNonce: so.Nonce(),
		Price:        new(big.Int).SetBytes(big.NewInt(0).Bytes()),
		GasLimit:     100000000,
		Recipient:    recipient,
		Amount:       new(big.Int).SetBytes(big.NewInt(0).Bytes()),
		Payload:      req.Input,
		Csdb:         k.CommitStateDB.WithContext(ctx),
		ChainID:      chainIDEpoch,
		TxHash:       &ethHash,
		Sender:       sender,
		Simulate:     ctx.IsCheckTx(),
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		return nil, types.ErrChainConfigNotFound
	}

	ret, err := st.StaticCall(ctx, config)
	if err != nil {
		return nil, err
	}

	return &types.QueryStaticCallResponse{Data: ret}, nil
}