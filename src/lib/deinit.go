package lib

func Deinit() {
	defer wg.Wait()
}