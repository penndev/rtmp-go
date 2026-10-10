package flv

// E.4.2.1 AudioTagHeader — SoundFormat UB[4]
const (
	SOUND_FORMAT_LINEAR_PCM_PE    byte = 0  // Linear PCM, platform endian
	SOUND_FORMAT_ADPCM            byte = 1  // ADPCM
	SOUND_FORMAT_MP3              byte = 2  // MP3
	SOUND_FORMAT_LINEAR_PCM_LE    byte = 3  // Linear PCM, little endian
	SOUND_FORMAT_NELLYMOSER_16KHZ byte = 4  // Nellymoser 16-kHz mono
	SOUND_FORMAT_NELLYMOSER_8KHZ  byte = 5  // Nellymoser 8-kHz mono
	SOUND_FORMAT_NELLYMOSER       byte = 6  // Nellymoser
	SOUND_FORMAT_G711_A_LAW       byte = 7  // G.711 A-law logarithmic PCM
	SOUND_FORMAT_G711_MU_LAW      byte = 8  // G.711 mu-law logarithmic PCM
	SOUND_FORMAT_RESERVED         byte = 9  // reserved
	SOUND_FORMAT_AAC              byte = 10 // AAC
	SOUND_FORMAT_SPEEX            byte = 11 // Speex
	SOUND_FORMAT_MP3_8KHZ         byte = 14 // MP3 8-Khz
	SOUND_FORMAT_DEVICE_SPECIFIC  byte = 15 // Device-specific sound
)

// E.4.2.1 AudioTagHeader — SoundRate UB[2]
const (
	SOUND_RATE_5_5KHZ byte = 0 // 5.5 kHz
	SOUND_RATE_11KHZ  byte = 1 // 11 kHz
	SOUND_RATE_22KHZ  byte = 2 // 22 kHz
	SOUND_RATE_44KHZ  byte = 3 // 44 kHz
)

// E.4.2.1 AudioTagHeader — SoundSize UB[1]
const (
	SOUND_SIZE_8BIT  byte = 0 // snd8Bit
	SOUND_SIZE_16BIT byte = 1 // snd16Bit
)

// E.4.2.1 AudioTagHeader — SoundType UB[1]
const (
	SOUND_TYPE_MONO   byte = 0 // sndMono
	SOUND_TYPE_STEREO byte = 1 // sndStereo
)
