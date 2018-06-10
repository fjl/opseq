This is my attempt to make a desktop version of the delightful sequencers built into the
Teenage Engineering OP-1. Right now it contains an incomplete implementation of Tombola.

To build, clone this project to your GOPATH and run

   go build

You can then run the sequencer:

   ./opseq -dev "usb" -inch 1 -outch 15
