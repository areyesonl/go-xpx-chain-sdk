// Copyright 2018 ProximaX Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sdk

import (
	"context"
	"fmt"
	"github.com/proximax-storage/go-xpx-utils/net"
	"net/http"
)

type MosaicService service

// GetMosaic returns
// @get /mosaic/{mosaicId}
func (ref *MosaicService) GetMosaic(ctx context.Context, mosaicId *MosaicId) (*MosaicInfo, error) {
	if mosaicId == nil {
		return nil, ErrNilMosaicId
	}

	url := net.NewUrl(fmt.Sprintf(mosaicRoute, mosaicId.toHexString()))

	dto := &mosaicInfoDTO{}

	resp, err := ref.client.DoNewRequest(ctx, http.MethodGet, url.Encode(), nil, dto)
	if err != nil {
		return nil, err
	}

	if err = handleResponseStatusCode(resp, map[int]error{404: ErrResourceNotFound, 409: ErrArgumentNotValid}); err != nil {
		return nil, err
	}

	mscInfo, err := dto.toStruct(ref.client.config.NetworkType)
	if err != nil {
		return nil, err
	}

	return mscInfo, nil
}

// GetMosaics get list mosaics Info
// post @/mosaic/
func (ref *MosaicService) GetMosaics(ctx context.Context, mscIds []*MosaicId) ([]*MosaicInfo, error) {
	if len(mscIds) == 0 {
		return nil, ErrEmptyMosaicIds
	}

	dtos := mosaicInfoDTOs(make([]*mosaicInfoDTO, 0))

	resp, err := ref.client.DoNewRequest(ctx, http.MethodPost, mosaicsRoute, &mosaicIds{mscIds}, &dtos)
	if err != nil {
		return nil, err
	}

	if err = handleResponseStatusCode(resp, map[int]error{400: ErrInvalidRequest, 409: ErrArgumentNotValid}); err != nil {
		return nil, err
	}

	mscInfos, err := dtos.toStruct(ref.client.config.NetworkType)
	if err != nil {
		return nil, err
	}

	return mscInfos, nil
}
