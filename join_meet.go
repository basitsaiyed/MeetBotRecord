package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"math/rand"

	"github.com/playwright-community/playwright-go"
)

func main() {

	// Start recording
	recordCmd := startRecording()
	defer recordCmd.Process.Kill() // Stop recording after meeting ends

	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // Set to playwright.Bool(true) for background execution
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--use-fake-ui-for-media-stream", // Prevents camera/mic prompts
			"--use-fake-device-for-media-stream",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatal(err)
	}

	meetingURL := "https://meet.google.com/exg-euwh-xyw" // Replace with actual meeting link
	fmt.Println("Joining:", meetingURL)

	page.Goto(meetingURL)

	// Wait for the "Join as guest" form (if available)
	// time.Sleep(5 * time.Second)
	// Add random delays between actions
	time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)

	// Add mouse movements
	page.Mouse().Move(100, 100)

	// Simulate scrolling
	page.Mouse().Wheel(0, 100)

	// Use more reliable selectors
	nameInput := page.Locator("input[aria-label='Your name']")
	if nameInput != nil {
		err = nameInput.Fill("clavirion")
		if err != nil {
			log.Printf("Could not fill name: %v", err)
		}
	}

	// Wait for the join button
	GotItButton := page.Locator("button:has-text('Got it')")
	if GotItButton != nil {
		err = GotItButton.Click()
		if err != nil {
			log.Printf("Could not click join button: %v", err)
		}
	}

	// Wait for the join button
	joinButton := page.Locator("button:has-text('Join now')")
	if joinButton != nil {
		err = joinButton.Click()
		if err != nil {
			log.Printf("Could not click Join Now button: %v", err)
		}
	} else {
		// If "Join now" button is missing, try "Ask to join"
		askJoinButton := page.Locator("button:has-text('Ask to join')")
		if askJoinButton != nil {
			err = askJoinButton.Click()
			if err != nil {
				log.Printf("Could not click Ask to Join button: %v", err)
			}
		}
	}

	nameInput.Fill("clavirion") // Set guest name
	GotItButton.Click()         // Click "Got it" button
	fmt.Println("Bot requested to join the meeting.")
	time.Sleep(2 * time.Second)
	joinButton.Click() // Click join button
	fmt.Println("Bot joined the meeting.")

	go func() {
		browser.On("disconnected", func() {
			fmt.Println("Browser closed unexpectedly. Stopping recording...")
			recordCmd.Process.Signal(os.Interrupt)
			recordCmd.Wait()
		})
		fmt.Println("Browser closed unexpectedly. Stopping recording...")
		recordCmd.Process.Signal(os.Interrupt)
		recordCmd.Wait()
	}()

	// Wait for a predefined meeting duration
	time.Sleep(10 * time.Minute) // Adjust duration as needed

	fmt.Println("Meeting ended. Stopping recording...")
	recordCmd.Process.Signal(os.Interrupt)
	recordCmd.Wait()
	browser.Close()
	pw.Stop()

}

func startRecording() *exec.Cmd {
	cmd := exec.Command("ffmpeg", "-f", "gdigrab", "-framerate", "30", "-i", "desktop",
		"-f", "dshow", "-i", "audio=CABLE Output (VB-Audio Virtual Cable)",
		"-c:v", "libx264", "-preset", "ultrafast", "-c:a", "aac", "meeting_record.mp4")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal("Failed to start recording:", err)
	}

	fmt.Println("Recording started...")
	return cmd
}
