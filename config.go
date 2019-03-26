package tron

import "time"

type Config struct {
	ReadBufSize   int           // 读缓冲区大小
	WriteBufSize  int           // 写缓冲区大小
	ReadChanSize  int           // 异步读 channel 大小
	WriteChanSize int           // 异步写 channel 大小
	IdleDuration  time.Duration // 连接的最大空闲时间
	SeqManager    *SeqManager   // 包序号管理
}

func NewDefaultConf(idle time.Duration) *Config {
	return NewConfig(16*1024, 16*1024, 100, 100, 1000, idle)
}

func NewConfig(rBufSize, wBufSize int, rChSize, wChSize int, maxSeq int32, idle time.Duration) *Config {
	c := &Config{
		ReadBufSize:   rBufSize,
		WriteBufSize:  wBufSize,
		ReadChanSize:  rChSize,
		WriteChanSize: wChSize,
		IdleDuration:  idle,
		SeqManager:    NewSeqManager(maxSeq),
	}
	return c
}
