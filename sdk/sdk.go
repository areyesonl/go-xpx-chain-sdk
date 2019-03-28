// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package sdk provides a client library for the Catapult REST API.
package sdk

import (
	"bytes"
	"errors"
	"github.com/google/go-querystring/query"
	"github.com/json-iterator/go"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Provides service configuration
type Config struct {
	reputationConfig *reputationConfig
	BaseURL          *url.URL
	NetworkType
}

type reputationConfig struct {
	minInteractions   uint64
	defaultReputation float64
}

var defaultRepConfig = reputationConfig{
	minInteractions:   10,
	defaultReputation: 0.9,
}

func NewReputationConfig(minInter uint64, defaultRep float64) (*reputationConfig, error) {
	if defaultRep < 0 || defaultRep > 1 {
		return nil, ErrInvalidReputationConfig
	}

	return &reputationConfig{minInteractions: minInter, defaultReputation: defaultRep}, nil
}

// Config constructor
func NewConfig(baseUrl string, networkType NetworkType) (*Config, error) {
	return NewConfigWithReputation(baseUrl, networkType, &defaultRepConfig)
}

// Config constructor
func NewConfigWithReputation(baseUrl string, networkType NetworkType, repConf *reputationConfig) (*Config, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	c := &Config{BaseURL: u, NetworkType: networkType, reputationConfig: repConf}

	return c, nil
}

// Catapult API Client configuration
type Client struct {
	client *http.Client // HTTP client used to communicate with the API.
	config *Config
	common service // Reuse a single struct instead of allocating one for each service on the heap.
	// Services for communicating to the Catapult REST APIs
	Blockchain  *BlockchainService
	Mosaic      *MosaicService
	Namespace   *NamespaceService
	Network     *NetworkService
	Transaction *TransactionService
	Account     *AccountService
	Contract    *ContractService

	//Factories
	AccountFactory AccountFactory

	//Converters
	MultisigAccountInfoConverter              multisigAccountInfoConverter
	BlockInfoConverter                        blockInfoConverter
	MosaicInfoConverter                       mosaicInfoConverter
	NamespaceInfoConverter                    namespaceInfoConverter
	MosaicDefinitionTransactionConverter      mosaicDefinitionTransactionConverter
	AbstractTransactionConverter              abstractTransactionConverter
	TransactionInfoConverter                  transactionInfoConverter
	MosaicSupplyChangeTransactionConverter    mosaicSupplyChangeTransactionConverter
	TransferTransactionConverter              transferTransactionConverter
	ModifyMultisigAccountTransactionConverter modifyMultisigAccountTransactionConverter
	ModifyContractTransactionConverter        modifyContractTransactionConverter
	RegisterNamespaceTransactionConverter     registerNamespaceTransactionConverter
	LockFundsTransactionConverter             lockFundsTransactionConverter
	SecretLockTransactionConverter            secretLockTransactionConverter
	SecretProofTransactionConverter           secretProofTransactionConverter
	AggregateTransactionCosignatureConverter  aggregateTransactionCosignatureConverter
	AggregateTransactionConverter             aggregateTransactionConverter
	MultisigCosignatoryModificationConverter  multisigCosignatoryModificationConverter
}

type service struct {
	client *Client
}

// NewClient returns a new Catapult API client.
// If httpClient is nil then it will create http.DefaultClient
func NewClient(httpClient *http.Client, conf *Config) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{client: httpClient, config: conf}
	c.common.client = c
	c.Blockchain = (*BlockchainService)(&c.common)
	c.Mosaic = (*MosaicService)(&c.common)
	c.Namespace = (*NamespaceService)(&c.common)
	c.Network = (*NetworkService)(&c.common)
	c.Transaction = (*TransactionService)(&c.common)
	c.Account = (*AccountService)(&c.common)
	c.Contract = (*ContractService)(&c.common)

	bindFactories(c)
	bindConverters(c)

	return c
}

func bindFactories(client *Client) {
	client.AccountFactory = NewAccountFactory()
}

func bindConverters(client *Client) {
	client.MultisigAccountInfoConverter = newMultisigAccountInfoDTOConverter(client.AccountFactory)
	client.BlockInfoConverter = newBlockInfoConverter(client.AccountFactory)
	client.MosaicInfoConverter = newMosaicInfoConverter(client.AccountFactory)
	client.NamespaceInfoConverter = newNamespaceInfoConverterImpl(client.AccountFactory)

	client.TransactionInfoConverter = newTransactionInfoConverter()
	client.AbstractTransactionConverter = newAbstractTransactionConverterImpl(client.AccountFactory)
	client.MosaicDefinitionTransactionConverter = newMosaicDefinitionTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.MosaicSupplyChangeTransactionConverter = newMosaicSupplyChangeTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.TransferTransactionConverter = newTransferTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.ModifyMultisigAccountTransactionConverter = newModifyMultisigAccountTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.ModifyContractTransactionConverter = newModifyContractTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.RegisterNamespaceTransactionConverter = newRegisterNamespaceTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.LockFundsTransactionConverter = newLockFundsTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.SecretLockTransactionConverter = newSecretLockTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.SecretProofTransactionConverter = newSecretProofTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter)
	client.AggregateTransactionCosignatureConverter = newAggregateTransactionCosignatureConverter(client.AccountFactory)
	client.AggregateTransactionConverter = newAggregateTransactionConverter(client.AbstractTransactionConverter, client.TransactionInfoConverter, client.AggregateTransactionCosignatureConverter)
	client.MultisigCosignatoryModificationConverter = newMultisigCosignatoryModificationConverter(client.AccountFactory)
}

// DoNewRequest creates new request, Do it & return result in V
func (c *Client) DoNewRequest(ctx context.Context, method string, path string, body interface{}, v interface{}) (*http.Response, error) {
	req, err := c.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(ctx, req, v)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Do sends an API Request and returns a parsed response
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {

	// set the Context for this request
	req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 226 || resp.StatusCode < 200 {
		b := &bytes.Buffer{}
		b.ReadFrom(resp.Body)
		return nil, errors.New(b.String())
	}
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return resp, err
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {

	u, err := c.config.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func handleResponseStatusCode(resp *http.Response, codeToErrs map[int]error) error {
	if resp == nil {
		return ErrInternalError
	}

	if codeToErrs != nil {
		if err, ok := codeToErrs[resp.StatusCode]; ok {
			return err
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return ErrNotAcceptedResponseStatusCode
	}

	return nil
}
