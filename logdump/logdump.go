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

/*
func main() {
	Print("Hello,World !")
}
*/