package qflow

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/threecommaio/qflow/pkg/durable"
	// log "github.com/sirupsen/logrus"
)

type Endpoint struct {
	Name           string
	Hosts          []string
	Writer         chan interface{}
	DurableChannel chan interface{}
}

type Handler struct {
	Endpoints []Endpoint
}

func ReplicateChannel(endpoint *Endpoint) {
	var count int
	var sizeEndpoints = len(endpoint.Hosts)

	for {
		item := <-endpoint.DurableChannel
		req := item.(durable.Request)
		count++

		if count%1000 == 0 {
			fmt.Println("Processed batch of 1000")
		}

		r := bytes.NewReader(req.Body)
		url := fmt.Sprintf("%s%s", endpoint.Hosts[count%sizeEndpoints], req.URL)
		proxyReq, err := http.NewRequest(req.Method, url, r)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		timeout := time.Duration(5 * time.Second)
		client := &http.Client{Timeout: timeout}
		proxyRes, err := client.Do(proxyReq)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			endpoint.Writer <- item
			continue
		}
		io.Copy(ioutil.Discard, proxyRes.Body)
		proxyRes.Body.Close()
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	r := &durable.Request{Method: req.Method, URL: req.URL.String(), Body: body}
	for _, endpoint := range h.Endpoints {
		endpoint.Writer <- r
	}
}

func ListenAndServe(config *Config, addr string, dataDir string) {
	var ep []Endpoint

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err = os.MkdirAll(dataDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, endpoint := range config.Endpoints {

		writer := make(chan interface{})
		c := durable.Channel(writer, &durable.Config{
			Name:            endpoint.Name,
			DataPath:        dataDir,
			MaxBytesPerFile: 102400,
			MinMsgSize:      0,
			MaxMsgSize:      1000,
			SyncEvery:       10000,
			SyncTimeout:     time.Second * 10,
			Logger:          log.New(os.Stdout, "", 0),
		})

		e := &Endpoint{
			Name:           endpoint.Name,
			Hosts:          endpoint.Hosts,
			Writer:         writer,
			DurableChannel: c,
		}
		ep = append(ep, *e)

		go ReplicateChannel(e)

	}

	handler := &Handler{Endpoints: ep}
	http.HandleFunc("/", handler.HandleRequest)
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
