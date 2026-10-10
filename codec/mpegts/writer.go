package mpegts

import "io"

// TS writes a single program: PAT, PMT, and one video plus one audio elementary stream.
type TS struct {
	w io.Writer

	videoType byte
	audioType byte
	hasAudio  bool

	patCC   byte
	pmtCC   byte
	videoCC byte
	audioCC byte

	psiVideo   byte
	psiAudio   bool
	psiWritten bool
}

func NewTS(w io.Writer) *TS {
	return &TS{w: w, videoType: StreamTypeH264, audioType: StreamTypeAAC}
}

func (t *TS) SetVideoStreamType(streamType byte) {
	t.videoType = streamType
}

func (t *TS) SetAudioStreamType(streamType byte) {
	t.hasAudio = true
	t.audioType = streamType
}

func (t *TS) WriteVideo(dtsMs uint32, ctsMs int32, annexB []byte, randomAccess bool) error {
	if len(annexB) == 0 {
		return nil
	}
	if err := t.writePSI(randomAccess); err != nil {
		return err
	}
	dts := Clock90(dtsMs, 0)
	pts := Clock90(dtsMs, ctsMs)
	pes := PES{StreamID: streamIDVideo, PTS: pts, DTS: dts, HasDTS: true, Data: annexB}.Marshal()
	return t.packets(PIDVideo, &t.videoCC, pes, true, randomAccess, &dts)
}

func (t *TS) WriteAudio(ptsMs uint32, adts []byte) error {
	if len(adts) == 0 {
		return nil
	}
	if err := t.writePSI(false); err != nil {
		return err
	}
	pts := Clock90(ptsMs, 0)
	pes := PES{StreamID: streamIDAudio, PTS: pts, Data: adts}.Marshal()
	return t.packets(PIDAudio, &t.audioCC, pes, true, false, nil)
}

func (t *TS) writePSI(force bool) error {
	if t.psiWritten && !force && t.psiVideo == t.videoType && t.psiAudio == t.hasAudio {
		return nil
	}
	pat := PAT{TransportStreamID: 1, ProgramNumber: 1, ProgramMapPID: PIDPMT}.Marshal()
	if err := t.packets(PIDPAT, &t.patCC, append([]byte{0x00}, pat...), true, false, nil); err != nil {
		return err
	}
	pmt := PMT{ProgramNumber: 1, PCRPID: PIDVideo, Streams: []ElementaryStream{
		{StreamType: t.videoType, ElementaryPID: PIDVideo},
	}}
	if t.hasAudio {
		pmt.Streams = append(pmt.Streams, ElementaryStream{StreamType: t.audioType, ElementaryPID: PIDAudio})
	}
	if err := t.packets(PIDPMT, &t.pmtCC, append([]byte{0x00}, pmt.Marshal()...), true, false, nil); err != nil {
		return err
	}
	t.psiWritten = true
	t.psiVideo = t.videoType
	t.psiAudio = t.hasAudio
	return nil
}

func (t *TS) packets(pid uint16, cc *byte, payload []byte, pusi, rai bool, pcr *uint64) error {
	for len(payload) > 0 {
		afBytes := 0
		if pusi && (rai || pcr != nil) {
			afBytes = 2
			if pcr != nil {
				afBytes += 6
			}
		}
		if 184-afBytes > len(payload) {
			afBytes = 184 - len(payload)
		}
		var af *AdaptationField
		if afBytes > 0 {
			af = &AdaptationField{Length: byte(afBytes - 1), RandomAccessIndicator: rai && pusi}
			if pcr != nil && pusi {
				af.PCRFlag = true
				af.PCR = *pcr
			}
		}
		n := 184 - afBytes
		pkt := Packet{
			SyncByte:                  SyncByte,
			PayloadUnitStartIndicator: pusi,
			PID:                       pid,
			AdaptationField:           af,
			ContinuityCounter:         *cc,
			Payload:                   payload[:n],
		}
		b, err := pkt.Marshal()
		if err != nil {
			return err
		}
		if _, err := t.w.Write(b); err != nil {
			return err
		}
		payload = payload[n:]
		pusi = false
		*cc = (*cc + 1) & 0x0f
	}
	return nil
}
