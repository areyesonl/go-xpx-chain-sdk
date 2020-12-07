// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sdk

import (
	"fmt"
	"testing"

	"github.com/json-iterator/go"
	"github.com/proximax-storage/go-xpx-utils/mock"
	"github.com/proximax-storage/go-xpx-utils/tests"
	"github.com/stretchr/testify/assert"
)

func init() {
	jsoniter.RegisterTypeEncoder("*NamespaceIds", testNamespaceIDs)
	jsoniter.RegisterTypeDecoder("*NamespaceIds", ad)

	namespaceCorr.Levels = []*NamespaceId{testNamespaceId}
	namespaceNameCorr.NamespaceId = testNamespaceId

	mockServer.AddRouter(&mock.Router{
		Path:     fmt.Sprintf(namespaceRoute, testNamespaceId.toHexString()),
		RespBody: tplInfo,
	})
}

// test data
const (
	pageSize        = 32
	mosaicNamespace = "84b3552d375ffa4b"
)

var (
	namespaceClient = mockServer.getPublicTestClientUnsafe().Namespace
	testAddresses   = []*Address{
		{Address: "SDRDGFTDLLCB67D4HPGIMIHPNSRYRJRT7DOBGWZY"},
		{Address: "SBCPGZ3S2SCC3YHBBTYDCUZV4ZZEPHM2KGCP4QXX"},
	}
	testAddress = Address{Address: "SCASIIAPS6BSFEC66V6MU5ZGEVWM53BES5GYBGLE"}

	testNamespaceId  = newNamespaceIdPanic(9562080086528621131)
	testNamespaceIDs = &namespaceIds{
		List: []*NamespaceId{
			testNamespaceId,
		},
	}
	ad   = &namespaceIds{}
	meta = `"meta": {
			"active": true,
			"index": 0,
			"id": "5B55E02EACCB7B00015DB6EB"
			}`
	tplInfo = "{" + meta + `
			  ,
			  "namespace": {
				"namespaceId": [
				  929036875,
				  2226345261
				],
				"type": 0,
				"depth": 1,
				"level0": [
				  929036875,
				  2226345261
				],
    			"alias": {
      				"type": 1,
      				"mosaicId": [
        				1382215848,
        				1583663204
					]
    			},
				"owner": "321DE652C4D3362FC2DDF7800F6582F4A10CFEA134B81F8AB6E4BE78BBA4D18E",
				"ownerAddress": "904A1B7A7432C968202264C2CBDE0E8E5547EED3AD66E52BAC",
				"startHeight": [
				  1,
				  0
				],
				"endHeight": [
				  4294967295,
				  4294967295
				]
			  }
			}`

	namespaceCorr = &NamespaceInfo{
		NamespaceId: newNamespaceIdPanic(uint64DTO{929036875, 2226345261}.toUint64()),
		Active:      true,
		Depth:       1,
		TypeSpace:   Root,
		Alias: &NamespaceAlias{
			newMosaicIdPanic(uint64DTO{1382215848, 1583663204}.toUint64()),
			&Address{
				MijinTest,
				"SCJW742TNBMMX2UO4DVKXGP6T3CO6XXR6ZRWMVU2",
			},
			MosaicAliasType,
		},
		Owner: &PublicAccount{
			Address: &Address{
				Type:    MijinTest,
				Address: "SBFBW6TUGLEWQIBCMTBMXXQORZKUP3WTVVTOKK5M",
			},
			PublicKey: "321DE652C4D3362FC2DDF7800F6582F4A10CFEA134B81F8AB6E4BE78BBA4D18E",
		},
		EndHeight:   uint64DTO{4294967295, 4294967295}.toStruct(),
		StartHeight: uint64DTO{1, 0}.toStruct(),
		Parent:      nil,
	}

	namespaceNameCorr = &NamespaceName{
		NamespaceId: newNamespaceIdPanic(0),
		FullName:    "nem.xem",
	}

	tplInfoArr = "[" + tplInfo + "]"
)

func TestNamespaceService_GetNamespaceInfo(t *testing.T) {
	nsInfo, err := namespaceClient.GetNamespaceInfo(ctx, testNamespaceId)

	assert.Nilf(t, err, "NamespaceService.GetNamespace returned error: %s", err)
	tests.ValidateStringers(t, namespaceCorr, nsInfo)
}

func TestNamespaceService_GetNamespaceInfosFromAccount(t *testing.T) {
	mockServer.AddRouter(&mock.Router{
		Path:     fmt.Sprintf(namespacesFromAccountRoutes, testAddress.Address),
		RespBody: tplInfoArr,
	})

	nsInfoArr, err := namespaceClient.GetNamespaceInfosFromAccount(ctx, &testAddress, nil, pageSize)

	assert.Nilf(t, err, "NamespaceService.GetNamespaceInfosFromAccount returned error: %s", err)

	for _, nsInfo := range nsInfoArr {
		tests.ValidateStringers(t, namespaceCorr, nsInfo)
	}
}

func TestNamespaceService_GetNamespaceInfosFromAccounts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockServer.AddRouter(&mock.Router{
			Path:     namespacesFromAccountsRoute,
			RespBody: tplInfoArr,
		})

		nsInfoArr, err := namespaceClient.GetNamespaceInfosFromAccounts(ctx, testAddresses, nil, pageSize)

		assert.Nilf(t, err, "NamespaceService.GetNamespaceInfosFromAccounts returned error: %s", err)

		for _, nsInfo := range nsInfoArr {
			tests.ValidateStringers(t, namespaceCorr, nsInfo)
		}
	})

	t.Run("no test addresses", func(t *testing.T) {
		_, err := namespaceClient.GetNamespaceInfosFromAccounts(ctx, nil, nil, pageSize)

		assert.NotNil(t, err, "request with empty Addresses must return error")
	})
}

func TestNamespaceService_GetNamespaceNames(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockServer.AddRouter(&mock.Router{
			Path: namespaceNamesRoute,
			RespBody: `[
			  {
				"namespaceId": [
				  929036875,
				  2226345261
				],
				"name": "nem.xem"
			  }
			]`,
		})

		nsInfoArr, err := namespaceClient.GetNamespaceNames(ctx, []*NamespaceId{testNamespaceId})

		assert.Nilf(t, err, "NamespaceService.GetNamespaceNames returned error: %s", err)

		for _, nsInfo := range nsInfoArr {
			tests.ValidateStringers(t, namespaceNameCorr, nsInfo)
		}
	})

	t.Run("empty namespaceIds", func(t *testing.T) {
		_, err := namespaceClient.GetNamespaceNames(ctx, []*NamespaceId{})

		assert.Equal(t, ErrEmptyNamespaceIds, err, "request with empty NamespaceIds must return error")
	})
}
