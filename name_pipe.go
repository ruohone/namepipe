package namepipe

var processNamePipe *mNpipe

func StartListenPipe() error {
	namePipe, err := newNpipe(`\\.\pipe\gameserver`)
	if err != nil {
		return err
	}
	processNamePipe = namePipe

	err = processNamePipe.Listen()
	if err != nil {
		return err
	}
	return processNamePipe.Listen()
}

func StopListenPipe() {
	processNamePipe.StopListen()
}

func getSuccessMsgCh() <-chan string {
	return processNamePipe.getMsgCha()
}
