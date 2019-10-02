package logdump
//package main

import(
	"fmt"
	"github.com/fatih/color"
	"time"
)

// Println is a function used to log with timestamp
func Println(msg string) {
	color.Set(color.FgGreen)
	fmt.Printf("INFO ")
	color.Unset()
	color.Set(color.FgMagenta)
	fmt.Printf(time.Now().Format("15:04:05")) 
	color.Unset()
	fmt.Println("",msg)
}

// DefaultPrint is a wrapeer around Println in fmt package
func DefaultPrintBlue(a ...interface{}) {
	color.Set(color.FgBlue)
	fmt.Println(a...)
	color.Unset()
}

func DefaultPrint(a ...interface{}) {
	fmt.Println(a...)
}

func Printtx(msg string) {
	color.Set(color.FgBlue)
	fmt.Printf(msg)
	color.Unset()
}


// func main() {
// 	DefaultPrint("Hello,World !")
// }
