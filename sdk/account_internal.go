package sdk

import (
	"encoding/base32"
	"encoding/hex"
	"github.com/proximax-storage/nem2-crypto-go"
)

var addressNet = map[uint8]NetworkType{
	'M': Mijin,
	'S': MijinTest,
	'X': Public,
	'V': PublicTest,
	'Z': Private,
	'W': PrivateTest,
}

type accountInfoDTO struct {
	Account struct {
		Address          string         `json:"address"`
		AddressHeight    uint64DTO      `json:"addressHeight"`
		PublicKey        string         `json:"publicKey"`
		PublicKeyHeight  uint64DTO      `json:"publicKeyHeight"`
		Importance       uint64DTO      `json:"importance"`
		ImportanceHeight uint64DTO      `json:"importanceHeight"`
		Mosaics          []*mosaicDTO   `json:"mosaics"`
		Reputation       *reputationDTO `json:"reputation"`
	} `json:"account"`
}

type reputationDTO struct {
	PositiveInteractions uint64DTO `json:"positiveInteractions"`
	NegativeInteractions uint64DTO `json:"negativeInteractions"`
}

func (ref *reputationDTO) toFloat(repConfig *reputationConfig) float64 {
	posInter := ref.PositiveInteractions.toBigInt().Uint64()
	negInter := ref.NegativeInteractions.toBigInt().Uint64()

	if posInter < repConfig.minInteractions {
		return repConfig.defaultReputation
	}

	rep := (posInter - negInter) / posInter

	return float64(rep)
}

func (dto *accountInfoDTO) toStruct(repConfig *reputationConfig) (*AccountInfo, error) {
	var (
		ms  = make([]*Mosaic, len(dto.Account.Mosaics))
		err error
	)

	for idx, m := range dto.Account.Mosaics {
		ms[idx], err = m.toStruct()
		if err != nil {
			return nil, err
		}
	}

	add, err := NewAddressFromEncoded(dto.Account.Address)
	if err != nil {
		return nil, err
	}

	acc := &AccountInfo{
		Address:          add,
		AddressHeight:    dto.Account.AddressHeight.toBigInt(),
		PublicKey:        dto.Account.PublicKey,
		PublicKeyHeight:  dto.Account.PublicKeyHeight.toBigInt(),
		Importance:       dto.Account.Importance.toBigInt(),
		ImportanceHeight: dto.Account.ImportanceHeight.toBigInt(),
		Mosaics:          ms,
		Reputation:       repConfig.defaultReputation,
	}

	if dto.Account.Reputation != nil {
		acc.Reputation = dto.Account.Reputation.toFloat(repConfig)
	}

	return acc, nil
}

type accountInfoDTOs []*accountInfoDTO

func (a accountInfoDTOs) toStruct(repConfig *reputationConfig) ([]*AccountInfo, error) {
	var (
		accountInfos = make([]*AccountInfo, len(a))
		err          error
	)

	for idx, dto := range a {
		accountInfos[idx], err = dto.toStruct(repConfig)
		if err != nil {
			return nil, err
		}
	}

	return accountInfos, nil
}

type multisigAccountInfoDTO struct {
	Multisig struct {
		Account          string   `json:"account"`
		MinApproval      int32    `json:"minApproval"`
		MinRemoval       int32    `json:"minRemoval"`
		Cosignatories    []string `json:"cosignatories"`
		MultisigAccounts []string `json:"multisigAccounts"`
	} `json:"multisig"`
}

type multisigAccountGraphInfoDTOEntry struct {
	Level     int32                    `json:"level"`
	Multisigs []multisigAccountInfoDTO `json:"multisigEntries"`
}

type multisigAccountGraphInfoDTOS []multisigAccountGraphInfoDTOEntry

type addresses []*Address

func (ref *addresses) MarshalJSON() (buf []byte, err error) {
	buf = []byte(`{"addresses":[`)
	for i, address := range *ref {
		b := []byte(`"` + address.Address + `"`)
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, b...)
	}

	buf = append(buf, ']', '}')
	return
}

func (ref *addresses) UnmarshalJSON(buf []byte) error {
	return nil
}

// generateEncodedAddress convert publicKey to address
func generateEncodedAddress(pKey string, version NetworkType) (string, error) {
	// step 1: sha3 hash of the public key
	pKeyD, err := hex.DecodeString(pKey)
	if err != nil {
		return "", err
	}
	sha3PublicKeyHash, err := crypto.HashesSha3_256(pKeyD)
	if err != nil {
		return "", err
	}
	// step 2: ripemd160 hash of (1)
	ripemd160StepOneHash, err := crypto.HashesRipemd160(sha3PublicKeyHash)
	if err != nil {
		return "", err
	}

	// step 3: add version byte in front of (2)
	versionPrefixedRipemd160Hash := append([]byte{uint8(version)}, ripemd160StepOneHash...)

	// step 4: get the checksum of (3)
	stepThreeChecksum, err := GenerateChecksum(versionPrefixedRipemd160Hash)
	if err != nil {
		return "", err
	}

	// step 5: concatenate (3) and (4)
	concatStepThreeAndStepSix := append(versionPrefixedRipemd160Hash, stepThreeChecksum...)

	// step 6: base32 encode (5)
	return base32.StdEncoding.EncodeToString(concatStepThreeAndStepSix), nil
}
