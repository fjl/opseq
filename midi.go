package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/scgolang/midi"
)

type midiInterface struct {
	dev      *midi.Device
	packetCh <-chan []midi.Packet
}

func openMIDI(devname string) *midiInterface {
	devs, err := midi.Devices()
	if err != nil {
		log.Fatalf("Can't initialize MIDI: %v", err)
	}
	if len(devs) == 0 {
		log.Fatalf("No MIDI devices")
	}
	if devname == "" {
		return openDevice(devs[0])
	}
	names := make([]string, len(devs))
	devname = strings.ToLower(devname)
	for i, dev := range devs {
		if matchDevice(dev, devname) {
			return openDevice(dev)
		}
		names[i] = dev.Name
	}
	log.Fatalf("Can't find MIDI device %q, have %v", devname, names)
	return nil
}

func matchDevice(dev *midi.Device, name string) bool {
	return strings.Contains(strings.ToLower(dev.Name), name) ||
		strings.Contains(strings.ToLower(dev.ID), name)
}

func openDevice(dev *midi.Device) *midiInterface {
	var err error
	if err = dev.Open(); err != nil {
		log.Fatalf("Can't open MIDI device: %v", err)
	}
	mi := &midiInterface{dev: dev}
	mi.packetCh, err = dev.Packets()
	if err != nil {
		log.Fatal("Can't create MIDI packet channel: %v", err)
	}
	return mi
}

func (mi *midiInterface) String() string {
	return fmt.Sprintf("MIDI interface %q (%s)", mi.dev.ID, mi.dev.Name)
}

func (mi *midiInterface) sendNoteOn(ch, pitch, vel uint8) {
	mi.dev.Write([]byte{0x90 | ch, pitch, vel})
}

func (mi *midiInterface) sendNoteOff(ch, pitch, vel uint8) {
	mi.dev.Write([]byte{0x80 | ch, pitch, vel})
}

func (mi *midiInterface) close() error {
	return mi.dev.Close()
}

func noteInfo(pkt [3]byte) (ch, pitch, vel uint8) {
	return pkt[0] & 0x0F, pkt[1], pkt[2]
}
