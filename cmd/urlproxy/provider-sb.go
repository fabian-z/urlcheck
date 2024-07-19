package main

import (
	"context"
	"log"
	"net/url"
	"os"

	"github.com/google/safebrowsing"
)

type SBProvider struct {
	sb *safebrowsing.SafeBrowser
}

func (sb *SBProvider) Init() error {

	var err error

	sb.sb, err = safebrowsing.NewSafeBrowser(safebrowsing.Config{
		APIKey: safeBrowsingAPIKey,
		Logger: os.Stdout,
	})

	if err != nil {
		return err
	}

	err = sb.sb.WaitUntilReady(context.Background())
	if err != nil {
		log.Fatal("Could not update Safe Browsing database:", err)
	}

	return nil
}

func (sb *SBProvider) Shutdown() error {
	return sb.sb.Close()
}

func (sb *SBProvider) Check(url *url.URL) (bool, error) {

	info, err := sb.sb.LookupURLs([]string{url.String()})
	if err != nil {
		return false, err
	}

	threatInfo := info[0]
	if len(threatInfo) > 0 {
		// Block any threat
		return true, nil
	}

	return false, nil
}
