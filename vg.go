package main

import (
	"log"

	"github.com/goxjs/gl"
	"github.com/goxjs/glfw"
	"github.com/shibukawa/nanovgo"
)

type drawerVG interface {
	draw(ctx *nanovgo.Context, winWidth, winHeight float32)
}

func runUI(app drawerVG) {
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		log.Fatalf("Can't initialize OpenGL: %v", err)
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.Resizable, 0)

	window, err := glfw.CreateWindow(300, 300, "opseq", nil, nil)
	if err != nil {
		log.Fatalf("Can't open window: %v", err)
	}
	window.MakeContextCurrent()

	ctx, err := nanovgo.NewContext(0)
	if err != nil {
		log.Fatalf("Can't create NanoVG context: %v", err)
	}
	defer ctx.Delete()

	fbWidth, fbHeight := window.GetFramebufferSize()
	winWidth, winHeight := window.GetSize()
	pixelRatio := float32(fbWidth) / float32(winWidth)

	gl.Viewport(0, 0, fbWidth, fbHeight)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.BLEND)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	glfw.SwapInterval(1)
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
		ctx.BeginFrame(winWidth, winHeight, pixelRatio)
		app.draw(ctx, float32(winWidth), float32(winHeight))
		ctx.EndFrame()
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
