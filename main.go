package main

import (
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Lshortfile)
	cfg := parseFlags()
	mi := openMIDI(cfg.devName)
	defer mi.close()
	log.Printf("Using %v", mi)
	runUI(newTombolaSeq(mi, cfg))
}

type config struct {
	inChannel  uint8
	outChannel uint8
	devName    string
}

func parseFlags() *config {
	var (
		in  = flag.Int("inch", 1, "input MIDI channel")
		out = flag.Int("outch", 2, "output MIDI channel")
		dev = flag.String("dev", "", "MIDI device")
	)
	flag.Parse()
	return &config{
		inChannel:  uint8(*in - 1),
		outChannel: uint8(*out - 1),
		devName:    *dev,
	}
}
