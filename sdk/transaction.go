// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sdk

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/proximax-storage/go-xpx-utils/net"
)

type TransactionService struct {
	*service
	BlockchainService *BlockchainService
}

// returns Transaction for passed transaction id or hash
func (txs *TransactionService) GetTransaction(ctx context.Context, id string) (Transaction, error) {
	var b bytes.Buffer

	url := net.NewUrl(fmt.Sprintf(transactionRoute, id))

	resp, err := txs.client.doNewRequest(ctx, http.MethodGet, url.Encode(), nil, &b)
	if err != nil {
		return nil, err
	}

	if err = handleResponseStatusCode(resp, map[int]error{404: ErrResourceNotFound, 409: ErrArgumentNotValid}); err != nil {
		return nil, err
	}

	return MapTransaction(&b, txs.client.GenerationHash())
}

// returns an array of Transaction's for passed array of transaction ids or hashes
func (txs *TransactionService) GetTransactions(ctx context.Context, ids []string) ([]Transaction, error) {
	var b bytes.Buffer
	txIds := &TransactionIdsDTO{
		ids,
	}

	resp, err := txs.client.doNewRequest(ctx, http.MethodPost, transactionsRoute, txIds, &b)
	if err != nil {
		return nil, err
	}

	if err = handleResponseStatusCode(resp, map[int]error{400: ErrInvalidRequest, 409: ErrArgumentNotValid}); err != nil {
		return nil, err
	}

	return MapTransactions(&b, txs.client.GenerationHash())
}

// returns transaction hash after announcing passed SignedTransaction
func (txs *TransactionService) Announce(ctx context.Context, tx *SignedTransaction) (string, error) {
	dto := signedTransactionDto{
		tx.EntityType,
		tx.Payload,
		tx.Hash.String(),
	}
	return txs.announceTransaction(ctx, &dto, transactionsRoute)
}

// returns transaction hash after announcing passed aggregate bounded SignedTransaction
func (txs *TransactionService) AnnounceAggregateBonded(ctx context.Context, tx *SignedTransaction) (string, error) {
	dto := signedTransactionDto{
		tx.EntityType,
		tx.Payload,
		tx.Hash.String(),
	}
	return txs.announceTransaction(ctx, &dto, announceAggregateRoute)
}

// returns transaction hash after announcing passed CosignatureSignedTransaction
func (txs *TransactionService) AnnounceAggregateBondedCosignature(ctx context.Context, c *CosignatureSignedTransaction) (string, error) {
	dto := cosignatureSignedTransactionDto{
		c.ParentHash.String(),
		c.Signature.String(),
		c.Signer,
	}
	return txs.announceTransaction(ctx, &dto, announceAggregateCosignatureRoute)
}

// returns TransactionStatus for passed transaction id or hash
func (txs *TransactionService) GetTransactionStatus(ctx context.Context, id string) (*TransactionStatus, error) {
	ts := &transactionStatusDTO{}

	resp, err := txs.client.doNewRequest(ctx, http.MethodGet, fmt.Sprintf(transactionStatusSingularRoute, id), nil, ts)
	if err != nil {
		return nil, err
	}

	if err = handleResponseStatusCode(resp, map[int]error{404: ErrResourceNotFound, 409: ErrArgumentNotValid}); err != nil {
		return nil, err
	}

	return ts.toStruct()
}

// returns TransactionsStatuses for passed transactions id or hashes
func (txs *TransactionService) GetTransactionsStatuses(ctx context.Context, hashes []string) ([]*TransactionStatus, error) {
	txIds := &TransactionHashesDTO{
		hashes,
	}

	dtos := transactionStatusDTOs(make([]*transactionStatusDTO, len(hashes)))
	resp, err := txs.client.doNewRequest(ctx, http.MethodPost, transactionsStatusPluralRoute, txIds, &dtos)
	if err != nil {
		return nil, err
	}

	if err = handleResponseStatusCode(resp, map[int]error{400: ErrInvalidRequest, 409: ErrArgumentNotValid}); err != nil {
		return nil, err
	}

	return dtos.toStruct()
}

// returns an array of TransactionStatus's for passed transaction ids or hashes
func (txs *TransactionService) GetTransactionStatuses(ctx context.Context, hashes []string) ([]*TransactionStatus, error) {
	txIds := &TransactionHashesDTO{
		hashes,
	}

	dtos := transactionStatusDTOs(make([]*transactionStatusDTO, len(hashes)))
	resp, err := txs.client.doNewRequest(ctx, http.MethodPost, transactionStatusesRoute, txIds, &dtos)
	if err != nil {
		return nil, err
	}

	if err = handleResponseStatusCode(resp, map[int]error{404: ErrResourceNotFound, 409: ErrArgumentNotValid}); err != nil {
		return nil, err
	}

	return dtos.toStruct()
}

func (txs *TransactionService) announceTransaction(ctx context.Context, tx interface{}, path string) (string, error) {
	m := struct {
		Message string `json:"message"`
	}{}

	resp, err := txs.client.doNewRequest(ctx, http.MethodPut, path, tx, &m)
	if err != nil {
		return "", err
	}

	if err = handleResponseStatusCode(resp, map[int]error{400: ErrInvalidRequest, 409: ErrArgumentNotValid}); err != nil {
		return "", err
	}

	return m.Message, nil
}

// Gets a transaction's effective paid fee
func (txs *TransactionService) GetTransactionEffectiveFee(ctx context.Context, transactionId string) (int, error) {
	tx, err := txs.GetTransaction(ctx, transactionId)
	if err != nil {
		return -1, err
	}

	block, err := txs.BlockchainService.GetBlockByHeight(ctx, tx.GetAbstractTransaction().Height)
	if err != nil {
		return -1, err
	}

	return int(block.FeeMultiplier) * tx.Size(), nil
}
