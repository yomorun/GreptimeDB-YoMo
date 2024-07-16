package main

import (
	"flag"
	"log"

	"github.com/nxadm/tail"
	"github.com/yomorun/yomo"
)

var (
	tag        = uint32(0xC001)
	credential = flag.String("c", "", "yomo zipper credential")
	zipperAddr = flag.String("z", "127.0.0.1:9000", "yomo zipper address")
	watchFile  = flag.String("f", "", "file to watch")
)

func init() {
	flag.Parse()
}

func main() {
	flag.Parse()

	source, err := newSource(*zipperAddr, *credential)
	if err != nil {
		log.Fatalln(err)
	}

	if *watchFile == "" {
		log.Fatalln("no file to watch, use -f flag to specify")
	}

	t, err := tail.TailFile(*watchFile, tail.Config{ReOpen: true, Follow: true, Logger: tail.DiscardingLogger})
	if err != nil {
		log.Fatalln(err)
	}
	for line := range t.Lines {
		log.Println(line.Text)
		if err := source.Write(tag, []byte(line.Text)); err != nil {
			log.Println(err)
			continue
		}
	}
}

func newSource(zipperAddr string, credential string) (yomo.Source, error) {
	opts := []yomo.SourceOption{
		yomo.WithSourceReConnect(),
	}
	if credential != "" {
		opts = append(opts, yomo.WithCredential(credential))
	}
	source := yomo.NewSource(
		"linner-source",
		zipperAddr,
		opts...,
	)
	if err := source.Connect(); err != nil {
		return nil, source.Connect()
	}
	return source, nil
}
