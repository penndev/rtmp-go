package rtmp

// // chan 传输内容
// type Pack struct {
// 	Type    byte
// 	Time    uint32
// 	Content []byte
// }

// type WorkPool struct {
// 	PlayList  map[string](map[chan Pack]bool)
// 	MateList  map[string]Pack
// 	VideoList map[string]Pack
// 	AudioList map[string]Pack
// }

// func (wp *WorkPool) Close(room string, push chan Pack) {
// 	for py := range wp.PlayList[room] {
// 		close(py)
// 	}
// 	delete(wp.PlayList, room)
// 	delete(wp.MateList, room)
// 	delete(wp.VideoList, room)
// 	delete(wp.AudioList, room)
// }

// func (wp *WorkPool) Publish(room string, push chan Pack, pk Pack) {

// 	item := amf.Decode(pk.Content)
// 	resp := []amf.Value{"onMetaData", item[2]}

// 	wp.MateList[room] = Pack{
// 		Type:    pk.Type,
// 		Time:    pk.Time,
// 		Content: amf.Encode(resp),
// 	}

// 	s := fmt.Sprint(time.Now().Unix())
// 	var flv av.FLV
// 	flv.GenFlv(room + s)
// 	flv.AddTag(pk.Type, pk.Time, pk.Content[16:])

// 	go func() {
// 		//存储解码信息。
// 		for pck := range push {
// 			flv.AddTag(pck.Type, pck.Time, pck.Content)
// 			if pck.Type == 9 {
// 				wp.VideoList[room] = pck
// 			}
// 			if pck.Type == 8 {
// 				wp.AudioList[room] = pck
// 			}
// 			if _, ok := wp.VideoList[room]; ok {
// 				if _, ok := wp.AudioList[room]; ok {
// 					break
// 				}
// 			}
// 		}

// 		for pck := range push {
// 			flv.AddTag(pck.Type, pck.Time, pck.Content)
// 			for py := range wp.PlayList[room] {
// 				py <- pck
// 			}
// 		}
// 		defer flv.Close()
// 	}()
// }

// func (wp *WorkPool) Play(room string, play chan Pack) {
// 	if _, ok := wp.PlayList[room]; !ok {
// 		wp.PlayList[room] = make(map[chan Pack]bool)
// 	}
// 	wp.PlayList[room][play] = true
// }

// func newWorkPool() *WorkPool {
// 	wp := &WorkPool{
// 		PlayList:  make(map[string]map[chan Pack]bool),
// 		MateList:  make(map[string]Pack),
// 		VideoList: make(map[string]Pack),
// 		AudioList: make(map[string]Pack),
// 	}
// 	return wp
// }
