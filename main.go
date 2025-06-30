package main

import (
    "fmt"
    "os"
    "os/exec"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
    "time"
)

// clearScreen attempts to clear the terminal screen.
// It tries 'clear' for Unix-like systems and 'cls' for Windows.
func clearScreen() {
    var cmd *exec.Cmd
    // Determine the correct clear command based on the operating system
    if os.Getenv("OS") == "Windows_NT" {
        cmd = exec.Command("cmd", "/c", "cls")
    } else {
        cmd = exec.Command("clear")
    }
    // We don't care about potential errors here, as clearing might not work
    // in all terminal environments. We just try our best.
    cmd.Stdout = os.Stdout
    cmd.Run()
}

func main() {
    // Argument parsing and validation
    if len(os.Args) < 3 {
        fmt.Println("Usage: go run main.go <interval_seconds> <command_to_run>")
        fmt.Println("Example: go run main.go 2 \"ls -l --color=always\"") // Added --color=always for example
        fmt.Println("         go run main.go 1 \"date\"")
        os.Exit(1)
    }

    intervalStr := os.Args[1]
    interval, err := strconv.ParseFloat(intervalStr, 64)
    if err != nil || interval <= 0 {
        fmt.Printf("Error: Invalid interval provided. It must be a positive number: %v\n", err)
        os.Exit(1)
    }

    // Join all subsequent arguments to form the complete command string
    commandToRun := strings.Join(os.Args[2:], " ")

    time.Sleep(1 * time.Second) // Give the user a moment to read the message

    // Set up signal handling to gracefully exit on Ctrl+C (SIGINT) or SIGTERM
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Use a ticker for precise, regular intervals
    ticker := time.NewTicker(time.Duration(interval * float64(time.Second)))
    defer ticker.Stop() // Ensure the ticker is stopped when main exits

    // Main loop for refreshing the command output
    for {
        select {
        case <-sigChan:
            // If Ctrl+C is caught, clean up and exit
            fmt.Println("\nExiting watch alternative.")
            clearScreen() // Clear the screen one last time for a clean terminal
            os.Exit(0)
        case <-ticker.C:
            // --- NEW FLICKER-FREE SEQUENCE ---

            // 1. Prepare the command to be executed
            var cmd *exec.Cmd
            if os.Getenv("OS") == "Windows_NT" {
                // On Windows, commands are typically run via 'cmd /c'
                cmd = exec.Command("cmd", "/c", commandToRun)
            } else {
                // On Unix-like systems, commands are run via 'sh -c'
                cmd = exec.Command("sh", "-c", commandToRun)
            }

            // 2. Execute the command and wait for it to finish, capturing all output
            //    This step embodies "execute", "wait", and "done"
            output, cmdErr := cmd.CombinedOutput()

            // 3. NOW that the new output is fully ready, clear the screen
            clearScreen()

            // 4. Print the header and the captured output
            fmt.Print(string(output)) // Print the captured output, including color codes

            // Handle any errors from the executed command, printing them after the main output
            if cmdErr != nil {
                if exitErr, ok := cmdErr.(*exec.ExitError); ok {
                    fmt.Printf("\n--- Command exited with non-zero status: %d ---\n", exitErr.ExitCode())
                } else {
                    fmt.Printf("\nError running command: %v\n", cmdErr)
                }
            }
            // The `ticker.C` will handle the waiting for the next interval automatically
        }
    }
}
