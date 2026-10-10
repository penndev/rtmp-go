package adapter

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/penndev/rtmp/codec/flv"
	"github.com/penndev/rtmp/codec/h264"
	"github.com/penndev/rtmp/codec/h265"
	"github.com/penndev/rtmp/codec/mpegts"
	"github.com/penndev/rtmp/rtmp/stream"
)

func AdapterTs(name string, sub stream.Subscriber) {
	defer sub.Close()
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		log.Println("ts: bad name", name)
		return
	}

	buf := &bytes.Buffer{}
	ts := mpegts.NewTS(buf)
	var avc *h264.AVCDecoderConfigurationRecord
	var hevc *h265.HEVCDecoderConfigurationRecord
	var asc []byte

	var seq int
	var segStart, last uint32
	var has bool
	closeSeg := func(end uint32) {
		if buf.Len() == 0 {
			return
		}
		dur := 0.0
		if end >= segStart {
			dur = float64(end-segStart) / 1000
		}
		dir := fmt.Sprintf("runtime/%s", name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Println(err)
			return
		}
		file := fmt.Sprintf("%s/%d.ts", dir, seq)
		if err := os.WriteFile(file, buf.Bytes(), 0644); err != nil {
			log.Println(err)
			return
		}
		buf.Reset()
		if dur <= 0 {
			dur = 0.001
		}
		hlsPlayMu.Lock()
		live := hlsPlay[name]
		if live == nil {
			live = &hlsLive{}
			hlsPlay[name] = live
		}
		live.push(tsSegment{path: file, duration: dur})
		hlsPlayMu.Unlock()
		seq++
	}
	defer func() {
		closeSeg(last)
		hlsPlayMu.Lock()
		if live := hlsPlay[name]; live != nil {
			live.endList = true
		}
		hlsPlayMu.Unlock()
	}()

	for m := range sub.Chan() {
		msg, ok := m.(flv.TagReader)
		if !ok || msg == nil {
			continue
		}
		tag, err := flv.FormatTag(msg.Type(), msg.Timestamp(), msg.Data(), nil)
		if err != nil {
			log.Println(err)
			continue
		}
		switch tag.TagType {
		case flv.TAG_TYPE_AUDIO:
			if tag.SoundFormat != flv.SOUND_FORMAT_AAC || len(msg.Data()) < 2 {
				continue
			}
			if tag.AACPacketType == flv.AAC_PACKET_TYPE_SEQUENCE_HEADER {
				asc = append([]byte(nil), msg.Data()[2:]...)
				ts.SetAudioStreamType(mpegts.StreamTypeAAC)
				continue
			}
			if asc == nil {
				continue
			}
			frame := adts(asc, msg.Data()[2:])
			if err := ts.WriteAudio(tag.Timestamp, frame); err != nil {
				log.Println("ts audio:", err)
				return
			}
			last = tag.Timestamp
		case flv.TAG_TYPE_VIDEO:
			if tag.FrameType == flv.FRAME_TYPE_VIDEO_INFO_COMMAND {
				continue
			}
			sequence := tag.CodecID == flv.CODEC_ID_AVC && tag.AVCPacketType == flv.AVC_PACKET_TYPE_SEQUENCE_HEADER
			if tag.IsExVideoHeader {
				sequence = tag.VideoPacketType == flv.VIDEO_PACKET_TYPE_SEQUENCE_START
			}
			if sequence {
				if tag.FourCC == flv.FOURCC_HEVC {
					rec := tag.HEVCDecoderConfigurationRecord
					hevc = &rec
					avc = nil
					ts.SetVideoStreamType(mpegts.StreamTypeH265)
				} else {
					rec := tag.AVCDecoderConfigurationRecord
					avc = &rec
					hevc = nil
					ts.SetVideoStreamType(mpegts.StreamTypeH264)
				}
				continue
			}
			coded := tag.CodecID == flv.CODEC_ID_AVC && tag.AVCPacketType == flv.AVC_PACKET_TYPE_NALU
			if tag.IsExVideoHeader {
				coded = tag.VideoPacketType == flv.VIDEO_PACKET_TYPE_CODED_FRAMES || tag.VideoPacketType == flv.VIDEO_PACKET_TYPE_CODED_FRAMES_X
			}
			if !coded || (avc == nil && hevc == nil) {
				continue
			}
			var body []byte
			if !tag.IsExVideoHeader {
				if len(msg.Data()) >= 5 {
					body = msg.Data()[5:]
				}
			} else if len(msg.Data()) >= 5 && msg.Data()[0]&0x0f != byte(flv.VIDEO_PACKET_TYPE_MODEX) && msg.Data()[0]&0x0f != byte(flv.VIDEO_PACKET_TYPE_MULTITRACK) {
				body = msg.Data()[5:]
				if tag.VideoPacketType == flv.VIDEO_PACKET_TYPE_CODED_FRAMES {
					if len(body) < 3 {
						body = nil
					} else {
						body = body[3:]
					}
				}
			}
			if body == nil {
				continue
			}
			size := 4
			if avc != nil {
				size = int(avc.LengthSizeMinusOne) + 1
			} else {
				size = int(hevc.LengthSizeMinusOne) + 1
			}
			var annex []byte
			key := tag.FrameType == flv.FRAME_TYPE_KEY
			if avc != nil {
				annex = append(annex, 0, 0, 0, 1, 0x09, 0xf0)
				if key {
					for _, nal := range avc.SequenceParameterSetNALUnit {
						annex = append(annex, 0, 0, 0, 1)
						annex = append(annex, nal...)
					}
					for _, nal := range avc.PictureParameterSetNALUnit {
						annex = append(annex, 0, 0, 0, 1)
						annex = append(annex, nal...)
					}
				}
			} else {
				annex = append(annex, 0, 0, 0, 1, 0x46, 0x01, 0x50)
				if key && hevc != nil {
					for _, arr := range hevc.NALArrays {
						for _, nal := range arr.NALUnits {
							annex = append(annex, 0, 0, 0, 1)
							annex = append(annex, nal...)
						}
					}
				}
			}
			if size < 1 || size > 4 {
				log.Println("ts video: bad nalu length size", size)
				continue
			}
			payload := body
			bad := false
			for len(payload) > 0 {
				if len(payload) < size {
					bad = true
					break
				}
				n := 0
				for i := 0; i < size; i++ {
					n = n<<8 | int(payload[i])
				}
				payload = payload[size:]
				if n < 0 || len(payload) < n {
					bad = true
					break
				}
				annex = append(annex, 0, 0, 0, 1)
				annex = append(annex, payload[:n]...)
				payload = payload[n:]
			}
			if bad {
				log.Println("ts video: truncated nalu")
				continue
			}
			if key && has && tag.Timestamp >= segStart && tag.Timestamp-segStart >= liveItemDuration {
				closeSeg(tag.Timestamp)
				segStart = tag.Timestamp
			}
			if err := ts.WriteVideo(tag.Timestamp, tag.CompositionTime, annex, key); err != nil {
				log.Println("ts video:", err)
				return
			}
			if !has {
				has = true
				segStart = tag.Timestamp
			}
			last = tag.Timestamp
		}
	}
}

func adts(asc, frame []byte) []byte {
	if len(asc) < 2 {
		return nil
	}
	objectType := int(asc[0] >> 3)
	if objectType < 1 {
		return nil
	}
	profile := objectType - 1
	freq := int((asc[0]&0x07)<<1 | asc[1]>>7)
	ch := int((asc[1] >> 3) & 0x0f)
	frameLen := len(frame) + 7
	out := make([]byte, frameLen)
	out[0] = 0xff
	out[1] = 0xf1
	out[2] = byte((profile&0x03)<<6) | byte((freq&0x0f)<<2) | byte((ch>>2)&0x01)
	out[3] = byte((ch&0x03)<<6) | byte((frameLen>>11)&0x03)
	out[4] = byte(frameLen >> 3)
	out[5] = byte((frameLen&0x07)<<5) | 0x1f
	out[6] = 0xfc
	copy(out[7:], frame)
	return out
}
