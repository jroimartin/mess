/*
capture-screen captures one of the active screens. It is useful when you want
to share a single screen from an application that only allows to share a window
or the full desktop (all the screens).

It depends on xrandr and ffplay.
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

type screen struct {
	name    string
	width   int
	height  int
	x       int
	y       int
	primary bool
}

func (scr *screen) String() string {
	p := ""
	if scr.primary {
		p = "*"
	}
	return fmt.Sprintf("%v%v\t%vx%v+%v+%v", scr.name, p, scr.width, scr.height, scr.x, scr.y)
}

func (scr *screen) Capture() error {
	xdisplay := os.Getenv("DISPLAY")
	if xdisplay == "" {
		return errors.New("missing env var DISPLAY")
	}

	cmd := exec.Command(
		"ffplay",
		"-f", "x11grab",
		"-i", fmt.Sprintf("%s+%v,%v", xdisplay, scr.x, scr.y),
		"-s", fmt.Sprintf("%vx%v", scr.width, scr.height),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffplay error: %w", err)
	}

	return nil
}

func main() {
	list := flag.Bool("list", false, "list screens")
	flag.Usage = usage
	flag.Parse()

	if err := checkDependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "missing dependency: %v\n", err)
		os.Exit(1)
	}

	if *list {
		if err := listScreens(); err != nil {
			fmt.Fprintf(os.Stderr, "could not list screens: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	switch flag.NArg() {
	case 0:
		if err := captureScreen(""); err != nil {
			fmt.Fprintf(os.Stderr, "could not capture primary screen: %v\n", err)
			os.Exit(1)
		}
	case 1:
		name := flag.Arg(0)
		if err := captureScreen(name); err != nil {
			fmt.Fprintf(os.Stderr, "could not capture screen %v: %v\n", name, err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func checkDependencies() error {
	cmd := exec.Command("ffplay", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffplay error: %w", err)
	}

	cmd = exec.Command("xrandr", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xrandr error: %w", err)
	}

	return nil
}

func listScreens() error {
	screens, err := getScreens()
	if err != nil {
		return fmt.Errorf("could not get screens: %w", err)
	}

	for _, scr := range screens {
		fmt.Println(scr)
	}

	return nil
}

func captureScreen(name string) error {
	screens, err := getScreens()
	if err != nil {
		return fmt.Errorf("could not get screens: %w", err)
	}

	for _, scr := range screens {
		if scr.name == name || (scr.primary && name == "") {
			if err := scr.Capture(); err != nil {
				return fmt.Errorf("could not capture screen: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("screen not found: %v", name)
}

var xrandrRe = regexp.MustCompile(`(?m)^([a-zA-Z0-9-]+) connected (primary )?(\d+)x(\d+)\+(\d+)\+(\d+).+$`)

func getScreens() ([]*screen, error) {
	cmd := exec.Command("xrandr")
	out, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, fmt.Errorf("xrandr error: %w: %s", exitError, exitError.Stderr)
		}
		return nil, fmt.Errorf("xrandr error: %w", err)
	}

	matches := xrandrRe.FindAllStringSubmatch(string(out), -1)

	var screens []*screen
	for _, m := range matches {
		width, err := strconv.Atoi(m[3])
		if err != nil {
			return nil, fmt.Errorf("invalid width: %v", m[3])
		}
		height, err := strconv.Atoi(m[4])
		if err != nil {
			return nil, fmt.Errorf("invalid height: %v", m[4])
		}
		x, err := strconv.Atoi(m[5])
		if err != nil {
			return nil, fmt.Errorf("invalid x: %v", m[5])
		}
		y, err := strconv.Atoi(m[6])
		if err != nil {
			return nil, fmt.Errorf("invalid y: %v", m[6])
		}

		scr := &screen{
			name:    m[1],
			width:   width,
			height:  height,
			x:       x,
			y:       y,
			primary: m[2] != "",
		}
		screens = append(screens, scr)
	}

	return screens, nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: capture-screen [flags] [screen]")
	fmt.Fprintln(os.Stderr, "flags:")
	flag.PrintDefaults()
}
