package grpc

import (
	"context"
	"encoding/hex"

	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/tx"
	zarb "github.com/zarbchain/zarb-go/www/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (zs *zarbServer) GetTransaction(ctx context.Context, request *zarb.TransactionRequest) (*zarb.TransactionResponse, error) {
	id, err := crypto.HashFromString(request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid transaction ID: %v", err.Error())

	}
	trx := zs.state.Transaction(id)
	if trx == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Transaction not found")
	}

	data, err := trx.Encode()
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	var json string
	if request.Verbosity == 1 {
		bz, err := trx.MarshalJSON()
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		json = string(bz)
	}
	res := &zarb.TransactionResponse{
		Data: data,
		Json: json,
	}

	return res, nil

}

func (zs *zarbServer) SendRawTransaction(ctx context.Context, request *zarb.SendRawTransactionRequest) (*zarb.SendRawTransactionResponse, error) {
	var tx tx.Tx

	hexDecoded, err := hex.DecodeString(request.Data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Couldn't decode transaction: %v", err)
	}
	if err := tx.Decode(hexDecoded); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Couldn't decode transaction: %v", err)
	}

	if err := tx.SanityCheck(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Couldn't Verify Transaction:  %v", err)
	}

	if err := zs.state.AddPendingTxAndBroadcast(&tx); err != nil {
		return nil, status.Errorf(codes.Aborted, "Couldn't add to Pending pool: %v", err)
	}

	return &zarb.SendRawTransactionResponse{
		Id: tx.ID().String(),
	}, nil
}
