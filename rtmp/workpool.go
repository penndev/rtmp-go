package rtmp

// 存储视频元数据
type metaPack struct {
	meta  []byte
	audit []byte
	video []byte
}

// 播放端列表
type stream struct {
	list map[chan Pack]bool
	meta map[string]metaPack
}

type App struct {
	Gloab map[chan Pack]bool
	List  map[string]stream
}

func (a *App) addGloab(pk chan Pack) {
	a.Gloab[pk] = true
}

func (a *App) addPublish(appName string, streamName string) {
	app, ok := a.List[appName]
	if !ok {
		a.List[appName] = stream{
			list: make(map[chan Pack]bool),
			meta: make(map[string]metaPack),
		}
		app = a.List[appName]
	}
	for pk, bl := range a.Gloab {
		app.list[pk] = bl
	}
}

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
