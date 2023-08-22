package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	asciiChars     = []string{" ", ".", ",", ":", ";", "+", "*", "?", "%", "S", "#", "@"}
	videoPath      = "videos/bad-apple.mp4"
	outputFile     = "output.txt"
	ffmpegPath     = "ffmpeg-master-latest-win64-gpl/bin/"
	terminalWidth  = 120
	terminalHeight = 30
	targetFPS      = 60.0
)

func main() {
	// Get video duration using FFmpeg
	duration, err := getVideoDuration(videoPath, ffmpegPath)
	if err != nil {
		fmt.Println("Error getting video duration:", err)
		return
	}

	// Create and open the output text file
	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer func(output *os.File) {
		err := output.Close()
		if err != nil {
			fmt.Println("Error closing output file:", err)
		}
	}(output)

	// Set the terminal size to match the ASCII dimensions
	setTerminalSize(terminalWidth, terminalHeight)

	// Create an FFmpeg command to generate ASCII frames
	cmd := exec.Command(ffmpegPath+"ffmpeg", "-i", videoPath, "-vf", fmt.Sprintf("scale=%d:%d,format=gray", terminalWidth, terminalHeight), "-f", "image2pipe", "-vcodec", "rawvideo", "-pix_fmt", "gray", "-")
	cmd.Stderr = os.Stderr

	// Start FFmpeg command and get the stdout pipe
	ffmpegOut, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error creating FFmpeg stdout pipe:", err)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting FFmpeg:", err)
		return
	}

	// Read ASCII frames from FFmpeg output and write them to the text file
	reader := bufio.NewReader(ffmpegOut)
	frameSize := terminalWidth * terminalHeight

	// Calculate the delay between each frame to simulate animation speed
	totalFrames := int(duration * targetFPS)

	for i := 0; i < totalFrames; i++ {
		frameData := make([]byte, frameSize)
		_, err := io.ReadFull(reader, frameData)
		if err != nil {
			fmt.Println("Error reading frame data:", err)
			return
		}
		// Convert grayscale values to ASCII characters
		asciiFrame := convertToASCII(frameData, terminalWidth, terminalHeight)

		// Write ASCII frame to the output file and console
		_, _ = output.WriteString(asciiFrame)
		fmt.Print(asciiFrame)

		// Calculate the expected time for the next frame
		expectedTime := time.Now().Add(time.Second / time.Duration(targetFPS))

		// Calculate the remaining time until the expected frame time
		remainingTime := expectedTime.Sub(time.Now())

		// Wait until the expected frame time
		time.Sleep(remainingTime)
	}

	// Wait for FFmpeg to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for FFmpeg:", err)
		return
	}
	fmt.Println("ASCII animation generated and saved to", outputFile)
}

func getVideoDuration(videoPath string, ffmpegPath string) (float64, error) {
	cmd := exec.Command(ffmpegPath+"ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	durationStr, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	duration, err := strconv.ParseFloat(strings.TrimSpace(string(durationStr)), 64)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func setTerminalSize(width, height int) {
	cmd := exec.Command("mode", fmt.Sprintf("con:cols=%d lines=%d", width, height))
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return
	}
}

func convertToASCII(frameData []byte, width int, height int) string {
	var asciiFrame strings.Builder

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := frameData[y*width+x]
			// Normalize the pixel value to the range 0-255
			normalizedPixel := float64(pixel) / 255.0
			// Map normalized pixel value to ASCII character index
			index := int(normalizedPixel * float64(len(asciiChars)-1))
			asciiFrame.WriteString(asciiChars[index])
		}
		asciiFrame.WriteString("\n")
	}
	return asciiFrame.String()
}
