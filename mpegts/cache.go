package mpegts

type ExtInf struct {
	Inf  uint32
	File string
}

var cache = make(map[string][]ExtInf)

func HlsLive(topic string) ([]ExtInf, int, bool) {
	if v, ok := cache[topic]; ok {
		l := len(v)
		if l < 3 {
			return v, 0, ok
		} else {
			return v[l-3:], l, ok
		}

	}
	return nil, 0, false
}
