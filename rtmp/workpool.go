package rtmp

type Play interface {
	//
}

type Publish interface {
	//
}

type WorkPool struct {
	//
}

func (wp *WorkPool) addPlayer(play *Play) error {
	return nil
}

func (wp *WorkPool) addPublisher(play *Publish) error {
	return nil
}

func (wp *WorkPool) run() {

}

func newWorkPool() *WorkPool {
	return &WorkPool{}
}
