package main

import (
	"github.com/nsf/termbox-go"
	"math/rand"
	"strconv"
	"time"
	"unicode/utf8"
)

const (
	AlignLeft = iota
	AlignCenter
	AlignRight
)

const (
	NoMedal = iota
	BronzeMedal
	SilverMedal
	GoldMedal
	PlatinumMedal
)

type Pipe struct {
	Xpos int
	Gap  int
}

type Clover struct {
	Xpos int
	Ypos int
}

// Global variables. I told you it wasn't pretty!
var running bool
var started bool
var gameover bool
var paused bool

var posx int
var posy int
var vely int

var wing rune

var width int
var height int
var framewidth int
var offset int

var score int
var medal int

var pipes []Pipe
var pipecolors [4]termbox.Cell
var pipespacing int
var pipecount int
var newpipe int
var pipeidx int

var clovers []Clover
var moveclover int

// This is a way to do this. It's not necessarily a good way to do it.
func Keyer() chan termbox.Event {
	ch := make(chan termbox.Event)
	go func() {
		for true {
			ch <- termbox.PollEvent()
		}
	}()
	return ch
}

func DrawPipe(xpos, gap int) {
	for x := 0; x < 4; x++ {
		if x+xpos >= 0 && x+xpos < framewidth {
			for y := 0; y < height; y++ {
				if y < gap || y > gap+5 {
					termbox.SetCell(x+offset+xpos, y, pipecolors[x].Ch, pipecolors[x].Fg, pipecolors[x].Bg)
				}
			}
		}
	}
}

func DrawClovers() {
	for i, _ := range clovers {
		if clovers[i].Xpos >= 0 && clovers[i].Xpos < framewidth {
			termbox.SetCell(clovers[i].Xpos+offset, clovers[i].Ypos, '♣', termbox.ColorGreen, termbox.ColorGreen|termbox.AttrBold)
		}
	}
}

func DrawCurtains() {
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x%2 == 0 {
				termbox.SetCell(x, y, '▒', termbox.ColorBlack, termbox.ColorRed)
			} else {
				termbox.SetCell(x, y, ' ', termbox.ColorBlack, termbox.ColorRed)
			}
		}
	}
}

// So I herd u liek unicodez
func DrawRunesColor(s []rune, x, y int, align int, fg, bg termbox.Attribute) {
	xoff := 0
	switch align {
	case AlignLeft:
		xoff = 0
	case AlignCenter:
		xoff = -len(s) / 2
	case AlignRight:
		xoff = -len(s)
	}
	for i := 0; i < len(s); i++ {
		r := s[i]
		termbox.SetCell(x+xoff+i, y, r, fg, bg)
	}
}

func DrawTextColor(s string, x, y int, align int, fg, bg termbox.Attribute) {
	xoff := 0
	switch align {
	case AlignLeft:
		xoff = 0
	case AlignCenter:
		xoff = -len(s) / 2
	case AlignRight:
		xoff = -len(s)
	}
	for i := 0; i < len(s); i++ {
		r, _ := utf8.DecodeLastRuneInString(string(s[i])) // (([])) !!!
		termbox.SetCell(x+xoff+i, y, r, fg, bg)
	}
}

// White text on black shortcut
func DrawText(s string, x, y int, align int) { 
	DrawTextColor(s, x, y, align, termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlack)
}

func DrawScore() {
	s := strconv.Itoa(score)
	DrawText(s, framewidth/2+offset, 8, AlignCenter)
}

func DrawGuide() {
	DrawText("F: Flap ", width, height-4, AlignRight)
	termbox.SetCell(width-8, height-4, '↑', termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlack)
	DrawText("Q: Quit ", width, height-3, AlignRight)
	DrawText("R: Reset", width, height-2, AlignRight)
	DrawText("P: Pause", width, height-1, AlignRight)

	DrawText("AsciiBird 2014", 0, height-2, AlignLeft)
	DrawText("Justin K Phillips", 0, height-1, AlignLeft)
}

func Draw() {
	// Draw sky and ground
	for x := 0; x < framewidth; x++ {
		for y := 0; y < height; y++ {
			if y > height-6 {
				termbox.SetCell(x+offset, y, ' ', termbox.ColorGreen, termbox.ColorGreen|termbox.AttrBold)
			} else {
				termbox.SetCell(x+offset, y, ' ', termbox.ColorWhite, termbox.ColorCyan|termbox.AttrBold)
			}
		}
	}
	
	DrawClovers()
	
	// Draw all pipes
	for _, pipe := range pipes {
		if pipe.Xpos > -4 && pipe.Xpos < framewidth {
			DrawPipe(pipe.Xpos, pipe.Gap)
		}
	}
	
	// Draw bird
	if posy%6 < 3 {
		termbox.SetCell(posx+offset, posy/6, wing, termbox.ColorYellow, termbox.ColorYellow|termbox.AttrBold)
		termbox.SetCell(posx+offset+1, posy/6, '▄', termbox.ColorRed|termbox.AttrBold, termbox.ColorWhite|termbox.AttrBold)
	} else { // Special handling for half-offset. Gives smoother bird motion.
		cb := termbox.CellBuffer()
		tl, bl, tr, br := cb[(posx+offset)+(posy/6*width)], cb[(posx+offset)+(posy/6+1)*width], cb[(posx+offset+1)+(posy/6*width)], cb[(posx+offset+1)+(posy/6+1)*width]
		if wing == '▄' {
			termbox.SetCell(posx+offset, posy/6, '▄', termbox.ColorYellow|termbox.AttrBold, tl.Bg)
			termbox.SetCell(posx+offset, posy/6+1, '▀', termbox.ColorYellow, bl.Bg)
		} else {
			termbox.SetCell(posx+offset, posy/6, '▄', termbox.ColorYellow, tl.Bg)
			termbox.SetCell(posx+offset, posy/6+1, '▀', termbox.ColorYellow|termbox.AttrBold, bl.Bg)
		}
		termbox.SetCell(posx+offset+1, posy/6, '▄', termbox.ColorWhite|termbox.AttrBold, tr.Bg)
		termbox.SetCell(posx+offset+1, posy/6+1, '▀', termbox.ColorRed|termbox.AttrBold, br.Bg)
	}
	
	DrawScore()
	
	if paused {
		DrawText("Paused!", framewidth/2+offset, 24, AlignCenter)
	}
	
	termbox.Flush()
}

func DrawMedal(x, y, medal int) {
	fg, bg := termbox.ColorMagenta|termbox.AttrBold, termbox.ColorMagenta // Default colors, just in case
	name := "????"
	switch medal {
	case 0:
		return // Nothing to draw!
	case BronzeMedal:
		fg, bg = termbox.ColorYellow, termbox.ColorRed
		name = "Bronze"
	case SilverMedal:
		fg, bg = termbox.ColorWhite, termbox.ColorBlack|termbox.AttrBold
		name = "Silver"
	case GoldMedal:
		fg, bg = termbox.ColorYellow|termbox.AttrBold, termbox.ColorYellow
		name = "Gold"
	case PlatinumMedal:
		fg, bg = termbox.ColorWhite|termbox.AttrBold, termbox.ColorWhite
		name = "Platinum"
	}
	l1 := []rune{'▓', '▓', '▓', '▓'}
	l2 := []rune{'▓', '▓', '░', '█', '▓', '▓'}
	l3 := []rune{'▓', '▓', '▒', '░', '▓', '▓'}
	l4 := []rune{'▓', '▓', '▓', '▓'}
	DrawRunesColor(l1, x, y, AlignCenter, fg, bg)
	DrawRunesColor(l2, x, y+1, AlignCenter, fg, bg)
	DrawRunesColor(l3, x, y+2, AlignCenter, fg, bg)
	DrawRunesColor(l4, x, y+3, AlignCenter, fg, bg)
	DrawTextColor(name, x, y+4, AlignCenter, fg, bg)
}

func AwardMedal() {
	switch {
	case score == 10:
		medal = BronzeMedal
	case score == 20:
		medal = SilverMedal
	case score == 30:
		medal = GoldMedal
	case score == 40:
		medal = PlatinumMedal
	case true:
		return // Don't draw if the medal hasn't changed
	}
	DrawCurtains() // Erase old medal by overwriting entire screen. Seems efficient.
	DrawMedal(width-8, 2, medal)
	DrawGuide()
}

func Flap() {
	if !gameover && !paused { 
		vely = 7
		posy = posy/6*6 - 2
	}
}

func InitGame() {
	// Bird start position
	posx, posy = 4, height/2*6
	vely = 0
	
	// Move pipes off screen
	for i, _ := range pipes {
		pipes[i].Xpos = -4
	}
	newpipe = pipespacing
	pipeidx = 0
	
	// Randomize clover locations
	for i, _ := range clovers {
		clovers[i].Xpos = i*6 + 1 - rand.Intn(3)
		clovers[i].Ypos = rand.Intn(4)+height-4
	}

	score = 0
	medal = NoMedal

	wing = '▀'

	running = true
	started = false
	gameover = false
	paused = false

	DrawCurtains()
	DrawGuide()
}

func main() {

	SetupConsole() // Force windows console to 80x40, because I rule with an iron elbow
	
	err := termbox.Init()
	if err != nil {
		return
	}
	defer termbox.Close()

	rand.Seed(time.Now().UnixNano())

	// Cross section of the pipe
	pipecolors = [4]termbox.Cell{
		termbox.Cell{'░', termbox.ColorGreen | termbox.AttrBold, termbox.ColorGreen},
		termbox.Cell{' ', termbox.ColorGreen | termbox.AttrBold, termbox.ColorGreen},
		termbox.Cell{' ', termbox.ColorGreen | termbox.AttrBold, termbox.ColorGreen},
		termbox.Cell{'░', termbox.ColorBlack, termbox.ColorGreen},
	}

	width, height = termbox.Size()

	framewidth = height // Width of actual game window

	offset = width/2 - framewidth/2 // Position of game window
	
	pipespacing = 30	
	pipecount = framewidth / pipespacing
	pipes = make([]Pipe, pipecount+1) // Initiate pipes

	clovers = make([]Clover, 7) // Initiate clovers
	moveclover = 0

	ticker := time.NewTicker(time.Second / 20) // Gotta keep time

	// Handle player input
	events := Keyer()
	go func() {
		for e := range events {
			if e.Type == termbox.EventKey {
				switch e.Ch {
				case 'q': // quit game
					running = false
				case 'r': // reset
					InitGame()
				case 'p':
					paused = !paused
				case 'f': // alternate flap key
					started = true
					Flap()
				}
				switch e.Key {
				case termbox.KeyArrowUp: // flap
					started = true
					Flap()
				}
			}
		}
	}()

	InitGame()
	
	for running {
		if started && !gameover && !paused {
		
			// Update position and velocity
			posy -= vely
			vely -= 1 // G = 1
			
			// Animate wing on flap
			if vely > 1 {
				wing = '▄'
			} else {
				wing = '▀'
			}
			
			// Check if bird touches ground
			if posy/6 > height-6 {
				gameover = true
			}
			
			// Stop bird from exiting top of screen
			if posy < 0 {
				posy = 0
				vely = 0
			}
			
			// Move clovers
			moveclover += 1
			if moveclover == 4 {
				for i, _ := range clovers {
					if clovers[i].Xpos > 0 {
						clovers[i].Xpos += -1
					} else {
						clovers[i].Xpos = framewidth + 1 - rand.Intn(3)
						clovers[i].Ypos = rand.Intn(4)+height-4
					}
				}
				moveclover = 0
			}
			
			// Move pipes and check for collision
			for i, _ := range pipes {
				pipes[i].Xpos -= 1                                 
				if pipes[i].Xpos <= posx+1 && pipes[i].Xpos+4 > posx { 
					if posy >= pipes[i].Gap*6 && posy < (pipes[i].Gap+5)*6 {
						if pipes[i].Xpos == posx {
							score += 1
							AwardMedal()
						}
					} else {
						gameover = true
					}
				}
			}
			
			// Recycle old pipes
			if newpipe == 0 {
				pipes[pipeidx].Xpos = framewidth
				pipes[pipeidx].Gap = rand.Intn(height-24) + 8
				newpipe = pipespacing
				pipeidx += 1
				if pipeidx > pipecount {
					pipeidx = 0
				}
			}

			newpipe -= 1
		}

		Draw()

		<-ticker.C
	}
}
