package main

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/jakecoffman/cp"
	"github.com/scgolang/midi"
	"github.com/shibukawa/nanovgo"
)

const (
	// Physics parameters:
	gravity           = 5.0
	ballRadius        = 0.03
	ballMass          = 0.05
	ballElasticity    = 0.9
	segmentSides      = 6
	segmentElasticity = 0.3
	segmentRotation   = 0.6
	physUpdateCycle   = 10 * time.Millisecond
	collSegment       = cp.CollisionType(1)
	collNote          = cp.CollisionType(2)
)

var (
	// Drawing-related parameters:
	segmentThickness = float32(1.0)
	segmentColor     = nanovgo.RGBf(0.8, 0.8, 0.8)
	ballColor        = nanovgo.RGBf(1.0, 1.0, 1.0)
)

type tombolaSeq struct {
	inChannel  uint8
	outChannel uint8
	midi       *midiInterface

	mu     sync.Mutex
	space  *cp.Space
	seg    *cp.Body
	notes  map[int]*tombolaNote
	noteID int
}

type tombolaNote struct {
	pitch uint8
	body  *cp.Body
}

func newTombolaSeq(mi *midiInterface, config *config) *tombolaSeq {
	t := &tombolaSeq{
		inChannel:  config.inChannel,
		outChannel: config.outChannel,
		midi:       mi,
		notes:      make(map[int]*tombolaNote),
	}
	t.setupPhysics()
	h := t.space.NewCollisionHandler(collNote, collSegment)
	h.PostSolveFunc = t.handleCollision
	go t.run()
	return t
}

func (t *tombolaSeq) setupPhysics() {
	// Set up the space.
	t.space = cp.NewSpace()
	t.space.SetGravity(cp.Vector{Y: gravity})

	// Create the container.
	t.seg = cp.NewKinematicBody()
	t.seg.SetAngularVelocity(segmentRotation)
	v1 := cp.Vector{1, 0}
	for i := 0; i < segmentSides; i++ {
		angle := (2 * math.Pi / segmentSides) * float64(i+1)
		v2 := cp.Vector{math.Cos(angle), math.Sin(angle)}
		shape := cp.NewSegment(t.seg, v1, v2, 0.01)
		shape.SetElasticity(segmentElasticity)
		shape.SetCollisionType(collSegment)
		t.space.AddShape(shape)
		v1 = v2
	}
	t.space.AddBody(t.seg)
}

func (t *tombolaSeq) run() {
	var (
		physUpdate = time.NewTicker(physUpdateCycle)
		lastTime   = time.Now()
	)
	for {
		select {
		case now := <-physUpdate.C:
			delta := now.Sub(lastTime)
			lastTime = now
			t.mu.Lock()
			t.space.Step(float64(delta) / float64(time.Second))
			t.deleteDropouts()
			t.mu.Unlock()

		case pkts := <-t.midi.packetCh:
			t.mu.Lock()
			t.processMIDI(pkts)
			t.mu.Unlock()
		}
	}
}

func (t *tombolaSeq) deleteDropouts() {
	const max = 3.0
	for id, n := range t.notes {
		pos := n.body.Position()
		if pos.X > max || pos.X < -max || pos.Y > max || pos.Y < -max {
			delete(t.notes, id)
			t.space.RemoveBody(n.body)
		}
	}
}

// ------------------------------------------------------------------------------
// MIDI

func (t *tombolaSeq) processMIDI(packets []midi.Packet) {
	for _, pkt := range packets {
		if pkt.Err != nil {
			log.Fatalf("MIDI error: %v", pkt.Err)
			continue
		}
		if midi.GetMessageType(pkt) == midi.MessageTypeNoteOn {
			if ch, pitch, _ := noteInfo(pkt.Data); ch == t.inChannel {
				t.addNote(pitch)
			}
		}
	}
}

func (t *tombolaSeq) addNote(pitch uint8) {
	n := &tombolaNote{
		pitch: pitch,
		body:  cp.NewBody(ballMass, cp.MomentForCircle(ballMass, ballRadius, 0, cp.Vector{})),
	}
	n.body.UserData = n
	shape := cp.NewCircle(n.body, ballRadius, cp.Vector{})
	shape.SetElasticity(ballElasticity)
	shape.SetCollisionType(collNote)
	n.body.SetMass(ballMass)
	t.space.AddBody(n.body)
	t.space.AddShape(shape)

	t.notes[t.noteID] = n
	t.noteID++
}

func (t *tombolaSeq) handleCollision(a *cp.Arbiter, s *cp.Space, d interface{}) {
	if !a.IsFirstContact() {
		return
	}
	nb, _ := a.Bodies()
	note := nb.UserData.(*tombolaNote)
	impulse := math.Min(1.0, a.TotalImpulse().Length()*10)
	vel := uint8(60 + 60*impulse)
	log.Printf("Note out: channel %d, pitch %d, velocity %d", t.outChannel, note.pitch, vel)
	t.midi.sendNoteOn(t.outChannel, note.pitch, vel)
}

// ------------------------------------------------------------------------------
// UI

func (t *tombolaSeq) draw(ctx *nanovgo.Context, winWidth, winHeight float32) {
	ctx.Translate(winWidth/2, winHeight/2)
	scale := float64(winWidth) / 3
	t.mu.Lock()
	t.doDraw(ctx, scale)
	t.mu.Unlock()
}

func (t *tombolaSeq) doDraw(ctx *nanovgo.Context, scale float64) {
	// Draw segments.
	ctx.Save()
	ctx.Rotate(float32(t.seg.Angle()))
	t.seg.EachShape(func(s *cp.Shape) {
		seg := s.Class.(*cp.Segment)
		a, b := seg.A().Mult(scale), seg.B().Mult(scale)
		ctx.BeginPath()
		ctx.MoveTo(float32(a.X), float32(a.Y))
		ctx.LineTo(float32(b.X), float32(b.Y))
		ctx.SetStrokeColor(segmentColor)
		ctx.SetStrokeWidth(segmentThickness)
		ctx.Stroke()
	})
	ctx.Restore()

	// Draw notes.
	ctx.SetFillColor(ballColor)
	for _, n := range t.notes {
		ctx.BeginPath()
		pos := n.body.Position().Mult(scale)
		ctx.Circle(float32(pos.X), float32(pos.Y), float32(ballRadius*scale))
		ctx.Fill()
	}
}
