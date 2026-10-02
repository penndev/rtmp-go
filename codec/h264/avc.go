package h264

import (
	"encoding/binary"
	"errors"
)

// ISO/IEC 14496-15 — AVCDecoderConfigurationRecord
type AVCDecoderConfigurationRecord struct {
	ConfigurationVersion byte
	AVCProfileIndication byte
	ProfileCompatibility byte
	AVCLevelIndication   byte
	// bit(6) reserved = '111111'b; unsigned int(2) lengthSizeMinusOne
	LengthSizeMinusOne byte
	// bit(3) reserved = '111'b; unsigned int(5) numOfSequenceParameterSets
	SequenceParameterSetNALUnit [][]byte
	PictureParameterSetNALUnit  [][]byte

	// if AVCProfileIndication == 100 || 110 || 122 || 144
	ChromaFormat                   byte // bit(6) reserved; unsigned int(2)
	BitDepthLumaMinus8             byte // bit(5) reserved; unsigned int(3)
	BitDepthChromaMinus8           byte // bit(5) reserved; unsigned int(3)
	SequenceParameterSetExtNALUnit [][]byte
}

func FormatAVCDecoderConfigurationRecord(b []byte) (AVCDecoderConfigurationRecord, error) {
	var r AVCDecoderConfigurationRecord
	if len(b) < 6 {
		return r, errors.New("h264: AVCDecoderConfigurationRecord too short")
	}
	r.ConfigurationVersion = b[0]
	r.AVCProfileIndication = b[1]
	r.ProfileCompatibility = b[2]
	r.AVCLevelIndication = b[3]
	r.LengthSizeMinusOne = b[4] & 0x03
	numSPS := int(b[5] & 0x1f)
	b = b[6:]

	var err error
	r.SequenceParameterSetNALUnit, b, err = readLengthPrefixedNALUs(b, numSPS)
	if err != nil {
		return r, err
	}
	if len(b) < 1 {
		return r, errors.New("h264: missing numOfPictureParameterSets")
	}
	numPPS := int(b[0])
	b = b[1:]
	r.PictureParameterSetNALUnit, b, err = readLengthPrefixedNALUs(b, numPPS)
	if err != nil {
		return r, err
	}

	switch r.AVCProfileIndication {
	case 100, 110, 122, 144:
		if len(b) < 4 {
			return r, errors.New("h264: AVCDecoderConfigurationRecord ext too short")
		}
		r.ChromaFormat = b[0] & 0x03
		r.BitDepthLumaMinus8 = b[1] & 0x07
		r.BitDepthChromaMinus8 = b[2] & 0x07
		numSPSExt := int(b[3])
		b = b[4:]
		r.SequenceParameterSetExtNALUnit, _, err = readLengthPrefixedNALUs(b, numSPSExt)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

func (r AVCDecoderConfigurationRecord) Marshal() ([]byte, error) {
	if len(r.SequenceParameterSetNALUnit) > 0x1f {
		return nil, errors.New("h264: too many SPS")
	}
	if len(r.PictureParameterSetNALUnit) > 0xff {
		return nil, errors.New("h264: too many PPS")
	}
	out := make([]byte, 0, 6)
	out = append(out,
		r.ConfigurationVersion,
		r.AVCProfileIndication,
		r.ProfileCompatibility,
		r.AVCLevelIndication,
		0xfc|r.LengthSizeMinusOne&0x03,
		0xe0|byte(len(r.SequenceParameterSetNALUnit)),
	)
	var err error
	out, err = appendLengthPrefixedNALUs(out, r.SequenceParameterSetNALUnit)
	if err != nil {
		return nil, err
	}
	out = append(out, byte(len(r.PictureParameterSetNALUnit)))
	out, err = appendLengthPrefixedNALUs(out, r.PictureParameterSetNALUnit)
	if err != nil {
		return nil, err
	}

	switch r.AVCProfileIndication {
	case 100, 110, 122, 144:
		if len(r.SequenceParameterSetExtNALUnit) > 0xff {
			return nil, errors.New("h264: too many SPS Ext")
		}
		out = append(out,
			0xfc|r.ChromaFormat&0x03,
			0xf8|r.BitDepthLumaMinus8&0x07,
			0xf8|r.BitDepthChromaMinus8&0x07,
			byte(len(r.SequenceParameterSetExtNALUnit)),
		)
		out, err = appendLengthPrefixedNALUs(out, r.SequenceParameterSetExtNALUnit)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func readLengthPrefixedNALUs(b []byte, n int) ([][]byte, []byte, error) {
	out := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		if len(b) < 2 {
			return nil, b, errors.New("h264: truncated NAL unit length")
		}
		l := int(binary.BigEndian.Uint16(b[:2]))
		b = b[2:]
		if len(b) < l {
			return nil, b, errors.New("h264: truncated NAL unit")
		}
		out = append(out, append([]byte(nil), b[:l]...))
		b = b[l:]
	}
	return out, b, nil
}

func appendLengthPrefixedNALUs(out []byte, nals [][]byte) ([]byte, error) {
	for _, nal := range nals {
		if len(nal) > 0xffff {
			return nil, errors.New("h264: NAL unit too long")
		}
		var lenBuf [2]byte
		binary.BigEndian.PutUint16(lenBuf[:], uint16(len(nal)))
		out = append(out, lenBuf[:]...)
		out = append(out, nal...)
	}
	return out, nil
}
