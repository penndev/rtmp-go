package flv

import (
	"errors"

	"github.com/penndev/rtmp/codec/h264"
)

func FormatTag(tagType TagType, timestamp uint32, data []byte, avcDecoderConfigurationRecord *h264.AVCDecoderConfigurationRecord) (*Tag, error) {
	t := &Tag{
		TagType:   tagType,
		Timestamp: timestamp,
		DataSize:  uint32(len(data)),
		Data:      data,
	}
	switch tagType {
	case TAG_TYPE_AUDIO:
		if len(data) < 1 {
			return nil, errors.New("flv: audio tag too short")
		}
		t.AudioTagHeader = AudioTagHeader{
			SoundFormat: data[0] >> 4,
			SoundRate:   (data[0] >> 2) & 0x03,
			SoundSize:   (data[0] >> 1) & 0x01,
			SoundType:   data[0] & 0x01,
		}
		if t.SoundFormat == SOUND_FORMAT_AAC {
			if len(data) < 2 {
				return nil, errors.New("flv: AAC audio tag too short")
			}
			t.AACPacketType = AACPacketType(data[1])
		}
	case TAG_TYPE_VIDEO:
		if len(data) < 1 {
			return nil, errors.New("flv: video tag too short")
		}
		t.VideoTagHeader = VideoTagHeader{
			IsExVideoHeader: data[0]&0x80 != 0,
			FrameType:       FrameType((data[0] >> 4) & 0x07),
		}
		if t.IsExVideoHeader {
			if err := TagEnhancedVideo(t, data); err != nil {
				return nil, err
			}
		} else {
			t.CodecID = CodecID(data[0] & 0x0f)
			if t.CodecID == CODEC_ID_AVC {
				if len(data) < 5 {
					return nil, errors.New("flv: AVC video tag too short")
				}
				t.AVCPacketType = AVCPacketType(data[1])
				cts := uint32(data[2])<<16 | uint32(data[3])<<8 | uint32(data[4])
				if cts&0x800000 != 0 {
					cts |= 0xff000000 // sign-extend SI24
				}
				t.CompositionTime = int32(cts)
				switch t.AVCPacketType {
				case AVC_PACKET_TYPE_SEQUENCE_HEADER:
					var err error
					t.AVCDecoderConfigurationRecord, err = h264.FormatAVCDecoderConfigurationRecord(data[5:])
					if err != nil {
						return nil, err
					}
				case AVC_PACKET_TYPE_NALU:
					if avcDecoderConfigurationRecord == nil {
						break
					}
					size := int(avcDecoderConfigurationRecord.LengthSizeMinusOne) + 1
					nalus, err := readAVCNalus(data[5:], size)
					if err != nil {
						return nil, err
					}
					t.Nalu = nalus
				}
			}
		}
	case TAG_TYPE_SCRIPT_DATA:
	default:
		return nil, errors.New("flv: unknown tag type")
	}
	return t, nil
}

func readAVCNalus(payload []byte, size int) ([]h264.Nalu, error) {
	var nalus []h264.Nalu
	for len(payload) >= size {
		l := 0
		for i := 0; i < size; i++ {
			l = l<<8 | int(payload[i])
		}
		payload = payload[size:]
		if len(payload) < l {
			return nil, errors.New("flv: truncated NAL unit")
		}
		n, err := h264.FormatNalu(payload[:l])
		if err != nil {
			return nil, err
		}
		nalus = append(nalus, n)
		payload = payload[l:]
	}
	return nalus, nil
}
