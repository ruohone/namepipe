package namepipe

import (
	"bufio"
	"fmt"
	"gopkg.in/natefinch/npipe.v2"
	"sync"
)

type mNpipe struct {
	PipeName  string
	Reader    *bufio.Reader
	MsgCh     chan string //<- chan 不能close
	SignalCh  chan string
	MsgSignal chan int
	Np        *npipe.PipeListener
	once      *sync.Once
}

func newNpipe(name string) (*mNpipe, error) {
	npipeListen, err := npipe.Listen(name)
	if err != nil {
		// handle error
		fmt.Println(err)
		return nil, err
	}

	return &mNpipe{
		PipeName:  name,
		MsgCh:     make(chan string, 4),
		SignalCh:  make(chan string, 1),
		MsgSignal: make(chan int, 1),
		Np:        npipeListen,
		once:      &sync.Once{},
	}, nil
}

func (p *mNpipe) Listen() error {
	go p.once.Do(func() {
		go p.accept()
		for {
			select {
			case <-p.SignalCh:
				{
					// fixme close 回导致debug报错。
					fmt.Println("listen end signal")
					return
				}
			case <-p.MsgSignal:
				{
					go p.accept()
				}
			}
		}
	})
	return nil
}

func (p *mNpipe) accept() {
	conn, err := p.Np.Accept()
	if err != nil {
		// handle error
		fmt.Println("accept pipe err:", err)
		_, notClose := <-p.MsgSignal
		if notClose { // close 导致的异常管道为空，因为已经关闭了。
			p.MsgSignal <- 0
		}
		return
	}

	r := bufio.NewReader(conn)
	b, _, err := r.ReadLine()
	if err != nil {
		// handle error
		conn.Close()
		fmt.Println("read err:", err)
		p.MsgSignal <- 0
		return
	}
	conn.Close()
	msg := string(b)
	fmt.Println(fmt.Sprintf("from pipe read msg:%s", msg))
	p.MsgCh <- msg
	p.MsgSignal <- 1
	return
}

func (p *mNpipe) StopListen() {
	p.SignalCh <- "1"
	if p.Np != nil {
		_ = p.Np.Close()
	}

	if p.SignalCh != nil {
		close(p.SignalCh)
	}

	if p.MsgCh != nil {
		close(p.MsgCh)
	}

	if p.MsgSignal != nil {
		close(p.MsgSignal)
	}
}

func (p *mNpipe) getMsgCha() <-chan string {
	return p.MsgCh
}
