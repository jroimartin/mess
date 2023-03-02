package main

import "os/exec"

func browse(url string) error {
	cmd := exec.Command("xdg-open", url)
	return cmd.Start()
}
