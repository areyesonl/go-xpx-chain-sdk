package sdk

import (
	"encoding/hex"
	"fmt"
	"github.com/google/flatbuffers/go"
	"github.com/proximax-storage/go-xpx-chain-sdk/transactions"
)

func NewPrepareDriveTransaction(
	deadline *Deadline,
	owner *PublicAccount,
	duration Duration,
	billingPeriod Duration,
	billingPrice Amount,
	driveSize StorageSize,
	replicas uint16,
	minReplicators uint16,
	percentApprovers uint8,
	networkType NetworkType,
) (*PrepareDriveTransaction, error) {

	mctx := PrepareDriveTransaction{
		AbstractTransaction: AbstractTransaction{
			Version:     PrepareDriveVersion,
			Deadline:    deadline,
			Type:        PrepareDrive,
			NetworkType: networkType,
		},
		Owner:            owner,
		Duration:         duration,
		BillingPeriod:    billingPeriod,
		BillingPrice:     billingPrice,
		DriveSize:        driveSize,
		Replicas:         replicas,
		MinReplicators:   minReplicators,
		PercentApprovers: percentApprovers,
	}

	return &mctx, nil
}

func (tx *PrepareDriveTransaction) GetAbstractTransaction() *AbstractTransaction {
	return &tx.AbstractTransaction
}

func (tx *PrepareDriveTransaction) String() string {
	return fmt.Sprintf(
		`
			"AbstractTransaction": %s,
			"Owner": %s,
			"Duration": %d,
			"BillingPeriod": %d,
			"BillingPrice": %d,
			"DriveSize": %d,
			"Replicas": %d,
			"MinReplicators": %d,
			"PercentApprovers": %d,
		`,
		tx.AbstractTransaction.String(),
		tx.Owner,
		tx.Duration,
		tx.BillingPeriod,
		tx.BillingPrice,
		tx.DriveSize,
		tx.Replicas,
		tx.MinReplicators,
		tx.PercentApprovers,
	)
}

func (tx *PrepareDriveTransaction) generateBytes() ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)

	v, signatureV, signerV, deadlineV, fV, err := tx.AbstractTransaction.generateVectors(builder)
	if err != nil {
		return nil, err
	}

	ownerB, err := hex.DecodeString(tx.Owner.PublicKey)
	if err != nil {
		return nil, err
	}

	ownerV := transactions.TransactionBufferCreateByteVector(builder, ownerB)
	durationV := transactions.TransactionBufferCreateUint32Vector(builder, tx.Duration.toArray())
	billingPeriodV := transactions.TransactionBufferCreateUint32Vector(builder, tx.BillingPeriod.toArray())
	billingPriceV := transactions.TransactionBufferCreateUint32Vector(builder, tx.BillingPrice.toArray())
	driveSizeV := transactions.TransactionBufferCreateUint32Vector(builder, tx.DriveSize.toArray())

	transactions.PrepareDriveTransactionBufferStart(builder)
	transactions.TransactionBufferAddSize(builder, tx.Size())
	tx.AbstractTransaction.buildVectors(builder, v, signatureV, signerV, deadlineV, fV)

	transactions.PrepareDriveTransactionBufferAddOwner(builder, ownerV)
	transactions.PrepareDriveTransactionBufferAddDuration(builder, durationV)
	transactions.PrepareDriveTransactionBufferAddBillingPeriod(builder, billingPeriodV)
	transactions.PrepareDriveTransactionBufferAddBillingPrice(builder, billingPriceV)
	transactions.PrepareDriveTransactionBufferAddDriveSize(builder, driveSizeV)

	transactions.PrepareDriveTransactionBufferAddReplicas(builder, tx.Replicas)
	transactions.PrepareDriveTransactionBufferAddMinReplicators(builder, tx.MinReplicators)
	transactions.PrepareDriveTransactionBufferAddPercentApprovers(builder, tx.PercentApprovers)
	t := transactions.TransactionBufferEnd(builder)
	builder.Finish(t)

	return prepareDriveTransactionSchema().serialize(builder.FinishedBytes()), nil
}

func (tx *PrepareDriveTransaction) Size() int {
	return PrepareDriveHeaderSize
}

type prepareDriveTransactionDTO struct {
	Tx struct {
		abstractTransactionDTO
		Owner            string    `json:"owner"`
		Duration         uint64DTO `json:"duration"`
		BillingPeriod    uint64DTO `json:"billingPeriod"`
		BillingPrice     uint64DTO `json:"billingPrice"`
		DriveSize        uint64DTO `json:"driveSize"`
		Replicas         uint16    `json:"replicas"`
		MinReplicators   uint16    `json:"minReplicators"`
		PercentApprovers uint8     `json:"percentApprovers"`
	} `json:"transaction"`
	TDto transactionInfoDTO `json:"meta"`
}

func (dto *prepareDriveTransactionDTO) toStruct() (Transaction, error) {
	info, err := dto.TDto.toStruct()
	if err != nil {
		return nil, err
	}

	atx, err := dto.Tx.abstractTransactionDTO.toStruct(info)
	if err != nil {
		return nil, err
	}

	owner, err := NewAccountFromPublicKey(dto.Tx.Owner, atx.NetworkType)
	if err != nil {
		return nil, err
	}

	return &PrepareDriveTransaction{
		*atx,
		owner,
		dto.Tx.Duration.toStruct(),
		dto.Tx.BillingPeriod.toStruct(),
		dto.Tx.BillingPrice.toStruct(),
		dto.Tx.DriveSize.toStruct(),
		dto.Tx.Replicas,
		dto.Tx.MinReplicators,
		dto.Tx.PercentApprovers,
	}, nil
}

func NewJoinToDriveTransaction(
	deadline *Deadline,
	driveKey *PublicAccount,
	networkType NetworkType,
) (*JoinToDriveTransaction, error) {

	tx := JoinToDriveTransaction{
		AbstractTransaction: AbstractTransaction{
			Version:     JoinToDriveVersion,
			Deadline:    deadline,
			Type:        JoinToDrive,
			NetworkType: networkType,
		},
		DriveKey: driveKey,
	}

	return &tx, nil
}

func (tx *JoinToDriveTransaction) GetAbstractTransaction() *AbstractTransaction {
	return &tx.AbstractTransaction
}

func (tx *JoinToDriveTransaction) String() string {
	return fmt.Sprintf(
		`
			"AbstractTransaction": %s,
			"DriveKey": %s,
		`,
		tx.AbstractTransaction.String(),
		tx.DriveKey,
	)
}

func (tx *JoinToDriveTransaction) generateBytes() ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)

	v, signatureV, signerV, deadlineV, fV, err := tx.AbstractTransaction.generateVectors(builder)
	if err != nil {
		return nil, err
	}

	b, err := hex.DecodeString(tx.DriveKey.PublicKey)
	if err != nil {
		return nil, err
	}

	hV := transactions.TransactionBufferCreateByteVector(builder, b)

	transactions.JoinToDriveTransactionBufferStart(builder)
	transactions.TransactionBufferAddSize(builder, tx.Size())
	tx.AbstractTransaction.buildVectors(builder, v, signatureV, signerV, deadlineV, fV)

	transactions.JoinToDriveTransactionBufferAddDriveKey(builder, hV)

	t := transactions.TransactionBufferEnd(builder)
	builder.Finish(t)

	return joinDriveTransactionSchema().serialize(builder.FinishedBytes()), nil
}

func (tx *JoinToDriveTransaction) Size() int {
	return JoinToDriveHeaderSize
}

type joinToDriveTransactionDTO struct {
	Tx struct {
		abstractTransactionDTO
		DriveKey string `json:"driveKey"`
	} `json:"transaction"`
	TDto transactionInfoDTO `json:"meta"`
}

func (dto *joinToDriveTransactionDTO) toStruct() (Transaction, error) {
	info, err := dto.TDto.toStruct()
	if err != nil {
		return nil, err
	}

	atx, err := dto.Tx.abstractTransactionDTO.toStruct(info)
	if err != nil {
		return nil, err
	}

	acc, err := NewAccountFromPublicKey(dto.Tx.DriveKey, atx.NetworkType)
	if err != nil {
		return nil, err
	}

	return &JoinToDriveTransaction{
		*atx,
		acc,
	}, nil
}

func NewDriveFileSystemTransaction(
	deadline *Deadline,
	driveKey *PublicAccount,
	newRootHash *Hash,
	oldRootHash *Hash,
	addActions []*AddAction,
	removeActions []*RemoveAction,
	networkType NetworkType,
) (*DriveFileSystemTransaction, error) {

	tx := DriveFileSystemTransaction{
		AbstractTransaction: AbstractTransaction{
			Version:     DriveFileSystemVersion,
			Deadline:    deadline,
			Type:        DriveFileSystem,
			NetworkType: networkType,
		},
		DriveKey:      driveKey,
		NewRootHash:   newRootHash,
		OldRootHash:   oldRootHash,
		AddActions:    addActions,
		RemoveActions: removeActions,
	}

	return &tx, nil
}

func (tx *DriveFileSystemTransaction) GetAbstractTransaction() *AbstractTransaction {
	return &tx.AbstractTransaction
}

func (tx *DriveFileSystemTransaction) String() string {
	return fmt.Sprintf(
		`
			"AbstractTransaction": %s,
			"DriveKey": %s,
			"NewRootHash": %s,
			"OldRootHash": %s,
			"AddActions": %s,
			"RemoveActions": %s,
		`,
		tx.AbstractTransaction.String(),
		tx.DriveKey,
		tx.NewRootHash,
		tx.OldRootHash,
		tx.AddActions,
		tx.RemoveActions,
	)
}

func addActionsToArrayToBuffer(builder *flatbuffers.Builder, addActions []*AddAction) (flatbuffers.UOffsetT, error) {
	msb := make([]flatbuffers.UOffsetT, len(addActions))
	for i, m := range addActions {

		rhV := transactions.TransactionBufferCreateByteVector(builder, m.FileHash[:])
		sizeDV := transactions.TransactionBufferCreateUint32Vector(builder, m.FileSize.toArray())
		transactions.AddActionBufferStart(builder)
		transactions.AddActionBufferAddFileHash(builder, rhV)
		transactions.AddActionBufferAddFileSize(builder, sizeDV)
		msb[i] = transactions.TransactionBufferEnd(builder)
	}

	return transactions.TransactionBufferCreateUOffsetVector(builder, msb), nil
}

func removeActionsToArrayToBuffer(builder *flatbuffers.Builder, removeActions []*RemoveAction) (flatbuffers.UOffsetT, error) {
	msb := make([]flatbuffers.UOffsetT, len(removeActions))
	for i, m := range removeActions {

		rhV := transactions.TransactionBufferCreateByteVector(builder, m.FileHash[:])
		transactions.RemoveActionBufferStart(builder)
		transactions.RemoveActionBufferAddFileHash(builder, rhV)
		msb[i] = transactions.TransactionBufferEnd(builder)
	}
	return transactions.TransactionBufferCreateUOffsetVector(builder, msb), nil
}

func (tx *DriveFileSystemTransaction) generateBytes() ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)

	v, signatureV, signerV, deadlineV, fV, err := tx.AbstractTransaction.generateVectors(builder)
	if err != nil {
		return nil, err
	}

	driveKeyB, err := hex.DecodeString(tx.DriveKey.PublicKey)
	if err != nil {
		return nil, err
	}

	driveV := transactions.TransactionBufferCreateByteVector(builder, driveKeyB)
	rhV := transactions.TransactionBufferCreateByteVector(builder, tx.NewRootHash[:])

	xorRootHash := tx.NewRootHash.Xor(tx.OldRootHash)
	xhV := transactions.TransactionBufferCreateByteVector(builder, xorRootHash[:])

	addActionsV, err := addActionsToArrayToBuffer(builder, tx.AddActions)
	if err != nil {
		return nil, err
	}

	removeActionsV, err := removeActionsToArrayToBuffer(builder, tx.RemoveActions)
	if err != nil {
		return nil, err
	}

	transactions.DriveFileSystemTransactionBufferStart(builder)
	transactions.TransactionBufferAddSize(builder, tx.Size())
	tx.AbstractTransaction.buildVectors(builder, v, signatureV, signerV, deadlineV, fV)

	transactions.DriveFileSystemTransactionBufferAddDriveKey(builder, driveV)
	transactions.DriveFileSystemTransactionBufferAddRootHash(builder, rhV)
	transactions.DriveFileSystemTransactionBufferAddXorRootHash(builder, xhV)

	transactions.DriveFileSystemTransactionBufferAddAddActionsCount(builder, uint16(len(tx.AddActions)))
	transactions.DriveFileSystemTransactionBufferAddRemoveActionsCount(builder, uint16(len(tx.RemoveActions)))

	transactions.DriveFileSystemTransactionBufferAddAddActions(builder, addActionsV)
	transactions.DriveFileSystemTransactionBufferAddRemoveActions(builder, removeActionsV)

	t := transactions.TransactionBufferEnd(builder)
	builder.Finish(t)

	return driveFileSystemTransactionSchema().serialize(builder.FinishedBytes()), nil
}

func (tx *DriveFileSystemTransaction) Size() int {
	return DriveFileSystemHeaderSize + len(tx.AddActions) + len(tx.RemoveActions)
}

type driveFileSystemAddActionDTO struct {
	FileHash hashDto   `json:"fileHash"`
	FileSize uint64DTO `json:"fileSize"`
}

type driveFileSystemRemoveActionDTO struct {
	FileHash hashDto `json:"fileHash"`
}

type driveFileSystemTransactionDTO struct {
	Tx struct {
		abstractTransactionDTO
		DriveKey           string                            `json:"driveKey"`
		RootHash           hashDto                           `json:"rootHash"`
		XorRootHash        hashDto                           `json:"xorRootHash"`
		AddActionsCount    uint16                            `json:"addActionsCount"`
		RemoveActionsCount uint16                            `json:"removeActionsCount"`
		AddActions         []*driveFileSystemAddActionDTO    `json:"addActions"`
		RemoveActions      []*driveFileSystemRemoveActionDTO `json:"removeActions"`
	} `json:"transaction"`
	TDto transactionInfoDTO `json:"meta"`
}

func (dto *driveFileSystemTransactionDTO) toStruct() (Transaction, error) {
	info, err := dto.TDto.toStruct()
	if err != nil {
		return nil, err
	}

	atx, err := dto.Tx.abstractTransactionDTO.toStruct(info)
	if err != nil {
		return nil, err
	}
	driveKey, err := NewAccountFromPublicKey(dto.Tx.DriveKey, atx.NetworkType)
	if err != nil {
		return nil, err
	}

	rHash, err := dto.Tx.RootHash.Hash()
	if err != nil {
		return nil, err
	}

	xorRootHash, err := dto.Tx.RootHash.Hash()
	if err != nil {
		return nil, err
	}

	addActs, err := addActionsDTOArrayToStruct(dto.Tx.AddActions)
	if err != nil {
		return nil, err
	}

	removeActs, err := removeActionsDTOArrayToStruct(dto.Tx.RemoveActions)
	if err != nil {
		return nil, err
	}

	return &DriveFileSystemTransaction{
		*atx,
		driveKey,
		rHash,
		xorRootHash.Xor(rHash),
		addActs,
		removeActs,
	}, nil
}

func addActionsDTOArrayToStruct(addAction []*driveFileSystemAddActionDTO) ([]*AddAction, error) {
	acts := make([]*AddAction, len(addAction))
	var err error = nil
	for i, m := range addAction {
		h, err := m.FileHash.Hash()
		if err != nil {
			return nil, err
		}

		s := m.FileSize.toUint64()

		acts[i] = &AddAction{
			File{
				FileHash: h,
			},
			StorageSize(s),
		}

	}

	return acts, err
}

func removeActionsDTOArrayToStruct(removeAction []*driveFileSystemRemoveActionDTO) ([]*RemoveAction, error) {
	removes := make([]*RemoveAction, len(removeAction))
	var err error = nil
	for i, m := range removeAction {
		h, err := m.FileHash.Hash()
		if err != nil {
			return nil, err
		}
		removes[i] = &RemoveAction{
			File{
				FileHash: h,
			},
		}

	}

	return removes, err
}

func NewFilesDepositTransaction(
	deadline *Deadline,
	driveKey *PublicAccount,
	files []*File,
	networkType NetworkType,
) (*FilesDepositTransaction, error) {

	tx := FilesDepositTransaction{
		AbstractTransaction: AbstractTransaction{
			Version:     FilesDepositVersion,
			Deadline:    deadline,
			Type:        FilesDeposit,
			NetworkType: networkType,
		},
		DriveKey: driveKey,
		Files:    files,
	}

	return &tx, nil
}

func (tx *FilesDepositTransaction) GetAbstractTransaction() *AbstractTransaction {
	return &tx.AbstractTransaction
}

func (tx *FilesDepositTransaction) String() string {
	return fmt.Sprintf(
		`
			"AbstractTransaction": %s,
			"DriveKey": %s,
			"Files": %s,
		`,
		tx.AbstractTransaction.String(),
		tx.DriveKey,
		tx.Files,
	)
}

func fileToArrayToBuffer(builder *flatbuffers.Builder, addActions []*File) (flatbuffers.UOffsetT, error) {
	msb := make([]flatbuffers.UOffsetT, len(addActions))
	for i, m := range addActions {

		rhV := transactions.TransactionBufferCreateByteVector(builder, m.FileHash[:])
		transactions.FileBufferStart(builder)
		transactions.AddActionBufferAddFileHash(builder, rhV)
		msb[i] = transactions.TransactionBufferEnd(builder)
	}

	return transactions.TransactionBufferCreateUOffsetVector(builder, msb), nil
}

func (tx *FilesDepositTransaction) generateBytes() ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)

	v, signatureV, signerV, deadlineV, fV, err := tx.AbstractTransaction.generateVectors(builder)
	if err != nil {
		return nil, err
	}

	b, err := hex.DecodeString(tx.DriveKey.PublicKey)
	if err != nil {
		return nil, err
	}

	hV := transactions.TransactionBufferCreateByteVector(builder, b)

	flsV, err := fileToArrayToBuffer(builder, tx.Files)
	if err != nil {
		return nil, err
	}

	transactions.DriveFileSystemTransactionBufferStart(builder)
	transactions.TransactionBufferAddSize(builder, tx.Size())
	tx.AbstractTransaction.buildVectors(builder, v, signatureV, signerV, deadlineV, fV)

	transactions.FilesDepositTransactionBufferAddDriveKey(builder, hV)

	transactions.FilesDepositTransactionBufferAddFilesCount(builder, uint16(len(tx.Files)))
	transactions.FilesDepositTransactionBufferAddFiles(builder, flsV)

	t := transactions.TransactionBufferEnd(builder)
	builder.Finish(t)

	return filesDepositTransactionSchema().serialize(builder.FinishedBytes()), nil
}

func (tx *FilesDepositTransaction) Size() int {
	return FilesDepositHeaderSize + len(tx.Files)
}

type fileDTO struct {
	FileHash hashDto `json:"fileHash"`
}

type filesDepositTransactionDTO struct {
	Tx struct {
		abstractTransactionDTO
		DriveKey   string     `json:"driveKey"`
		FilesCount uint16     `json:"filesCount"`
		Files      []*fileDTO `json:"files"`
	} `json:"transaction"`
	TDto transactionInfoDTO `json:"meta"`
}

func (dto *filesDepositTransactionDTO) toStruct() (Transaction, error) {
	info, err := dto.TDto.toStruct()
	if err != nil {
		return nil, err
	}

	atx, err := dto.Tx.abstractTransactionDTO.toStruct(info)
	if err != nil {
		return nil, err
	}

	fls, err := filesDTOArrayToStruct(dto.Tx.Files)
	if err != nil {
		return nil, err
	}

	acc, err := NewAccountFromPublicKey(dto.Tx.DriveKey, atx.NetworkType)
	if err != nil {
		return nil, err
	}

	return &FilesDepositTransaction{
		*atx,
		acc,
		fls,
	}, nil
}

func filesDTOArrayToStruct(files []*fileDTO) ([]*File, error) {
	filesResult := make([]*File, len(files))
	var err error = nil
	for i, m := range files {
		h, err := m.FileHash.Hash()
		if err != nil {
			return nil, err
		}
		filesResult[i] = &File{
			FileHash: h,
		}

	}

	return filesResult, err
}

func NewEndDriveTransaction(
	deadline *Deadline,
	driveKey *PublicAccount,
	networkType NetworkType,
) (*EndDriveTransaction, error) {

	tx := EndDriveTransaction{
		AbstractTransaction: AbstractTransaction{
			Version:     EndDriveVersion,
			Deadline:    deadline,
			Type:        EndDrive,
			NetworkType: networkType,
		},
		DriveKey: driveKey,
	}

	return &tx, nil
}

func (tx *EndDriveTransaction) GetAbstractTransaction() *AbstractTransaction {
	return &tx.AbstractTransaction
}

func (tx *EndDriveTransaction) String() string {
	return fmt.Sprintf(
		`
			"AbstractTransaction": %s,
			"DriveKey": %s,
		`,
		tx.AbstractTransaction.String(),
		tx.DriveKey,
	)
}

func (tx *EndDriveTransaction) generateBytes() ([]byte, error) {
	builder := flatbuffers.NewBuilder(0)

	v, signatureV, signerV, deadlineV, fV, err := tx.AbstractTransaction.generateVectors(builder)
	if err != nil {
		return nil, err
	}

	transactions.EndDriveTransactionBufferStart(builder)
	transactions.TransactionBufferAddSize(builder, tx.Size())
	tx.AbstractTransaction.buildVectors(builder, v, signatureV, signerV, deadlineV, fV)
	t := transactions.TransactionBufferEnd(builder)
	builder.Finish(t)

	return endDriveTransactionSchema().serialize(builder.FinishedBytes()), nil
}

func (tx *EndDriveTransaction) Size() int {
	return EndDriveHeaderSize
}

type endDriveTransactionDTO struct {
	Tx struct {
		abstractTransactionDTO
		DriveKey string `json:"driveKey"`
	} `json:"transaction"`
	TDto transactionInfoDTO `json:"meta"`
}

func (dto *endDriveTransactionDTO) toStruct() (Transaction, error) {
	info, err := dto.TDto.toStruct()
	if err != nil {
		return nil, err
	}

	atx, err := dto.Tx.abstractTransactionDTO.toStruct(info)
	if err != nil {
		return nil, err
	}

	driveKey, err := NewAccountFromPublicKey(dto.Tx.DriveKey, atx.NetworkType)
	if err != nil {
		return nil, err
	}

	return &EndDriveTransaction{
		*atx,
		driveKey,
	}, nil
}
