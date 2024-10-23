package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
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

func main() {
	clear()
	fmt.Println("Minesweeper written in Go - Q to quit")
	fmt.Println("Choose game type:")

	selectedDifficulty = 0

	PrintDifficulties()
	PrintSelector(difficulties, 1, selectedDifficulty)

	difficulty := GetDifficulty()

	width, height, mineNum := uint(0), uint(0), uint(0)

	switch difficulty {
	case "Easy":
		break
	case "Medium":
		break
	case "Hard":
		break
	case "Expert":
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
	PrintMinefield(CreateMinefield(width, height, mineNum))

}

func PrintMinefield(field minefield) {
	mineStrings := make([]string, field.height, field.height)
	for _, v := range field.hasMine {
		for y, b := range v {
			if b {
				mineStrings[y] += "."
			} else {
				mineStrings[y] += "#"
			}
		}
	}
	for _, v := range mineStrings {
		fmt.Println(v)
	}
}

type minefield struct {
	width   uint
	height  uint
	mineNum uint
	hasMine [][]bool
}

func CreateMinefield(width uint, height uint, mineNum uint) minefield {
	hasMine := make([][]bool, width, width)
	for i := range hasMine {
		hasMine[i] = make([]bool, height, height)
	}

	field := minefield{width, height, mineNum, hasMine}
	for x := uint(0); x < width; x++ {
		for y := uint(0); y < height; y++ {
			hasMine[x][y] = false
		}
	}

	for i := uint(0); i < mineNum; i++ {
		for {
			if hasMine[rng.Uint()%width][rng.Uint()%height] != true {
				hasMine[rng.Uint()%width][rng.Uint()%height] = true
				break
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
		char, key, err := keyboard.GetSingleKey()
		if err != nil {
			time.Sleep(time.Millisecond * 30)
			continue
		}

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
	return ConsoleGetUint("Choose Width:", 0, 100), ConsoleGetUint("Choose Height:", 0, 100), ConsoleGetUint("Choose Number of Mines:", 0, 100)
}

func GetDifficulty() string {
	for {
		char, key, err := keyboard.GetSingleKey()
		if err != nil {
			time.Sleep(time.Millisecond * 30)
			continue
		}

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
