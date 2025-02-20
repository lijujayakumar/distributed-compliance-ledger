// Copyright 2020 DSR Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//nolint:testpackage,lll
package compliancetest

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	testconstants "github.com/zigbee-alliance/distributed-compliance-ledger/integration_tests/constants"
	"github.com/zigbee-alliance/distributed-compliance-ledger/x/auth"
	"github.com/zigbee-alliance/distributed-compliance-ledger/x/compliancetest/internal/types"
	"github.com/zigbee-alliance/distributed-compliance-ledger/x/model"
)

type TestSetup struct {
	Cdc                  *amino.Codec
	Ctx                  sdk.Context
	CompliancetestKeeper Keeper
	authKeeper           auth.Keeper
	ModelKeeper          model.Keeper
	Handler              sdk.Handler
	Querier              sdk.Querier
	TestHouse            sdk.AccAddress
}

func Setup() TestSetup {
	// Init Codec
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	// Init KVSore
	db := dbm.NewMemDB()

	dbStore := store.NewCommitMultiStore(db)

	complianceKey := sdk.NewKVStoreKey(StoreKey)
	dbStore.MountStoreWithDB(complianceKey, sdk.StoreTypeIAVL, nil)

	authKey := sdk.NewKVStoreKey(auth.StoreKey)
	dbStore.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, nil)

	modelKey := sdk.NewKVStoreKey(model.StoreKey)
	dbStore.MountStoreWithDB(modelKey, sdk.StoreTypeIAVL, nil)

	_ = dbStore.LoadLatestVersion()

	// Init Keepers
	compliancetestKeeper := NewKeeper(complianceKey, cdc)
	authKeeper := auth.NewKeeper(authKey, cdc)
	modelKeeper := model.NewKeeper(modelKey, cdc)

	// Create context
	ctx := sdk.NewContext(dbStore, abci.Header{ChainID: testconstants.ChainID}, false, log.NewNopLogger())

	// Create Handler and Querier
	querier := NewQuerier(compliancetestKeeper)
	handler := NewHandler(compliancetestKeeper, modelKeeper, authKeeper)

	account := auth.NewAccount(testconstants.Address1, testconstants.PubKey1, auth.AccountRoles{auth.TestHouse}, testconstants.VendorID1)
	account.AccountNumber = authKeeper.GetNextAccountNumber(ctx)
	authKeeper.SetAccount(ctx, account)

	setup := TestSetup{
		Cdc:                  cdc,
		Ctx:                  ctx,
		CompliancetestKeeper: compliancetestKeeper,
		ModelKeeper:          modelKeeper,
		authKeeper:           authKeeper,
		Handler:              handler,
		Querier:              querier,
		TestHouse:            account.Address,
	}

	return setup
}

func TestMsgAddTestingResult(signer sdk.AccAddress, vid uint16, pid uint16, softwareVersion uint32, softwareVersionString string) MsgAddTestingResult {
	return MsgAddTestingResult{
		VID:                   vid,
		PID:                   pid,
		SoftwareVersion:       softwareVersion,
		SoftwareVersionString: softwareVersionString,
		TestResult:            testconstants.TestResult,
		TestDate:              testconstants.TestDate,
		Signer:                signer,
	}
}

func CheckTestingResult(t *testing.T, receivedTestingResult types.TestingResultItem,
	expectedTestingResult types.MsgAddTestingResult) {
	require.Equal(t, receivedTestingResult.Owner, expectedTestingResult.Signer)
	require.Equal(t, receivedTestingResult.TestResult, expectedTestingResult.TestResult)
	require.Equal(t, receivedTestingResult.TestDate, expectedTestingResult.TestDate)
}
