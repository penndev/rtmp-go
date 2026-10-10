package h265

import (
	"bytes"
	"testing"
)

func TestHEVCDecoderConfigurationRecordRoundTrip(t *testing.T) {
	in := HEVCDecoderConfigurationRecord{
		ConfigurationVersion:             1,
		GeneralProfileSpace:              0,
		GeneralTierFlag:                  0,
		GeneralProfileIDC:                1,
		GeneralProfileCompatibilityFlags: 0x60000000,
		GeneralConstraintIndicatorFlags:  [6]byte{0xb0, 0, 0, 0, 0, 0},
		GeneralLevelIDC:                  93,
		MinSpatialSegmentationIDC:        0,
		ParallelismType:                  0,
		ChromaFormatIDC:                  1,
		BitDepthLumaMinus8:               2,
		BitDepthChromaMinus8:             2,
		AvgFrameRate:                     0,
		ConstantFrameRate:                0,
		NumTemporalLayers:                1,
		TemporalIDNested:                 1,
		LengthSizeMinusOne:               3,
		NALArrays: []HEVCNALArray{{
			ArrayCompleteness: true,
			NALUnitType:       32,
			NALUnits:          [][]byte{{0x40, 0x01}},
		}},
	}
	raw, err := in.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	got, err := FormatHEVCDecoderConfigurationRecord(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.ConfigurationVersion != 1 || got.GeneralProfileIDC != 1 ||
		got.GeneralProfileCompatibilityFlags != 0x60000000 ||
		got.GeneralConstraintIndicatorFlags != in.GeneralConstraintIndicatorFlags ||
		got.GeneralLevelIDC != 93 || got.ChromaFormatIDC != 1 ||
		got.BitDepthLumaMinus8 != 2 || got.BitDepthChromaMinus8 != 2 ||
		got.NumTemporalLayers != 1 || got.TemporalIDNested != 1 ||
		got.LengthSizeMinusOne != 3 || len(got.NALArrays) != 1 {
		t.Fatalf("%+v", got)
	}
	arr := got.NALArrays[0]
	if !arr.ArrayCompleteness || arr.NALUnitType != 32 || len(arr.NALUnits) != 1 ||
		!bytes.Equal(arr.NALUnits[0], []byte{0x40, 0x01}) {
		t.Fatalf("array=%+v", arr)
	}
}
