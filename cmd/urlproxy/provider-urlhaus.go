package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	urlHausList = "https://urlhaus.abuse.ch/downloads/text/"
	userAgent   = "kt-urlproxy/v0.1"
)

type URLHausProvider struct {
	lock sync.RWMutex
	list map[string]struct{}

	shutdown chan bool
	timer    *time.Ticker
}

func (up *URLHausProvider) Init() error {
	log.Println("Initializing URLHaus list")

	err := up.updateList()
	if err != nil {
		return err
	}

	up.shutdown = make(chan bool)

	// List is updated all 5 minutes
	up.timer = time.NewTicker(10 * time.Minute)

	go func() {
		for {
			select {
			case <-up.shutdown:
				return
			case <-up.timer.C:
				log.Println("Updating URLHaus list")
				err := up.updateList()
				if err != nil {
					log.Println("Error updating URLHaus list:", err)
				}
			}
		}
	}()

	return nil
}

func (up *URLHausProvider) updateList() error {

	req, err := http.NewRequest("GET", urlHausList, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code %v", resp.StatusCode)
	}

	list, err := io.ReadAll(io.LimitReader(resp.Body, 50*1024*1024))

	if err != nil {
		return err
	}

	entries := bytes.Count(list, []byte{"\n"[0]})
	newList := make(map[string]struct{}, entries)

	scan := bufio.NewScanner(bytes.NewReader(list))
	for scan.Scan() {
		newList[strings.TrimSpace(scan.Text())] = struct{}{}
	}

	up.lock.Lock()
	up.list = newList
	defer up.lock.Unlock()

	return nil
}

func (up *URLHausProvider) Shutdown() error {
	up.shutdown <- true
	up.timer.Stop()
	return nil
}

func (up *URLHausProvider) Check(url *url.URL) (bool, error) {

	up.lock.RLock()
	defer up.lock.RUnlock()

	_, ok := up.list[url.String()]
	if ok {
		return true, nil
	} else {
		return false, nil
	}

}
