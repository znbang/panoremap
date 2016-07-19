package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"golang.org/x/text/encoding/traditionalchinese"
    "golang.org/x/text/transform"
)

func getImageSize(srcPath string) (width int, height int, err error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	image, format, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	if format != "jpeg" {
		return 0, 0, fmt.Errorf("Only support JPEG.")
	}

	return image.Width, image.Height, nil
}

func createScript(scriptPath string, srcPath string, width int, height int, yaw int, pitch int, roll int) (err error) {
	file, err := os.Create(scriptPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Big5 for Windows Traditional Chinese edition, don't care others
	var out io.Writer
	if runtime.GOOS == "windows" {
		out = transform.NewWriter(file, traditionalchinese.Big5.NewEncoder())
	} else {
		out = file
	}

	_, err = fmt.Fprintf(out, "p w%d h%d f2 v360 u0 n\"JPEG q100 g0\"\n", width, height)
	_, err = fmt.Fprintf(out, "o y%d p%d r%d f4 v360 n\"%s\"\n", yaw, pitch, roll, srcPath)
	
	return err
}

func remap(exePath string, scriptPath string, dstPath string) (err error) {
	cmd := exec.Command(exePath, "-o", dstPath, scriptPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func findExe() (exePath string, err error) {
	exeDir := filepath.Dir(os.Args[0])
	exeDir, err = filepath.Abs(exeDir)
	if err != nil {
		return "", err
	}
	exeList := []string{
		"PTStitcherNG",
		"PTStitcherNG.exe",
		"PTStitcherNG_cuda",
		"PTStitcherNG_cuda.exe",
	}

	for _, exe := range exeList {
		exePath = filepath.Join(exeDir, exe)
		if _, err := os.Stat(exePath); err == nil {
			return exePath, nil
		}
	}

	return "", fmt.Errorf("PTStitcherNG not found.")
}

func main() {
	fmt.Println("PanoRemap is a front end for PTStitcherNG.")
	fmt.Println("PTStitcherNG Copyright (c) 2008 2009 2010 Helmut Dersch.")
	fmt.Println("All rights reserved.")
	
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s filename.jpg\n", os.Args[0])
		return
	}

	exePath, err := findExe()
	if err != nil {
		fmt.Println(err)
		return
	}
	
	srcPath := os.Args[1]
	ext := filepath.Ext(srcPath)
	remapExt := "_remap" + ext
	newExt := "_new" + ext
	scriptPath := strings.TrimSuffix(srcPath, ext) + ".pts"

	width, height, err := getImageSize(srcPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	var dstPath string
	if strings.Contains(srcPath, remapExt) {
		// create new image
		dstPath = strings.TrimSuffix(srcPath, remapExt) + newExt
		err = createScript(scriptPath, srcPath, width, height, -90, 0, 90)
	} else {
		// remap image
		dstPath = strings.TrimSuffix(srcPath, ext) + remapExt
		err = createScript(scriptPath, srcPath, width, height, 90, 90, 0)
	}
	defer os.Remove(scriptPath)

	if err == nil {
		err = remap(exePath, scriptPath, dstPath)
	}

	if err != nil {
		fmt.Println(err)
		return
	}
}
