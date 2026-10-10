package h264

import "errors"

// 7.3.1 NAL unit syntax — nal_unit header
type Nalu struct {
	// forbidden_zero_bit f(1): shall be equal to 0
	ForbiddenZeroBit byte
	// nal_ref_idc u(2): not equal to 0 specifies that the content of the NAL unit
	// contains a sequence parameter set, a picture parameter set, a slice of a
	// reference picture, or a slice data partition of a reference picture
	NalRefIdc byte
	// nal_unit_type u(5): Table 7-1
	NalUnitType NalUnitType
}

func FormatNalu(b []byte) (Nalu, error) {
	if len(b) < 1 {
		return Nalu{}, errors.New("h264: nalu too short")
	}
	n := Nalu{
		ForbiddenZeroBit: b[0] >> 7,
		NalRefIdc:        (b[0] >> 5) & 0x03,
		NalUnitType:      NalUnitType(b[0] & 0x1f),
	}
	if n.ForbiddenZeroBit != 0 {
		return Nalu{}, errors.New("h264: forbidden_zero_bit must be 0")
	}
	return n, nil
}
