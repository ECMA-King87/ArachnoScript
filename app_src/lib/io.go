package lib

import (
	// "bufio"
	"fmt"
	// "os"
	// "strings"
)

func catch[T any](v T, err error) T {
	if err != nil {
		Panic(Sprintf("Error: %s", err))
	}
	return v
}

func Panic(v any) {
	if DEBUG_MODE {
		panic(v)
	} else {
		Println(v)
		ExitWith1()
	}
}

func Print(data ...any) {
	catch(fmt.Print(data...))
}

func Sprint(data ...any) string {
	return fmt.Sprint(data...)
}

func Println(data ...any) {
	catch(fmt.Println(data...))
}

func Sprintln(data ...any) string {
	return fmt.Sprintln(data...)
}

func Printf(f string, data ...any) {
	catch(fmt.Printf(f, data...))
}

func Sprintf(f string, data ...any) string {
	return fmt.Sprintf(f, data...)
}

func Format(f string, data ...any) string {
	return fmt.Sprintf(f, data...)
}

func Errorf(f string, data ...any) error {
	return fmt.Errorf(f, data...)
}

func Prompt(msg string) string {
	Print(msg)
	// reader := bufio.NewReader(os.Stdin)
	// input, err := reader.ReadString('\n')
	// if err != nil {
	// 	return string([]byte{0})
	// }
	// // trim trailing newline/carriage return
	// input = strings.TrimRight(input, "\r\n")
	input := ""
	fmt.Scanln(&input)
	return input
}
