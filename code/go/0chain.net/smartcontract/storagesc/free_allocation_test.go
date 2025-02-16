package storagesc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"0chain.net/smartcontract/provider"

	"0chain.net/smartcontract/stakepool/spenum"

	"github.com/0chain/common/core/currency"

	cstate "0chain.net/chaincore/chain/state"

	"0chain.net/core/encryption"

	"0chain.net/chaincore/chain/state/mocks"
	sci "0chain.net/chaincore/smartcontractinterface"
	"0chain.net/chaincore/transaction"
	"0chain.net/core/common"
	"github.com/0chain/common/core/util"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func mockSetValue(v interface{}) interface{} {
	return mock.MatchedBy(func(c interface{}) bool {
		cv := reflect.ValueOf(c)
		if cv.Kind() != reflect.Ptr {
			panic(fmt.Sprintf("%t must be a pointer, %v", v, cv.Kind()))
		}

		vv := reflect.ValueOf(v)
		if vv.Kind() == reflect.Ptr {
			if vv.Type() != cv.Type() {
				return false
			}
			cv.Elem().Set(vv.Elem())
		} else {
			if vv.Type() != cv.Elem().Type() {
				return false
			}

			cv.Elem().Set(vv)
		}
		return true
	})
}

func TestAddFreeStorageAssigner(t *testing.T) {
	const (
		mockCooperationId        = "mock cooperation id"
		mockPublicKey            = "mock public key"
		mockAnotherPublicKey     = "another mock public key"
		mockIndividualTokenLimit = 20
		mockTotalTokenLimit      = 3000
		mockNotOwner             = "mock not owner"
	)

	type args struct {
		ssc      *StorageSmartContract
		txn      *transaction.Transaction
		input    []byte
		balances cstate.StateContextI
	}
	type want struct {
		err    bool
		errMsg string
	}
	type parameters struct {
		clientId string
		info     newFreeStorageAssignerInfo
		exists   bool
		existing freeStorageAssigner
	}
	var conf = &Config{
		MaxIndividualFreeAllocation: zcnToBalance(mockIndividualTokenLimit),
		MaxTotalFreeAllocation:      zcnToBalance(mockTotalTokenLimit),
		OwnerId:                     owner,
		MaxBlobbersPerAllocation:    40,
	}

	setExpectations := func(t *testing.T, name string, p parameters, want want) args {
		var balances = &mocks.StateContextI{}
		var txn = &transaction.Transaction{
			ClientID: p.clientId,
		}
		var ssc = &StorageSmartContract{

			SmartContract: sci.NewSC(ADDRESS),
		}
		input, err := json.Marshal(p.info)
		require.NoError(t, err)

		balances.On("GetTrieNode", scConfigKey(ADDRESS),
			mockSetValue(conf)).Return(nil).Once()

		//var newRedeemed []freeStorageRedeemed
		if p.exists {
			balances.On(
				"GetTrieNode",
				freeStorageAssignerKey(ssc.ID, p.info.Name),
				mockSetValue(&p.existing),
			).Return(nil).Once()
		} else {
			balances.On(
				"GetTrieNode", freeStorageAssignerKey(ssc.ID, p.info.Name), mock.Anything,
			).Return(util.ErrValueNotPresent).Once()
		}

		balances.On("InsertTrieNode", freeStorageAssignerKey(ssc.ID, p.info.Name),
			&freeStorageAssigner{
				ClientId:        p.info.Name,
				PublicKey:       p.info.PublicKey,
				IndividualLimit: zcnToBalance(p.info.IndividualLimit),
				TotalLimit:      zcnToBalance(p.info.TotalLimit),
				CurrentRedeemed: p.existing.CurrentRedeemed,
				RedeemedNonces:  p.existing.RedeemedNonces,
			}).Return("", nil).Once()

		return args{ssc, txn, input, balances}
	}

	testCases := []struct {
		name       string
		parameters parameters
		want       want
	}{
		{
			name: "ok_new",
			parameters: parameters{
				clientId: owner,
				info: newFreeStorageAssignerInfo{
					Name:            mockCooperationId + "ok_new",
					PublicKey:       mockPublicKey,
					IndividualLimit: mockIndividualTokenLimit,
					TotalLimit:      mockTotalTokenLimit,
				},
				exists: false,
			},
		},
		{
			name: "ok_existing",
			parameters: parameters{
				clientId: owner,
				info: newFreeStorageAssignerInfo{
					Name:            mockCooperationId + "ok_existing",
					PublicKey:       mockPublicKey,
					IndividualLimit: mockIndividualTokenLimit,
					TotalLimit:      mockTotalTokenLimit,
				},
				exists: true,
				existing: freeStorageAssigner{
					ClientId:        mockCooperationId + "ok_existing",
					PublicKey:       mockAnotherPublicKey,
					IndividualLimit: mockIndividualTokenLimit / 2,
					TotalLimit:      mockTotalTokenLimit / 2,
					CurrentRedeemed: mockTotalTokenLimit / 4,
					RedeemedNonces:  []int64{20, 30, 50, 70, 110, 130, 170},
				},
			},
		},
		{
			name: "not_owner",
			parameters: parameters{
				clientId: mockNotOwner,
				info: newFreeStorageAssignerInfo{
					Name:            mockCooperationId + "ok_new",
					PublicKey:       mockPublicKey,
					IndividualLimit: mockIndividualTokenLimit,
					TotalLimit:      mockTotalTokenLimit,
				},
				exists: false,
			},
			want: want{
				true,
				"add_free_storage_assigner: unauthorized access - only the owner can access",
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			args := setExpectations(t, test.name, test.parameters, test.want)

			_, err := args.ssc.addFreeStorageAssigner(args.txn, args.input, args.balances)

			require.EqualValues(t, test.want.err, err != nil)
			if err != nil {
				require.EqualValues(t, test.want.errMsg, err.Error())
				return
			}
			require.True(t, mock.AssertExpectationsForObjects(t, args.balances))
		})
	}
}

func TestFreeAllocationRequest(t *testing.T) {
	const (
		mockCooperationId        = "mock cooperation id"
		mockNumBlobbers          = 10
		mockRecipient            = "mock recipient"
		mockNonce                = 7000
		mockUserPublicKey        = "mock user public key"
		mockTransactionHash      = "12345678"
		mockReadPoolFraction     = 0.2
		mockMinLock              = 10
		mockFreeTokens           = 5 * mockMinLock
		mockIndividualTokenLimit = mockFreeTokens + 1
		mockTotalTokenLimit      = mockIndividualTokenLimit * 300
		newSaSaved               = "new storage allocation saved"
	)
	var (
		mockMaxAnnualFreeAllocation = zcnToBalance(100354)
		timeUnit                    = 24 * 365 * time.Hour
		mockFreeAllocationSettings  = freeAllocationSettings{
			DataShards:       5,
			ParityShards:     5,
			Size:             123456,
			ReadPriceRange:   PriceRange{0, 5000},
			WritePriceRange:  PriceRange{0, 5000},
			ReadPoolFraction: mockReadPoolFraction,
		}
		mockAllBlobbers = &StorageNodes{}
		conf            = &Config{
			MinAllocSize:                 1027,
			MaxChallengeCompletionRounds: 720,
			MaxTotalFreeAllocation:       mockMaxAnnualFreeAllocation,
			FreeAllocationSettings:       mockFreeAllocationSettings,
			TimeUnit:                     timeUnit,
			ReadPool: &readPoolConfig{
				MinLock: mockMinLock,
			},
			MaxBlobbersPerAllocation: 40,
		}
		now = common.Timestamp(23000000)
	)

	blob := make([]string, mockNumBlobbers)
	for i := 0; i < mockNumBlobbers; i++ {
		blob[i] = strconv.Itoa(i)
		mockBlobber := &StorageNode{}
		mockBlobber.SetEntity(&storageNodeV2{
			Provider: provider.Provider{
				ID:              blob[i],
				ProviderType:    spenum.Blobber,
				LastHealthCheck: now - blobberHealthTime + 1,
			},
			Capacity:  536870912,
			Allocated: 73,
			Terms: Terms{
				ReadPrice: mockFreeAllocationSettings.ReadPriceRange.Max,
			},
		})
		mockAllBlobbers.Nodes.add(mockBlobber)
	}

	type args struct {
		ssc      *StorageSmartContract
		txn      *transaction.Transaction
		input    []byte
		balances cstate.StateContextI
	}
	type want struct {
		err    bool
		errMsg string
	}
	type parameters struct {
		assigner freeStorageAssigner
		marker   freeStorageMarker
	}

	setExpectations := func(t *testing.T, name string, p parameters, want want) args {
		balances := newTestBalances(t, false)
		// init config
		_, err := balances.InsertTrieNode(scConfigKey(ADDRESS), conf)
		require.NoError(t, err)

		_, err = balances.InsertTrieNode(freeStorageAssignerKey(ADDRESS, p.marker.Assigner), &p.assigner)
		require.NoError(t, err)

		var txn = &transaction.Transaction{
			ClientID:     p.marker.Recipient,
			ToClientID:   ADDRESS,
			PublicKey:    mockUserPublicKey,
			CreationDate: now,
		}
		txn.Value, err = currency.ParseZCN(p.marker.FreeTokens)
		require.NoError(t, err)

		txn.Hash = mockTransactionHash
		var ssc = &StorageSmartContract{
			SmartContract: sci.NewSC(ADDRESS),
		}

		p.marker.Signature, p.assigner.PublicKey = signFreeAllocationMarker(t, p.marker)

		inputBytes, err := json.Marshal(&p.marker)
		require.NoError(t, err)
		inputObj := freeStorageAllocationInput{
			RecipientPublicKey: mockUserPublicKey,
			Marker:             string(inputBytes),
		}
		input, err := json.Marshal(&inputObj)
		require.NoError(t, err)

		return args{ssc, txn, input, balances}
	}

	testCases := []struct {
		name       string
		parameters parameters
		want       want
	}{
		{
			name: "ok_no_previous",
			parameters: parameters{
				marker: freeStorageMarker{
					Assigner:   mockCooperationId + "ok_no_previous",
					Recipient:  mockRecipient,
					FreeTokens: mockFreeTokens,
					Nonce:      mockNonce,
					Blobbers:   blob,
				},
				assigner: freeStorageAssigner{
					ClientId:        mockCooperationId + "ok_no_previous",
					IndividualLimit: zcnToBalance(mockIndividualTokenLimit),
					TotalLimit:      zcnToBalance(mockTotalTokenLimit),
				},
			},
			want: want{
				err: false,
			},
		},
		{
			name: "Total_limit_exceeded",
			parameters: parameters{
				marker: freeStorageMarker{
					Assigner:   mockCooperationId + "Total_limit_exceeded",
					Recipient:  mockRecipient,
					FreeTokens: mockFreeTokens,
					Nonce:      mockNonce,
					Blobbers:   blob,
				},
				assigner: freeStorageAssigner{
					ClientId:        mockCooperationId + "Total_limit_exceeded",
					IndividualLimit: zcnToBalance(mockIndividualTokenLimit),
					TotalLimit:      zcnToBalance(mockTotalTokenLimit),
					CurrentRedeemed: zcnToBalance(mockTotalTokenLimit),
				},
			},
			want: want{
				true,
				"free_allocation_failed: marker verification failed: 153500000000000 exceeded total permitted free storage limit 153000000000000",
			},
		},
		{
			name: "individual_limit_exceeded",
			parameters: parameters{
				marker: freeStorageMarker{
					Assigner:   mockCooperationId + "individual_limit_exceeded",
					Recipient:  mockRecipient,
					FreeTokens: mockIndividualTokenLimit + 1,
					Nonce:      mockNonce,
					Blobbers:   blob,
				},
				assigner: freeStorageAssigner{
					ClientId:        mockCooperationId + "individual_limit_exceeded",
					PublicKey:       "",
					IndividualLimit: zcnToBalance(mockIndividualTokenLimit),
					TotalLimit:      zcnToBalance(mockTotalTokenLimit),
				},
			},
			want: want{
				true,
				"free_allocation_failed: marker verification failed: 520000000000 exceeded permitted free storage  510000000000",
			},
		},
		{
			name: "repeated_old_nonce",
			parameters: parameters{
				marker: freeStorageMarker{
					Assigner:   mockCooperationId + "repeated_old_nonce",
					Recipient:  mockRecipient,
					FreeTokens: mockFreeTokens,
					Nonce:      mockNonce,
					Blobbers:   blob,
				},
				assigner: freeStorageAssigner{
					ClientId:        mockCooperationId + "repeated_old_nonce",
					IndividualLimit: zcnToBalance(mockIndividualTokenLimit),
					TotalLimit:      zcnToBalance(mockTotalTokenLimit),
					RedeemedNonces:  []int64{190, mockNonce},
				},
			},
			want: want{
				true,
				"free_allocation_failed: marker verification failed: marker already redeemed, nonce: 7000",
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			args := setExpectations(t, test.name, test.parameters, test.want)

			_, err := args.ssc.freeAllocationRequest(args.txn, args.input, args.balances)

			if (test.want.err) != (err != nil) {
				require.EqualValues(t, test.want.err, err != nil, err)
			}

			if err != nil {
				require.EqualValues(t, test.want.errMsg, err.Error())
				return
			}
			require.True(t, mock.AssertExpectationsForObjects(t, args.balances))
		})
	}
}

func signFreeAllocationMarker(t *testing.T, frm freeStorageMarker) (string, string) {
	var ids string
	for _, b := range frm.Blobbers {
		ids += b
	}
	marker := fmt.Sprintf("%s:%f:%d:%s", frm.Recipient, frm.FreeTokens, frm.Nonce, ids)
	signatureScheme := encryption.NewBLS0ChainScheme()
	err := signatureScheme.GenerateKeys()
	require.NoError(t, err)
	signature, err := signatureScheme.Sign(hex.EncodeToString([]byte(marker)))
	require.NoError(t, err)
	return signature, signatureScheme.GetPublicKey()
}
