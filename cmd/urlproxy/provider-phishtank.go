package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	phishTankURL       = "http://data.phishtank.com/data/online-valid.json.gz"
	userAgentPhishtank = "phishtank/kt-urlproxy-v0.1"
)

type PhishTankRecord struct {
	PhishID          int       `json:"phish_id"`
	URL              string    `json:"url"`
	PhishDetailURL   string    `json:"phish_detail_url"`
	SubmissionTime   time.Time `json:"submission_time"`
	Verified         string    `json:"verified"`
	VerificationTime time.Time `json:"verification_time"`
	Online           string    `json:"online"`
	Details          []struct {
		IPAddress         string    `json:"ip_address"`
		CidrBlock         string    `json:"cidr_block"`
		AnnouncingNetwork string    `json:"announcing_network"`
		Rir               string    `json:"rir"`
		Country           string    `json:"country"`
		DetailTime        time.Time `json:"detail_time"`
	} `json:"details"`
	Target string `json:"target"`
}

type PhishTankProvider struct {
	lock sync.RWMutex
	list map[string]struct{}

	shutdown chan bool
	timer    *time.Ticker
}

func (pp *PhishTankProvider) Init() error {
	log.Println("Initializing PhishTank list")

	err := pp.updateList()
	if err != nil {
		return err
	}

	pp.shutdown = make(chan bool)

	// List is updated all 60 minutes
	// Rate Limits allow for a maximum of 75 fetches in 72h
	// TODO try to get user account by personal contacts, registration disabled
	pp.timer = time.NewTicker(120 * time.Minute)

	go func() {
		for {
			select {
			case <-pp.shutdown:
				return
			case <-pp.timer.C:
				log.Println("Updating URLHaus list")
				err := pp.updateList()
				if err != nil {
					log.Println("Error updating URLHaus list:", err)
				}
			}
		}
	}()

	return nil
}

func (pp *PhishTankProvider) updateList() error {

	req, err := http.NewRequest("GET", phishTankURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgentPhishtank)

	// Rate Limit has issues with IPv6, use IPv4 instead
	var dialer net.Dialer
	var client http.Client
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp4", addr)
	}
	client.Transport = transport

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	compressedReader, err := gzip.NewReader(io.LimitReader(resp.Body, 50*1024*1024))
	if err != nil {
		return err
	}

	jsonList, err := io.ReadAll(compressedReader)
	if err != nil {
		return err
	}

	var records []PhishTankRecord
	err = json.Unmarshal(jsonList, &records)
	if err != nil {
		return err
	}

	newList := make(map[string]struct{}, len(records))

	for _, record := range records {
		if record.Verified == "yes" {
			newList[record.URL] = struct{}{}
		}
	}

	log.Printf("Got %d verified PhishTank Records", len(newList))

	pp.lock.Lock()
	pp.list = newList
	defer pp.lock.Unlock()

	return nil
}

func (pp *PhishTankProvider) Shutdown() error {
	pp.shutdown <- true
	pp.timer.Stop()
	return nil
}

func (pp *PhishTankProvider) Check(url *url.URL) (bool, error) {

	pp.lock.RLock()
	defer pp.lock.RUnlock()

	_, ok := pp.list[url.String()]
	if ok {
		return true, nil
	} else {
		return false, nil
	}

}
