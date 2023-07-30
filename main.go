package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Starters []*Starter
}

type Starter struct {
	Name string
	Cmd  string
	Args []string
	Btn  *widget.Clickable
}

func main() {
	cfg := &Config{}
	cfgfile := flag.String("config", "config.toml", "config file defines starter list")
	flag.Parse()
	target := ""
	args := flag.Args()
	if len(args) != 0 {
		target = args[0]
	}
	_, err := toml.DecodeFile(*cfgfile, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	if target != "" {
		for _, s := range cfg.Starters {
			s.Args = append(s.Args, target)
		}
	}
	go func() {
		w := app.NewWindow(
			app.Title("Starter"),
			app.Size(unit.Dp(200), unit.Dp(50*len(cfg.Starters))),
		)
		err := run(w, cfg)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window, cfg *Config) error {
	th := material.NewTheme(gofont.Collection())
	var ops op.Ops
	btns := make([]layout.FlexChild, 0, len(cfg.Starters))
	for _, s := range cfg.Starters {
		s.Btn = new(widget.Clickable)
		name := s.Name
		btn := s.Btn
		btns = append(btns, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			inset := layout.UniformInset(unit.Dp(5))
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				b := material.Button(th, btn, name)
				return b.Layout(gtx)
			})

		}))
	}
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			for _, s := range cfg.Starters {
				if s.Btn.Clicked() {
					if s.Cmd == "" {
						continue
					}
					cmd := exec.Command(s.Cmd, s.Args...)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Start()
					return err
				}
			}
			gtx := layout.NewContext(&ops, e)
			inset := layout.UniformInset(unit.Dp(5))
			inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(
					gtx,
					btns...,
				)
			})
			e.Frame(gtx.Ops)
		}
	}
	return nil
}
