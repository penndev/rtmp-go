package h265

import (
	"encoding/binary"
	"errors"
)

// ISO/IEC 14496-15 8.3.3.1.2 — HEVCDecoderConfigurationRecord
type HEVCDecoderConfigurationRecord struct {
	ConfigurationVersion byte

	// unsigned int(2) general_profile_space
	GeneralProfileSpace byte
	// unsigned int(1) general_tier_flag
	GeneralTierFlag byte
	// unsigned int(5) general_profile_idc
	GeneralProfileIDC byte
	// unsigned int(32) general_profile_compatibility_flags
	GeneralProfileCompatibilityFlags uint32
	// unsigned int(48) general_constraint_indicator_flags
	GeneralConstraintIndicatorFlags [6]byte
	// unsigned int(8) general_level_idc
	GeneralLevelIDC byte

	// bit(4) reserved = 1111b; unsigned int(12) min_spatial_segmentation_idc
	MinSpatialSegmentationIDC uint16
	// bit(6) reserved = 111111b; unsigned int(2) parallelismType
	ParallelismType byte
	// bit(6) reserved = 111111b; unsigned int(2) chroma_format_idc
	ChromaFormatIDC byte
	// bit(5) reserved = 11111b; unsigned int(3) bit_depth_luma_minus8
	BitDepthLumaMinus8 byte
	// bit(5) reserved = 11111b; unsigned int(3) bit_depth_chroma_minus8
	BitDepthChromaMinus8 byte

	// unsigned int(16) avgFrameRate
	AvgFrameRate uint16
	// unsigned int(2) constantFrameRate
	ConstantFrameRate byte
	// unsigned int(3) numTemporalLayers
	NumTemporalLayers byte
	// unsigned int(1) temporalIdNested
	TemporalIDNested byte
	// unsigned int(2) lengthSizeMinusOne
	LengthSizeMinusOne byte

	NALArrays []HEVCNALArray
}

// One iteration of the numOfArrays loop.
type HEVCNALArray struct {
	// unsigned int(1) array_completeness
	ArrayCompleteness bool
	// unsigned int(6) NAL_unit_type
	NALUnitType byte
	NALUnits    [][]byte
}

func FormatHEVCDecoderConfigurationRecord(b []byte) (HEVCDecoderConfigurationRecord, error) {
	var r HEVCDecoderConfigurationRecord
	if len(b) < 23 {
		return r, errors.New("h265: HEVCDecoderConfigurationRecord too short")
	}
	r.ConfigurationVersion = b[0]
	r.GeneralProfileSpace = b[1] >> 6
	r.GeneralTierFlag = (b[1] >> 5) & 0x01
	r.GeneralProfileIDC = b[1] & 0x1f
	r.GeneralProfileCompatibilityFlags = binary.BigEndian.Uint32(b[2:6])
	copy(r.GeneralConstraintIndicatorFlags[:], b[6:12])
	r.GeneralLevelIDC = b[12]
	r.MinSpatialSegmentationIDC = binary.BigEndian.Uint16(b[13:15]) & 0x0fff
	r.ParallelismType = b[15] & 0x03
	r.ChromaFormatIDC = b[16] & 0x03
	r.BitDepthLumaMinus8 = b[17] & 0x07
	r.BitDepthChromaMinus8 = b[18] & 0x07
	r.AvgFrameRate = binary.BigEndian.Uint16(b[19:21])
	r.ConstantFrameRate = b[21] >> 6
	r.NumTemporalLayers = (b[21] >> 3) & 0x07
	r.TemporalIDNested = (b[21] >> 2) & 0x01
	r.LengthSizeMinusOne = b[21] & 0x03
	numArrays := int(b[22])
	b = b[23:]

	r.NALArrays = make([]HEVCNALArray, 0, numArrays)
	for j := 0; j < numArrays; j++ {
		if len(b) < 3 {
			return r, errors.New("h265: truncated NAL array")
		}
		arr := HEVCNALArray{
			ArrayCompleteness: b[0]&0x80 != 0,
			NALUnitType:       b[0] & 0x3f,
		}
		numNalus := int(binary.BigEndian.Uint16(b[1:3]))
		b = b[3:]
		arr.NALUnits = make([][]byte, 0, numNalus)
		for i := 0; i < numNalus; i++ {
			if len(b) < 2 {
				return r, errors.New("h265: truncated NAL unit length")
			}
			n := int(binary.BigEndian.Uint16(b[:2]))
			b = b[2:]
			if len(b) < n {
				return r, errors.New("h265: truncated NAL unit")
			}
			arr.NALUnits = append(arr.NALUnits, append([]byte(nil), b[:n]...))
			b = b[n:]
		}
		r.NALArrays = append(r.NALArrays, arr)
	}
	return r, nil
}

func (r HEVCDecoderConfigurationRecord) Marshal() ([]byte, error) {
	if len(r.NALArrays) > 0xff {
		return nil, errors.New("h265: too many NAL arrays")
	}
	out := make([]byte, 23)
	out[0] = r.ConfigurationVersion
	out[1] = (r.GeneralProfileSpace&0x03)<<6 | (r.GeneralTierFlag&0x01)<<5 | (r.GeneralProfileIDC & 0x1f)
	binary.BigEndian.PutUint32(out[2:6], r.GeneralProfileCompatibilityFlags)
	copy(out[6:12], r.GeneralConstraintIndicatorFlags[:])
	out[12] = r.GeneralLevelIDC
	binary.BigEndian.PutUint16(out[13:15], 0xf000|(r.MinSpatialSegmentationIDC&0x0fff))
	out[15] = 0xfc | (r.ParallelismType & 0x03)
	out[16] = 0xfc | (r.ChromaFormatIDC & 0x03)
	out[17] = 0xf8 | (r.BitDepthLumaMinus8 & 0x07)
	out[18] = 0xf8 | (r.BitDepthChromaMinus8 & 0x07)
	binary.BigEndian.PutUint16(out[19:21], r.AvgFrameRate)
	out[21] = (r.ConstantFrameRate&0x03)<<6 | (r.NumTemporalLayers&0x07)<<3 | (r.TemporalIDNested&0x01)<<2 | (r.LengthSizeMinusOne & 0x03)
	out[22] = byte(len(r.NALArrays))

	for _, arr := range r.NALArrays {
		if len(arr.NALUnits) > 0xffff {
			return nil, errors.New("h265: too many NAL units")
		}
		hdr := arr.NALUnitType & 0x3f
		if arr.ArrayCompleteness {
			hdr |= 0x80
		}
		out = append(out, hdr)
		var n [2]byte
		binary.BigEndian.PutUint16(n[:], uint16(len(arr.NALUnits)))
		out = append(out, n[:]...)
		for _, nal := range arr.NALUnits {
			if len(nal) > 0xffff {
				return nil, errors.New("h265: NAL unit too long")
			}
			binary.BigEndian.PutUint16(n[:], uint16(len(nal)))
			out = append(out, n[:]...)
			out = append(out, nal...)
		}
	}
	return out, nil
}
