package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eiannone/keyboard"
)

var selectedDifficulty uint

var difficulties = []string{"Easy", "Medium", "Hard", "Expert", "Custom"}

var difficultyNum = len(difficulties)

var leftKeys = []keyboard.Key{keyboard.KeyArrowLeft, keyboard.KeyArrowUp}
var leftRunes = []rune{'h', 'j'}
var rightKeys = []keyboard.Key{keyboard.KeyArrowRight, keyboard.KeyArrowDown}
var rightRunes = []rune{'k', 'l'}
var actionRunes = []rune{'h', 'j', 'k', 'l', 'c', 'm'}
var actionKeys = []keyboard.Key{keyboard.KeyArrowLeft, keyboard.KeyArrowRight, keyboard.KeyArrowUp, keyboard.KeyArrowDown, keyboard.KeySpace, keyboard.KeyEnter}

var d = rand.NewPCG(uint64(time.Now().Unix()), uint64(time.Now().UnixMicro()))
var rng = rand.New(d)

func clear() {
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("clear")
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func PrintDifficulties() {
	for i, diff := range difficulties {

		if i != difficultyNum-1 {
			fmt.Print(diff + " ")
		} else {
			fmt.Print(diff + "\n")
		}

	}
}

func CheckQuit(char rune, key keyboard.Key) {
	if key == keyboard.KeyEsc || key == keyboard.KeyCtrlC || char == 'q' {
		keyboard.Close()
		os.Exit(0)
	}
}

var (
	keyBuffMut sync.Mutex
	keyBuffUse bool
)

type cursor_mode int

const (
	CLEAR = iota
	MARK
)

type currKey struct {
	keyVal keystroke
	fired  bool
}

type position struct {
	x int
	y int
}

type cursor struct {
	pos  position
	mode cursor_mode
}

type keystroke struct {
	char rune
	key  keyboard.Key
}

var offsets [][]int = [][]int{
	{0, 1},
	{1, 0},
	{0, -1},
	{-1, 0},
	{-1, -1},
	{1, 1},
	{1, -1},
	{-1, 1}}

func AddKeystroke(buff *currKey) {
	char, key, _ := keyboard.GetKey()
	keyBuffMut.Lock()
	if buff.fired {
		buff.keyVal = keystroke{char, key}
		buff.fired = false
	}
	keyBuffMut.Unlock()
}

func main() {
	clear()
	fmt.Println("Minesweeper written in Go - Q to quit")
	fmt.Println("Choose game type:")

	selectedDifficulty = 0

	keyboard.Open()

	PrintDifficulties()
	PrintSelector(difficulties, 1, selectedDifficulty)

	difficulty := GetDifficulty()

	width, height, mineNum := uint(0), uint(0), uint(0)

	switch difficulty {
	case "Easy":
		width = 10
		height = 10
		mineNum = 20
		break
	case "Medium":
		width = 20
		height = 14
		mineNum = 65
		break
	case "Hard":
		width = 32
		height = 15
		mineNum = 120
		break
	case "Expert":
		width = 40
		height = 15
		mineNum = 170
		break
	case "Custom":
		for {
			width, height, mineNum = GetCustomDifficulty()
			if mineNum >= width*height {
				clear()
				fmt.Println("Too many mines! :(")
				time.Sleep(1000 * time.Millisecond)
			} else {
				break
			}
		}
		break
	}

	clear()
	field := CreateMinefield(width, height, mineNum)
	cur := cursor{position{int(width / 2), int(height / 2)}, CLEAR}
	frameNum := uint(0)

	var repeatStack string = ""

	var startTime time.Time = time.Now()
	var endTime time.Time = time.Now()

	var keyBuffer = new(currKey)
	(*keyBuffer).fired = true

	gameEnd := false
	endMessage := ""

	go func(*currKey) {
		for !gameEnd {
			AddKeystroke(keyBuffer)
			time.Sleep(time.Millisecond * 1)
		}
	}(keyBuffer)

	cursorDrawn := true

	for !gameEnd {
		startTime = time.Now()

		update := false
		keystroke := keystroke{}
		keyBuffMut.Lock()
		if !keyBuffer.fired {
			keystroke = keyBuffer.keyVal
			keyBuffer.fired = true
			update = true
		}
		keyBuffMut.Unlock()

		if update {
			char := keystroke.char
			key := keystroke.key

			if repeatStack != "" && keystroke.key == keyboard.KeyEsc {
				repeatStack = ""
			} else {
				CheckQuit(char, key)
			}

			if char >= '0' && char <= '9' {
				repeatStack += string(char)
			}

			numRep := int(1)
			if repeatStack != "" {
				numTmp, err := strconv.Atoi(repeatStack)
				if err == nil {
					numRep = max(numTmp, 1)
				}
			}

			if slices.Contains(actionKeys, key) {
				switch key {
				case keyboard.KeyArrowLeft:
					cur.pos.x = max(cur.pos.x-numRep, 0)
					break
				case keyboard.KeyArrowRight:
					cur.pos.x = min(cur.pos.x+numRep, int(width-1))
					break
				case keyboard.KeyArrowDown:
					cur.pos.y = min(cur.pos.y+numRep, int(height-1))
					break
				case keyboard.KeyArrowUp:
					cur.pos.y = max(cur.pos.y-numRep, 0)
					break
				case keyboard.KeySpace, keyboard.KeyEnter:
					if cur.mode == CLEAR {
						field.area[cur.pos.x][cur.pos.y].isCleared = true
						if field.area[cur.pos.x][cur.pos.y].hasMine {
							gameEnd = true
							endMessage = "You lose! :("
						} else if field.area[cur.pos.x][cur.pos.y].mineNum == 0 {
							pos := make([]position, 1, 1)
							pos[0] = cur.pos
							ClearBlank(&field, pos, make([]position, 0, 10))
						}
					} else if cur.mode == MARK {
						field.area[cur.pos.x][cur.pos.y].hasMark = !field.area[cur.pos.x][cur.pos.y].hasMark
					}
				}
			}

			if slices.Contains(actionRunes, char) {
				switch char {
				case 'h':
					cur.pos.x = max(cur.pos.x-numRep, 0)
					break
				case 'l':
					cur.pos.x = min(cur.pos.x+numRep, int(width-1))
					break
				case 'j':
					cur.pos.y = min(cur.pos.y+numRep, int(height-1))
					break
				case 'k':
					cur.pos.y = max(cur.pos.y-numRep, 0)
					break
				case 'c':
					cur.mode = CLEAR
					field.area[cur.pos.x][cur.pos.y].isCleared = true
					if field.area[cur.pos.x][cur.pos.y].hasMine {
						gameEnd = true
						endMessage = "You lose! :("
					} else if field.area[cur.pos.x][cur.pos.y].mineNum == 0 {
						pos := make([]position, 1, 1)
						pos[0] = cur.pos
						ClearBlank(&field, pos, make([]position, 0, 10))
					}
				case 'm':
					cur.mode = MARK
					field.area[cur.pos.x][cur.pos.y].hasMark = !field.area[cur.pos.x][cur.pos.y].hasMark
				}
			}

			if slices.Contains(actionKeys, key) || slices.Contains(actionRunes, char) {
				repeatStack = ""
			}
		}
		endTime = time.Now()
		time.Sleep(startTime.Sub(endTime) + time.Millisecond*20)
		if frameNum%20 == 0 {
			cursorDrawn = !cursorDrawn
		}

		if update || frameNum%20 == 0 {
			clear()
			PrintMinefield(field, cur, cursorDrawn)
			if cur.mode == CLEAR {
				fmt.Print("CLEAR MODE: (c)")
			} else {
				fmt.Print("MARK MODE: (m)")
			}
			if repeatStack != "" {
				fmt.Print(" - ", repeatStack)
			}
			fmt.Print("\n")
		}
		frameNum++
	}

	clear()
	for x := range field.area {
		for y := range field.area[x] {
			if field.area[x][y].hasMine {
				field.area[x][y].isCleared = true
			}
		}
	}
	PrintMinefield(field, cur, false)
	fmt.Println(endMessage)

}

func ClearBlank(field *minefield, pos []position, finPos []position) {
	newPos := make([]position, 0, 8*len(pos))
	for _, v := range pos {
		if field.area[v.x][v.y].mineNum == 0 && !slices.Contains(finPos, v) {
			field.area[v.x][v.y].isCleared = true
			for _, off := range offsets {
				if v.x+off[0] >= 0 && v.x+off[0] < int(field.width) && v.y+off[1] >= 0 && v.y+off[1] < int(field.height) {
					field.area[v.x+off[0]][v.y+off[1]].isCleared = true
					newPos = append(newPos, position{v.x + off[0], v.y + off[1]})
				}
			}
		}
		finPos = append(finPos, v)
	}

	if len(newPos) > 0 {
		ClearBlank(field, newPos, finPos)
	}
}

func GetPlotChar(plot plot) string {
	var char string = ""

	if plot.isCleared {
		if plot.hasMine {
			char = "*"
		} else if plot.mineNum > 0 {
			char = strconv.Itoa(int(plot.mineNum))
		} else {
			char = " "
		}
	} else if plot.hasMark {
		char = "!"
	} else {
		char = "#"
	}

	return char
}

func PrintMinefield(field minefield, cur cursor, cursorDrawn bool) {
	mineStrings := make([]string, field.height, field.height)
	for x, v := range field.area {
		for y, plot := range v {
			if cur.pos.x == x && cur.pos.y == y {
				if cursorDrawn {
					if cur.mode == CLEAR {
						mineStrings[y] += "c"
					} else {
						mineStrings[y] += "m"
					}
				} else {
					mineStrings[y] += GetPlotChar(plot)
				}
			} else {
				mineStrings[y] += GetPlotChar(plot)
			}
		}
	}
	for _, v := range mineStrings {
		fmt.Println(v)
	}
}

type plot struct {
	hasMine   bool
	isCleared bool
	hasMark   bool
	mineNum   uint
}

type minefield struct {
	width   uint
	height  uint
	mineNum uint
	area    [][]plot
}

func CreateMinefield(width uint, height uint, mineNum uint) minefield {
	area := make([][]plot, width, width)
	for i := range area {
		area[i] = make([]plot, height, height)
	}

	field := minefield{width, height, mineNum, area}
	for x := uint(0); x < width; x++ {
		for y := uint(0); y < height; y++ {
			field.area[x][y].hasMine = false
			field.area[x][y].hasMark = false
			field.area[x][y].isCleared = false
			field.area[x][y].mineNum = 0
		}
	}

	// Randomly place mines
	for i := uint(0); i < mineNum; i++ {
		for {
			if !field.area[rng.Uint()%width][rng.Uint()%height].hasMine {
				field.area[rng.Uint()%width][rng.Uint()%height].hasMine = true
				break
			}
		}
	}

	// Count adjacent mines for each plot
	for x := uint(0); x < width; x++ {
		for y := uint(0); y < height; y++ {
			for _, off := range offsets {
				xTotal := int(x) + off[0]
				yTotal := int(y) + off[1]
				if !(xTotal < 0 || xTotal >= int(width) || yTotal < 0 || yTotal >= int(height)) {
					if field.area[xTotal][yTotal].hasMine {
						field.area[x][y].mineNum++
					}
				}
			}
		}
	}

	return field
}

func PrintSelector(options []string, spacing int, selectedId uint) {
	var retString string = ""
	var repeatString string
	for i, v := range options {
		if uint(i) == selectedId {
			repeatString = "#"
		} else {
			repeatString = "_"
		}

		var wordSpace = spacing
		if int(math.Floor(float64(len(v))/float64(2))) <= spacing {
			wordSpace = 0
		}

		retString += strings.Repeat(" ", wordSpace) + strings.Repeat(repeatString, len(v)-wordSpace*2) + strings.Repeat(" ", wordSpace) + " "

	}

	fmt.Println(retString)
}

func ConsoleGetUint(message string, min uint, max uint) uint {
	var num uint = 0

	clear()
	fmt.Println(message)
	for {
		char, key, _ := keyboard.GetKey()

		CheckQuit(char, key)

		if key == keyboard.KeyEnter && num > min && num < max {
			break
		} else if char >= '0' && char <= '9' {
			num = num*10 + uint(char-'0')
		} else if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
			num = uint(math.Floor(float64(num) / float64(10)))
		}

		clear()
		fmt.Println(message)
		if num != 0 {
			fmt.Print(num)
		}

	}

	return num
}

func GetCustomDifficulty() (uint, uint, uint) {
	return ConsoleGetUint("Choose Width:", 0, 100), ConsoleGetUint("Choose Height:", 0, 100), ConsoleGetUint("Choose Number of Mines:", 0, 1000)
}

func GetDifficulty() string {
	for {
		char, key, _ := keyboard.GetKey()

		CheckQuit(char, key)

		if key == keyboard.KeyEnter {
			break
		}
		var prevDiff = selectedDifficulty
		if slices.Contains(leftKeys, key) || slices.Contains(leftRunes, char) {
			if selectedDifficulty == 0 {
				selectedDifficulty = uint(difficultyNum - 1)
			} else {
				selectedDifficulty = uint((int(selectedDifficulty) - 1) % difficultyNum)
			}
		} else if slices.Contains(rightKeys, key) || slices.Contains(rightRunes, char) {
			selectedDifficulty = uint((int(selectedDifficulty+1) % difficultyNum))
		}
		if prevDiff != selectedDifficulty {
			clear()
			fmt.Println("Minesweeper written in Go - Q to quit")
			fmt.Println("Choose game type:")
			PrintDifficulties()
			PrintSelector(difficulties, 1, selectedDifficulty)
		}
	}
	return difficulties[selectedDifficulty]
}
