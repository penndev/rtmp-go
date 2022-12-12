package mpegts

type ExtInf struct {
	Inf  uint32
	File string
}

var cache map[string][]ExtInf

func init() {
	cache = make(map[string][]ExtInf)
}

func LiveList(topic string) ([]ExtInf, bool) {
	if v, ok := cache[topic]; ok {
		if len(v) < 3 {
			return v, ok
		} else {
			return v[len(v)-3:], ok
		}

	}
	return nil, false
}
