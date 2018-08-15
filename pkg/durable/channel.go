package durable

// Original: github.com/dutchcoders/durable

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"time"
)

type Request struct {
	Method string
	URL    string
	Body   []byte
}

type Config struct {
	Name            string
	DataPath        string
	MaxBytesPerFile int64
	MinMsgSize      int32
	MaxMsgSize      int32
	SyncEvery       int64
	SyncTimeout     time.Duration
	Logger          *log.Logger
}

func defaultConfig() *Config {
	return &Config{
		Name:            "",
		DataPath:        "./data",
		MaxBytesPerFile: 102400,
		MinMsgSize:      0,
		MaxMsgSize:      1000,
		SyncEvery:       10,
		SyncTimeout:     time.Second * 10,
		Logger:          log.New(ioutil.Discard, "", 0),
	}
}

type channel struct {
	in     chan interface{}
	out    chan interface{}
	dq     *diskQueue
	config *Config
}

func newChannel(c chan interface{}, config *Config) chan interface{} {
	out := make(chan interface{})

	b := channel{
		in:     c,
		out:    out,
		config: config,
	}

	b.dq = newDiskQueue(config)

	go b.reader()
	go b.writer()

	return out
}

func (b channel) reader() {
	for data := range b.dq.ReadChan() {
		var item Request
		dec := gob.NewDecoder(bytes.NewReader(data))
		if err := dec.Decode(&item); err != nil {
			b.config.Logger.Printf("Error unmarshalling object: %s\n", err.Error())
		}
		b.out <- item
	}
}

func (b channel) writer() {
	for {
		var network bytes.Buffer
		item := <-b.in
		enc := gob.NewEncoder(&network)

		if err := enc.Encode(item); err != nil {
			b.config.Logger.Printf("Error marshalling object: %s\n", err.Error())
		} else if err := b.dq.Put(network.Bytes()); err != nil {
			b.config.Logger.Printf("Error putting object: %s\n", err.Error())
		}
	}
}

func Channel(c chan interface{}, config *Config) chan interface{} {
	if config == nil {
		config = defaultConfig()
	}

	return newChannel(c, config)
}
