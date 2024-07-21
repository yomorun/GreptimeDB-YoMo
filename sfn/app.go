package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/yomorun/yomo/serverless"
)

var (
	globalWriter     *lineWriter
	greptimeHttpAddr = os.Getenv("GREPTIMEDB_HTTP_ADDR")
)

func Init() error {
	if greptimeHttpAddr == "" {
		return errors.New("GREPTIMEDB_HTTP_ADDR not set")
	}

	apiUrl := fmt.Sprintf("http://%s/v1/influxdb/write", greptimeHttpAddr)

	globalWriter = newLineWriter(apiUrl)
	return nil
}

func DataTags() []uint32 {
	return []uint32{0xC001}
}

func Handler(ctx serverless.Context) {
	data := ctx.Data()

	_, err := globalWriter.Write(data)
	if err != nil {
		log.Println(err)
	}
}

type lineWriter struct {
	influxURL  string
	baseHeader http.Header
	HTTPClient http.Client
}

func newLineWriter(influxURL string) *lineWriter {
	baseHeader := make(http.Header)

	return &lineWriter{
		influxURL:  influxURL,
		baseHeader: baseHeader,
		HTTPClient: http.Client{},
	}
}

func (g *lineWriter) Write(p []byte) (n int, err error) {
	req, err := http.NewRequest(http.MethodPost, g.influxURL, bytes.NewReader(p))
	if err != nil {
		return 0, err
	}
	req.Header = g.baseHeader

	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to send, code=%s: res=%s,res=%s", resp.Status, string(body), string(p))
	}

	return len(p), nil
}
