package qflow

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/threecommaio/qflow/pkg/durable"
)

type Endpoint struct {
	Name           string
	Hosts          []string
	Writer         chan interface{}
	DurableChannel chan interface{}
	Timeout        time.Duration
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
			log.Debug("processed 1000 operations")
		}

		r := bytes.NewReader(req.Body)
		url := fmt.Sprintf("%s%s", endpoint.Hosts[count%sizeEndpoints], req.URL)
		proxyReq, err := http.NewRequest(req.Method, url, r)
		if err != nil {
			log.Debugf("error: %s\n", err)
			continue
		}

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}
		client := &http.Client{Transport: transport, Timeout: endpoint.Timeout}
		proxyRes, err := client.Do(proxyReq)
		if err != nil {
			log.Debugf("error: %s\n", err)
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
		log.Debugf("error reading body: %v", err)
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
	var timeout = config.HTTP.Timeout

	if timeout.Seconds() == 0.0 {
		timeout = 10 * time.Second
	}

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Infof("creating data directory: %s", dataDir)
		err = os.MkdirAll(dataDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, endpoint := range config.Endpoints {
		for _, host := range endpoint.Hosts {
			if !isValidUrl(host) {
				log.Fatalf("(%s) [%s] is not a valid endpoint url", endpoint.Name, host)
			}
		}

		log.Infof("registered (%s) with endpoints: [%s]", endpoint.Name, strings.Join(endpoint.Hosts, ","))
		log.Infof("config options: (http timeout: %s)", timeout)

		writer := make(chan interface{})
		c := durable.Channel(writer, &durable.Config{
			Name:            endpoint.Name,
			DataPath:        dataDir,
			MaxBytesPerFile: 102400,
			MinMsgSize:      0,
			MaxMsgSize:      1000,
			SyncEvery:       10000,
			SyncTimeout:     time.Second * 10,
		})

		e := &Endpoint{
			Name:           endpoint.Name,
			Hosts:          endpoint.Hosts,
			Writer:         writer,
			DurableChannel: c,
			Timeout:        timeout,
		}
		ep = append(ep, *e)

		go ReplicateChannel(e)

	}

	handler := &Handler{Endpoints: ep}
	http.HandleFunc("/", handler.HandleRequest)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// isValidUrl handles checking if a url is valid
func isValidUrl(s string) bool {
	url, err := url.ParseRequestURI(s)

	if url.Scheme != "http" && url.Scheme != "https" {
		return false
	}

	if err != nil {
		return false
	} else {
		return true
	}
}
