package pubsub

import (
	"log"
	"sync"

	"github.com/penndev/rtmp/codec/flv"
)

type FlvMeta struct {
	FlvScriptData                   *Message
	CodecID                         uint32
	VideoDecoderConfigurationRecord *Message
	SoundFormat                     uint32
	AudioSpecificConfig             *Message
}

var (
	flvMetaMu  sync.RWMutex
	FlvMetaMap = make(map[string]*FlvMeta)
)

var rtmpBroker *Broker = New()

func PubTopic(name string) *Topic {
	top := rtmpBroker.Topic(name)
	flvMetaMu.Lock()
	meta := &FlvMeta{}
	FlvMetaMap[name] = meta
	flvMetaMu.Unlock()

	initVideo := false
	initAudio := false
	initScript := false
	top.BeforePublish = func(msg *Message) {
		if initScript && initVideo && initAudio {
			top.BeforePublish = nil
			return
		}

		raw, ok := msg.Data.(flv.FlvTag)
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
			flvMetaMu.Lock()
			meta.FlvScriptData = msg
			flvMetaMu.Unlock()
			initScript = true
		case flv.TAG_TYPE_AUDIO:
			if initAudio {
				return
			}
			flvMetaMu.Lock()
			meta.SoundFormat = uint32(tag.SoundFormat)
			if tag.SoundFormat == flv.SOUND_FORMAT_AAC && tag.AACPacketType == flv.AAC_PACKET_TYPE_SEQUENCE_HEADER {
				meta.AudioSpecificConfig = msg
			}
			flvMetaMu.Unlock()
			if tag.SoundFormat != flv.SOUND_FORMAT_AAC || tag.AACPacketType == flv.AAC_PACKET_TYPE_SEQUENCE_HEADER {
				initAudio = true
			}
		case flv.TAG_TYPE_VIDEO:
			if initVideo {
				return
			}
			flvMetaMu.Lock()
			meta.CodecID = uint32(tag.CodecID)
			if tag.CodecID == flv.CODEC_ID_AVC && tag.AVCPacketType == flv.AVC_PACKET_TYPE_SEQUENCE_HEADER {
				meta.VideoDecoderConfigurationRecord = msg
			}
			flvMetaMu.Unlock()
			if tag.CodecID != flv.CODEC_ID_AVC || tag.AVCPacketType == flv.AVC_PACKET_TYPE_SEQUENCE_HEADER {
				initVideo = true
			}
		}
	}
	return top
}

func SubTopic(name string) *Subscription {
	top := rtmpBroker.Topic(name)
	sub := &Subscription{
		topic: top,
		ch:    make(chan *Message, 64),
	}
	flvMetaMu.RLock()
	meta := FlvMetaMap[name]
	var script, video, audio *Message
	if meta != nil {
		script = meta.FlvScriptData
		if meta.CodecID == uint32(flv.CODEC_ID_AVC) {
			video = meta.VideoDecoderConfigurationRecord
		}
		if meta.SoundFormat == uint32(flv.SOUND_FORMAT_AAC) {
			audio = meta.AudioSpecificConfig
		}
	}
	flvMetaMu.RUnlock()
	if script != nil {
		sub.Write(script)
	}
	if video != nil {
		sub.Write(video)
	}
	if audio != nil {
		sub.Write(audio)
	}
	sub.Filter = func(msg *Message) bool {
		tag, ok := msg.Data.(flv.FlvTag)
		if !ok || tag.Type() != flv.TAG_TYPE_VIDEO {
			return false
		}
		data := tag.Data()
		if len(data) < 1 || data[0]>>4 != flv.FRAME_TYPE_KEY {
			return false
		}
		sub.Filter = nil
		return true
	}
	top.Attach(sub)
	return sub
}
