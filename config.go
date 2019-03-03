package tron

type Config struct {
	ReadBufSize   int // 读缓冲区大小
	WriteBufSize  int // 写缓冲区大小
	ReadChanSize  int // 异步读 channel 大小
	WriteChanSize int // 异步写 channel 大小
}

func NewConfig(rBufSize, wBufSize int, rChSize, wChSize int) *Config {
	c := &Config{
		ReadBufSize:   rBufSize,
		WriteBufSize:  wBufSize,
		ReadChanSize:  rChSize,
		WriteChanSize: wChSize,
	}
	return c
}
