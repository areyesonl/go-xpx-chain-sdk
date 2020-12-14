// Copyright 2019 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sdk

import (
	"bytes"
	jsonLib "encoding/json"
	"fmt"
	"sync"
)

type DriveState uint8

const (
	NotStarted DriveState = iota
	Pending
	InProgress
	Finished
)

type PaymentInformation struct {
	Receiver *PublicAccount
	Amount   Amount
	Height   Height
}

func (info *PaymentInformation) String() string {
	return fmt.Sprintf(
		`{ "Receiver": %s, "Amount": %s, "Height": %s }`,
		info.Receiver,
		info.Amount,
		info.Height,
	)
}

type BillingDescription struct {
	Start    Height
	End      Height
	Payments []*PaymentInformation
}

func (desc *BillingDescription) String() string {
	return fmt.Sprintf(
		`
			"Start": %s,
			"End": %s,
			"Payments": %s,
		`,
		desc.Start,
		desc.End,
		desc.Payments,
	)
}

type ReplicatorInfo struct {
	Account                   *PublicAccount
	Start                     Height
	End                       Height
	Index                     int
	ActiveFilesWithoutDeposit map[Hash]bool
}

func (info *ReplicatorInfo) String() string {
	return fmt.Sprintf(
		`
			"Account": %s,
			"Start": %s,
			"End": %s,
			"Index": %d,
			"ActiveFilesWithoutDeposit": %+v,
		`,
		info.Account,
		info.Start,
		info.End,
		info.Index,
		info.ActiveFilesWithoutDeposit,
	)
}

type Drive struct {
	DriveAccount     *PublicAccount
	Start            Height
	State            DriveState
	OwnerAccount     *PublicAccount
	RootHash         *Hash
	Duration         Duration
	BillingPeriod    Duration
	BillingPrice     Amount
	DriveSize        StorageSize
	OccupiedSpace    StorageSize
	Replicas         uint16
	MinReplicators   uint16
	PercentApprovers uint8
	BillingHistory   []*BillingDescription
	Files            map[Hash]StorageSize
	Replicators      map[string]*ReplicatorInfo
	UploadPayments   []*PaymentInformation
}

func (drive *Drive) String() string {
	return fmt.Sprintf(
		`
			"DriveAccount": %s,
			"Start": %s,
			"State": %d,
			"OwnerAccount": %s,
			"RootHash": %s,
			"Duration": %d,
			"BillingPeriod": %d,
			"BillingPrice": %d,
			"DriveSize": %d,
			"OccupiedSpace": %d,
			"Replicas": %d,
			"MinReplicators": %d,
			"PercentApprovers": %d,
			"BillingHistory": %s,
			"Files": %s,
			"Replicators": %s,
			"UploadPayments": %s,
		`,
		drive.DriveAccount,
		drive.Start,
		drive.State,
		drive.OwnerAccount,
		drive.RootHash,
		drive.Duration,
		drive.BillingPeriod,
		drive.BillingPrice,
		drive.DriveSize,
		drive.OccupiedSpace,
		drive.Replicas,
		drive.MinReplicators,
		drive.PercentApprovers,
		drive.BillingHistory,
		drive.Files,
		drive.Replicators,
		drive.UploadPayments,
	)
}

type drivesPageDTO struct {
	Drives []jsonLib.RawMessage `json:"data"`

	Pagination struct {
		TotalEntries uint64 `json:"totalEntries"`
		PageNumber   uint64 `json:"pageNumber"`
		PageSize     uint64 `json:"pageSize"`
		TotalPages   uint64 `json:"totalPages"`
	} `json:"pagination"`
}

func (t *drivesPageDTO) toStruct(networkType NetworkType) (*DrivesPage, error) {
	var wg sync.WaitGroup
	page := &DrivesPage{
		Drives: make([]Drive, len(t.Drives)),
		Pagination: Pagination{
			TotalEntries: t.Pagination.TotalEntries,
			PageNumber:   t.Pagination.PageNumber,
			PageSize:     t.Pagination.PageSize,
			TotalPages:   t.Pagination.TotalPages,
		},
	}

	errs := make([]error, len(t.Drives))
	for i, t := range t.Drives {
		wg.Add(1)
		go func(i int, t jsonLib.RawMessage) {
			defer wg.Done()
			currDr, currErr := MapDrive(bytes.NewBuffer([]byte(t)), networkType)
			page.Drives[i], errs[i] = *currDr, currErr
		}(i, t)
	}

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return page, err
		}
	}

	return page, nil
}

func MapDrive(b *bytes.Buffer, networkType NetworkType) (*Drive, error) {
	dtoD := driveDTO{}

	err := json.Unmarshal(b.Bytes(), &dtoD)
	if err != nil {
		return nil, err
	}

	d, err := dtoD.toStruct(networkType)
	if err != nil {
		return nil, err
	}

	return d, nil
}

type DrivesPage struct {
	Drives     []Drive
	Pagination Pagination
}

type DriveFilters struct {
	Start     uint64
	StartType DriveStartValueType
	States    []uint32
}

type DriveSortOptions struct {
	SortField string
	Direction DriveSortDirection
}

type DriveSortDirection uint8

const (
	ASC  DriveSortDirection = 0
	DESC DriveSortDirection = 1
)

func (sD DriveSortDirection) String() string {
	return [...]string{"asc", "desc"}[sD]
}

type DriveStartValueType uint8

const (
	Start     DriveStartValueType = 0
	FromStart DriveStartValueType = 1
	ToStart   DriveStartValueType = 2
)

func (vT DriveStartValueType) String() string {
	return [...]string{"start", "fromStart", "toStart"}[vT]
}

// Prepare Drive Transaction
type PrepareDriveTransaction struct {
	AbstractTransaction
	Owner            *PublicAccount
	Duration         Duration
	BillingPeriod    Duration
	BillingPrice     Amount
	DriveSize        StorageSize
	Replicas         uint16
	MinReplicators   uint16
	PercentApprovers uint8
}

// Join Drive Transaction

type JoinToDriveTransaction struct {
	AbstractTransaction
	DriveKey *PublicAccount
}

type File struct {
	FileHash *Hash
}

func (file *File) String() string {
	return fmt.Sprintf(
		`
			"FileHash": %s,
		`,
		file.FileHash,
	)
}

type Action struct {
	FileHash *Hash
	FileSize StorageSize
}

func (action *Action) String() string {
	return fmt.Sprintf(
		`
			"FileHash": %s,
			"FileSize": %s,
		`,
		action.FileHash,
		action.FileSize,
	)
}

type DriveFileSystemTransaction struct {
	AbstractTransaction
	DriveKey      string
	NewRootHash   *Hash
	OldRootHash   *Hash
	AddActions    []*Action
	RemoveActions []*Action
}

// Files Deposit Transaction
type FilesDepositTransaction struct {
	AbstractTransaction
	DriveKey *PublicAccount
	Files    []*File
}

// End Drive Transaction

type EndDriveTransaction struct {
	AbstractTransaction
	DriveKey *PublicAccount
}

type UploadInfo struct {
	Participant  *PublicAccount
	UploadedSize Amount
}

func (info *UploadInfo) String() string {
	return fmt.Sprintf(
		`
			"Participant": %s,
			"UploadedSize": %s,
		`,
		info.Participant,
		info.UploadedSize,
	)
}

// Drive Files Reward Transaction

type DriveFilesRewardTransaction struct {
	AbstractTransaction
	UploadInfos []*UploadInfo
}

// Start Drive Verification Transaction

type StartDriveVerificationTransaction struct {
	AbstractTransaction
	DriveKey *PublicAccount
}

type FailureVerification struct {
	Replicator  *PublicAccount
	BlochHashes []*Hash
}

func (fail *FailureVerification) Size() int {
	return SizeSize + len(fail.BlochHashes)*Hash256 + KeySize
}

func (fail *FailureVerification) String() string {
	return fmt.Sprintf(
		`
			"Replicator": %s,
			"BlochHashes": %s,
		`,
		fail.Replicator,
		fail.BlochHashes,
	)
}

// End Drive Verification Transaction

type EndDriveVerificationTransaction struct {
	AbstractTransaction
	Failures []*FailureVerification
}

type VerificationStatus struct {
	Active    bool
	Available bool
}

// Start File Download Transaction

type DownloadFile = Action

type StartFileDownloadTransaction struct {
	AbstractTransaction
	Drive *PublicAccount
	Files []*DownloadFile
}

type DownloadInfo struct {
	OperationToken *Hash
	DriveAccount   *PublicAccount
	FileRecipient  *PublicAccount
	Height         Height
	Files          []*DownloadFile
}

// End File Download Transaction

type EndFileDownloadTransaction struct {
	AbstractTransaction
	Recipient      *PublicAccount
	OperationToken *Hash
	Files          []*DownloadFile
}
