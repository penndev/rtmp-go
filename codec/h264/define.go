package h264

// Table 7-1 – NAL unit type codes, syntax element categories, and NAL unit type classes
// T-REC-H.264
type NalUnitType byte

const (
	NAL_UNIT_TYPE_SLICE           NalUnitType = 1  // Coded slice of a non-IDR picture
	NAL_UNIT_TYPE_A               NalUnitType = 2  // Coded slice data partition A
	NAL_UNIT_TYPE_B               NalUnitType = 3  // Coded slice data partition B
	NAL_UNIT_TYPE_C               NalUnitType = 4  // Coded slice data partition C
	NAL_UNIT_TYPE_IDR             NalUnitType = 5  // Coded slice of an IDR picture
	NAL_UNIT_TYPE_SEI             NalUnitType = 6  // Supplemental enhancement information (SEI)
	NAL_UNIT_TYPE_SPS             NalUnitType = 7  // Sequence parameter set
	NAL_UNIT_TYPE_PPS             NalUnitType = 8  // Picture parameter set
	NAL_UNIT_TYPE_AUD             NalUnitType = 9  // Access unit delimiter
	NAL_UNIT_TYPE_END_OF_SEQUENCE NalUnitType = 10 // End of sequence
	NAL_UNIT_TYPE_END_OF_STREAM   NalUnitType = 11 // End of stream
	NAL_UNIT_TYPE_FILLER          NalUnitType = 12 // Filler data
	NAL_UNIT_TYPE_SPS_EXT         NalUnitType = 13 // Sequence parameter set extension
	NAL_UNIT_TYPE_PREFIX          NalUnitType = 14 // Prefix NAL unit
	NAL_UNIT_TYPE_SUBSET_SPS      NalUnitType = 15 // Subset sequence parameter set
	NAL_UNIT_TYPE_DPS             NalUnitType = 16 // Depth parameter set
	NAL_UNIT_TYPE_AUX_SLICE       NalUnitType = 19 // Coded slice of an auxiliary coded picture without partitioning
	NAL_UNIT_TYPE_EXT_SLICE       NalUnitType = 20 // Coded slice extension
	NAL_UNIT_TYPE_DEPTH_EXT_SLICE NalUnitType = 21 // Coded slice extension for a depth view component or a 3D-AVC texture view component
)
