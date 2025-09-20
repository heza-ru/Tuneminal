package main

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Create application
	app := tview.NewApplication()
	
	// Create a simple text view
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	
	textView.SetText(`[white]╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║  [yellow]TUNEMINAL KARAOKE MACHINE[white]                              ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝

[green]Welcome to Tuneminal![white]

[cyan]Features:[white]
• Real-time karaoke lyrics
• Audio visualizer
• Scoring system
• Professional interface

[cyan]Controls:[white]
• [yellow]Space[white] - Play/Pause
• [yellow]Q[white] - Quit
• [yellow]H[white] - Help

[yellow]Press 'q' to quit or 'h' for help[white]`)
	
	textView.SetTextAlign(tview.AlignCenter)
	textView.SetBorder(true)
	textView.SetTitle("[blue]Tuneminal Karaoke[white]")
	
	// Set up key bindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'h':
				// Show help
				helpModal := tview.NewModal().
					SetText("Tuneminal Karaoke Help\n\n" +
						"Navigation: ↑/↓ to navigate songs\n" +
						"Playback: Space to play/pause\n" +
						"Karaoke: Sing along with highlighted lyrics\n" +
						"Scoring: Complete lines for points\n\n" +
						"Press any key to continue").
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						app.SetRoot(textView, true)
					})
				app.SetRoot(helpModal, true)
				return nil
			}
		}
		return event
	})
	
	// Run the application
	if err := app.SetRoot(textView, true).Run(); err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Thanks for using Tuneminal!")
}
