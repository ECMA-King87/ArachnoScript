package lib

func Red(str string) string {
	return Format("\x1b[31m%s\x1b[0m", str)
}

func Yellow(str string) string {
	return Format("\x1b[33m%s\x1b[0m", str)
}

func Green(str string) string {
	return Format("\x1b[32m%s\x1b[0m", str)
}

func Blue(str string) string {
	return Format("\x1b[34m%s\x1b[0m", str)
}

func Magenta(str string) string {
	return Format("\x1b[35m%s\x1b[0m", str)
}

func Cyan(str string) string {
	return Format("\x1b[36m%s\x1b[0m", str)
}

func White(str string) string {
	// return Format("\x1b[37m%s\x1b[0m", str)
	return Format("\x1b[97m%s\x1b[0m", str)
}

func Black(str string) string {
	return Format("\x1b[30m%s\x1b[0m", str)
}

func Bold(str string) string {
	return Format("\x1b[1m%s\x1b[0m", str)
}

func Dull(str string) string {
	return Format("\x1b[2m%s\x1b[0m", str)
}

func Italic(str string) string {
	return Format("\x1b[3m%s\x1b[0m", str)
}

func Underline(str string) string {
	return Format("\x1b[4m%s\x1b[0m", str)
}

func WhiteBG(str string) string {
	return Format("\x1b[7m%s\x1b[0m", str)
}

func RedBG(str string) string {
	return Format("\x1b[101m%s\x1b[0m", str)
}

func GreenBG(str string) string {
	return Format("\x1b[102m%s\x1b[0m", str)
}

func YellowBG(str string) string {
	return Format("\x1b[103m%s\x1b[0m", str)
}

func CyanBG(str string) string {
	return Format("\x1b[104m%s\x1b[0m", str)
}

func MagentaBG(str string) string {
	return Format("\x1b[105m%s\x1b[0m", str)
}

func Invisible(str string) string {
	return Format("\x1b[8m%s\x1b[0m", str)
}

func StrikeThrough(str string) string {
	return Format("\x1b[9m%s\x1b[0m", str)
}
