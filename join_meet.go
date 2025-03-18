package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"math/rand"

	"github.com/playwright-community/playwright-go"
)

func main() {
	// Start recording
	recordCmd := startRecording()
	defer func() {
		fmt.Println("Stopping recording...")
		recordCmd.Process.Signal(os.Interrupt)
		recordCmd.Wait()
	}()

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Failed to start Playwright: %v", err)
	}
	defer pw.Stop()

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // Set to true for background execution
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--use-fake-ui-for-media-stream",
			"--use-fake-device-for-media-stream",
		},
	})
	if err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	defer browser.Close()

	// Create new page
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}

	// Meeting configuration
	meetingURL := "https://meet.google.com/fho-nohr-kdg" // Replace with actual meeting link
	guestName := "clavirion"                             // Set guest name
	meetingDuration := 10 * time.Minute                  // Set meeting duration

	fmt.Printf("Joining meeting: %s as %s\n", meetingURL, guestName)

	// Navigate to the meeting URL
	if _, err := page.Goto(meetingURL); err != nil {
		log.Fatalf("Failed to navigate to meeting URL: %v", err)
	}

	// Add random delay to simulate human behavior
	randomDelay(2, 5)

	// Add realistic user behavior
	simulateHumanBehavior(page)

	// Join the meeting
	if err := joinMeeting(page, guestName); err != nil {
		log.Printf("Error joining meeting: %v", err)
	}

	// Set up browser disconnect handler
	setupDisconnectHandler(browser, recordCmd)

	// Wait for the meeting duration
	fmt.Printf("Staying in meeting for %v\n", meetingDuration)
	time.Sleep(meetingDuration)

	fmt.Println("Meeting time completed. Exiting...")
}

// simulateHumanBehavior adds random mouse movements and scrolling to appear more human-like
func simulateHumanBehavior(page playwright.Page) {
	page.Mouse().Move(100+float64(rand.Intn(300)), 100+float64(rand.Intn(200)))
	page.Mouse().Wheel(0, 100)
	randomDelay(1, 2)
}

// joinMeeting handles the process of joining a Google Meet
func joinMeeting(page playwright.Page, guestName string) error {
	// Fill in name if the field is available
	nameInput := page.Locator("input[aria-label='Your name']")
	if nameInput != nil {
		isVisible, err := nameInput.IsVisible()
		if err == nil && isVisible {
			if err := nameInput.Fill(guestName); err != nil {
				log.Printf("Could not fill name: %v", err)
			} else {
				fmt.Println("Entered guest name")
			}
		}
	}

	// Click "Got it" button if visible
	handleButton(page, "button:has-text('Got it')", "Got it")

	// Ensure microphone and camera are off
	handleButton(page, "[aria-label='Turn off microphone']", "Turn off microphone")
	handleButton(page, "[aria-label='Turn off camera']", "Turn off camera")

	// Try to join the meeting - first try "Join now" button
	if !handleButton(page, "button:has-text('Join now')", "Join now") {
		// If "Join now" failed, try "Ask to join" button
		if !handleButton(page, "button:has-text('Ask to join')", "Ask to join") {
			return fmt.Errorf("could not find any join button")
		}
	}

	fmt.Println("Successfully requested to join the meeting")
	return nil
}

// handleButton attempts to click a button identified by selector
func handleButton(page playwright.Page, selector string, buttonName string) bool {
	button := page.Locator(selector)
	if button == nil {
		return false
	}

	isVisible, err := button.IsVisible()
	if err != nil || !isVisible {
		return false
	}

	if err := button.Click(); err != nil {
		log.Printf("Could not click %s button: %v", buttonName, err)
		return false
	}

	fmt.Printf("Clicked %s button\n", buttonName)
	randomDelay(1, 3)
	return true
}

// setupDisconnectHandler creates a handler for browser disconnects
func setupDisconnectHandler(browser playwright.Browser, recordCmd *exec.Cmd) {
	go func() {
		browser.On("disconnected", func() {
			fmt.Println("Browser closed unexpectedly. Stopping recording...")
			recordCmd.Process.Signal(os.Interrupt)
			recordCmd.Wait()
		})
	}()
}

// startRecording starts the FFmpeg process to record the meeting audio
func startRecording() *exec.Cmd {
	// Ensure the recordings folder exists
	recordingFolder := "recordings"
	if _, err := os.Stat(recordingFolder); os.IsNotExist(err) {
		os.Mkdir(recordingFolder, os.ModePerm)
	}

	// Generate a unique filename with timestamp
	filename := fmt.Sprintf("meeting_%s.mp3", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(recordingFolder, filename)

	// FFmpeg command to record audio
	cmd := exec.Command("ffmpeg",
		"-f", "dshow",
		"-i", "audio=CABLE Output (VB-Audio Virtual Cable)",
		"-ac", "2",       // Stereo
		"-ar", "44100",   // Sample rate 44.1kHz
		"-c:a", "libmp3lame", // MP3 codec
		"-b:a", "192k",   // Bitrate 192kbps
		filepath,
	)

	// Capture FFmpeg output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start FFmpeg recording: %v", err)
	}

	fmt.Println("Recording started:", filepath)
	return cmd
}

// randomDelay adds a random delay between actions to simulate human behavior
func randomDelay(min, max int) {
	delay := min + rand.Intn(max-min+1)
	time.Sleep(time.Duration(delay) * time.Second)
}