// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package decodeutils

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

// DecodePlmnIDDecStrToBytes decodes the PLMNID string type to bytes type
func DecodePlmnIDDecStrToBytes(plmnidDecStr string) ([]byte, error) {
	var plmnBytes [3]uint8
	n, err := strconv.ParseUint(plmnidDecStr, 10, 32)
	if err != nil {
		return nil, err
	}
	plmnid := uint32(n)

	plmnBytes[0] = uint8(plmnid & 0xFF)
	plmnBytes[1] = uint8((plmnid >> 8) & 0xFF)
	plmnBytes[2] = uint8((plmnid >> 16) & 0xFF)

	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, plmnBytes); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeCIDDecStrToUint64 decodes the CID string type to bytes type
func DecodeCIDDecStrToUint64(cidDecStr string) (uint64, error) {
	cid, err := strconv.ParseUint(cidDecStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return cid, err
}
