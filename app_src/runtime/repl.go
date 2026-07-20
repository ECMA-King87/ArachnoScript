package runtime

import "aspire/are/main/lib"

// The REPL is inefficient and young.
func REPL() {
	lib.Print(lib.Version())
	r := NewRuntime(true)
	session := lib.NewStringBuilder()
	d_pressed := false
	for {
		input := lib.Prompt(lib.Magenta(">> "))
		if input == "" {
			continue
		} else if input[0] == 4 { // ctrl+d ^D
			if d_pressed {
				break
			}
			d_pressed = true
			lib.Println("pressed Ctrl+d again to exit.")
			continue
		} else if input[0] == 0 || input == ".exit" {
			break
		} else if input == ".save" {
			p, _ := lib.UserHomeDir()
			path := p + "\\Documents\\are-session-" + lib.Sprint(lib.TimeNow().UnixMilli()) + ".txt"
			err := lib.WriteTextFile(path, session.String())
			if err != nil {
				lib.Println(err)
			} else {
				lib.Println("File Saved to " + path + "'.")
			}
			continue
		}
		w := r.Worker(input, false)
		w.Run()
		session.WriteString(input)
		d_pressed = false
	}
}
