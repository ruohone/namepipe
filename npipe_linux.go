package namepipe

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

type mNpipe struct {
	PipeName  string
	Reader    *bufio.Reader
	MsgCh     chan string //<- chan 不能close
	SignalCh  chan string
	MsgSignal chan int
	f         *os.File
	once      *sync.Once
}

func newNpipe(name string) (*mNpipe, error) {
	_, err := os.OpenFile(name, os.O_CREATE, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(name, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)

	return &mNpipe{
		PipeName:  name,
		MsgCh:     make(chan string, 4),
		SignalCh:  make(chan string, 1),
		MsgSignal: make(chan int, 1),
		Reader:    reader,
		f:         file,
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
					if file != nil {
						_ = file.Close()
					}
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

func (p *mNpipe) getMsgCha() <-chan string {
	return p.MsgCh
}

func (p *mNpipe) StopListen() {
	p.SignalCh <- "1"
	if p.f != nil {
		_ = p.f.Close()
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

func (p *mNpipe) accept() {
	for {
		if p.Reader == nil {
			p.MsgSignal <- 0
			return
		}

		//fixme 需要修改为 \0
		//line, _,err := p.Reader.ReadLine()
		line, err := p.Reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Println("read err:", err)
			p.MsgSignal <- 0
			return
		}
		if err == nil {
			msg := string(line)
			fmt.Println(fmt.Sprintf("from pipe read msg:%s", msg))
			p.MsgCh <- msg
			p.MsgSignal <- 1
			return
		}
	}
}
