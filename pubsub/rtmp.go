package pubsub

import (
	"errors"
	"log"
	"sync"

	"github.com/penndev/rtmp/codec/flv"
	"github.com/penndev/rtmp/rtmp/stream"
)

type flvMeta struct {
	FlvScriptData                   any
	VideoDecoderConfigurationRecord any
	AudioSpecificConfig             any
}

type Hub struct {
	mu     sync.RWMutex
	meta   map[string]*flvMeta
	broker *Broker
}

func NewRtmp() *Hub {
	return &Hub{
		meta:   make(map[string]*flvMeta),
		broker: NewBroker(),
	}
}

func (h *Hub) Publish(name string) (stream.Publisher, error) {
	h.mu.Lock()
	if _, ok := h.meta[name]; ok {
		h.mu.Unlock()
		return nil, errors.New("already publishing")
	}
	meta := &flvMeta{}
	h.meta[name] = meta
	h.mu.Unlock()

	top := h.broker.Topic(name)

	top.OnClose = func() {
		h.mu.Lock()
		if h.meta[name] == meta {
			delete(h.meta, name)
		}
		h.mu.Unlock()
	}

	initVideo := false
	initAudio := false
	initScript := false
	// action before publish
	// handle h264 sps pps
	// handle aac sequence header
	top.BeforePublish = func(msg any) {
		if initScript && initVideo && initAudio {
			top.BeforePublish = nil // ！- action after publish
			return
		}

		raw, ok := msg.(flv.TagReader)
		if !ok {
			return
		}
		tag, err := flv.FormatTag(raw.Type(), raw.Timestamp(), raw.Data(), nil)
		if err != nil {
			log.Println(err)
			return
		}
		switch tag.TagType {
		case flv.TAG_TYPE_SCRIPT_DATA:
			if initScript {
				return
			}
			h.mu.Lock()
			meta.FlvScriptData = msg
			h.mu.Unlock()
			initScript = true
		case flv.TAG_TYPE_AUDIO:
			if initAudio {
				return
			}
			h.mu.Lock()
			if tag.SoundFormat == flv.SOUND_FORMAT_AAC && tag.AACPacketType == flv.AAC_PACKET_TYPE_SEQUENCE_HEADER {
				meta.AudioSpecificConfig = msg
			}
			h.mu.Unlock()
			if tag.SoundFormat != flv.SOUND_FORMAT_AAC || tag.AACPacketType == flv.AAC_PACKET_TYPE_SEQUENCE_HEADER {
				initAudio = true
			}
		case flv.TAG_TYPE_VIDEO:
			if initVideo {
				return
			}
			avcSeq := tag.CodecID == flv.CODEC_ID_AVC && tag.AVCPacketType == flv.AVC_PACKET_TYPE_SEQUENCE_HEADER
			hevcSeq := tag.FourCC == flv.FOURCC_HEVC && tag.VideoPacketType == flv.VIDEO_PACKET_TYPE_SEQUENCE_START
			h.mu.Lock()
			if avcSeq || hevcSeq {
				meta.VideoDecoderConfigurationRecord = msg
			}
			h.mu.Unlock()
			if tag.CodecID == flv.CODEC_ID_AVC {
				initVideo = avcSeq
			} else if tag.FourCC == flv.FOURCC_HEVC {
				initVideo = hevcSeq
			} else {
				initVideo = true
			}
		}
	}
	return top, nil
}

func (h *Hub) Play(name string) (stream.Subscriber, error) {
	h.mu.RLock()
	meta := h.meta[name]
	var head []any
	if meta != nil {
		head = []any{meta.FlvScriptData, meta.VideoDecoderConfigurationRecord, meta.AudioSpecificConfig}
	}
	h.mu.RUnlock()
	if meta == nil {
		return nil, errors.New("no stream")
	}

	top := h.broker.Topic(name)
	sub := NewSubscription(top)
	for _, m := range head {
		if m != nil {
			sub.Write(m)
		}
	}
	sub.Filter = func(msg any) bool {
		raw, ok := msg.(flv.TagReader)
		if !ok {
			return false
		}
		tag, err := flv.FormatTag(raw.Type(), raw.Timestamp(), raw.Data(), nil)
		if err != nil {
			return false
		}
		if tag.TagType != flv.TAG_TYPE_VIDEO {
			return false
		}

		if tag.FrameType != flv.FRAME_TYPE_KEY {
			return false
		}

		sub.Filter = nil
		return true
	}
	top.Attach(sub)
	return sub, nil
}
