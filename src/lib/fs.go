package lib

import (
	"encoding/gob"
	"io"
	"os"
)

type File = *os.File
type FileInfo = os.FileInfo
type FileMode = os.FileMode

var (
	Stdin  = os.Stdin
	Stdout = os.Stdout
	Stderr = os.Stderr
)

const AnonymousPath = "(anonymous)"

type CacheFile struct {
	b       []byte
	modTime int64
}

var paths = NewMap[string, int]()
var cache = NewMap[int, CacheFile]()
var unchanged = NewMap[int, []byte]()

func UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// Returns the associated key of the cached path.
// Returns 0 if there is no such cached path.
func KeyOfPath(path string) int {
	if k, ok := paths.Get(path); ok {
		return k
	}
	return 0
}

// Returns the path of the cached file with the specified key.
// If there is no such key, AnonymousPath is returned.
func PathFromKey(key int) string {
	path := ""
	found := false
	paths.Until(func(p string, value, _ int) bool {
		found = value == key
		if found {
			path = p
		}
		return found
	})
	if found {
		return path
	}
	// if DEBUG_MODE && key != 0 {
	// 	Panic("could not find path with key: " + Sprint(key))
	// }
	return AnonymousPath
}

// Must be positive
var lastKey int

// The files is read once and not revalidated.
// Used in parser and source debugging.
func ReadUnchangedFile(path string) ([]byte, int, error) {
	if !IsAbs(path) {
		path = Abs(path)
	}
	if k, ok := paths.Get(path); ok {
		if cached, ok := cache.Get(k); ok {
			return cached.b, k, nil
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}

	size := info.Size()
	bytes := make([]byte, size)
	if size > 0 {
		n, err := io.ReadFull(f, bytes)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, 0, err
		}
		bytes = bytes[:n]
	}

	lastKey++
	paths.Set(path, lastKey)
	cache.Set(lastKey, CacheFile{
		b:       bytes,
		modTime: getModTime(info),
	})
	return bytes, lastKey, nil
}

func getModTime(info os.FileInfo) int64 {
	return info.ModTime().UnixNano()
}

var AnonymousCount = 0

// Writes to cache the contents of bytes with the negative integer key as the cache key.
// key: Must be negative, Anonymous files use negative keys.
func WriteAnonymousFile(bytes []byte, key int) int {
	if key > 0 {
		Println("Bug: anonymous path key is not negative.")
	}
	if key == 0 {
		AnonymousCount--
		key = AnonymousCount
	}
	cache.Set(key, CacheFile{b: bytes, modTime: 0})
	return key
}

func WriteCacheFile(bytes []byte, path string) (key int) {
	key = lastKey
	paths.Set(path, lastKey)
	cache.Set(lastKey, CacheFile{b: bytes, modTime: 0})
	lastKey++
	return
}

// ReadFile reads the cached file with the specified key and returns the contents.
func ReadCachedFile(key int) []byte {
	if b, ok := cache.Get(key); ok {
		return b.b
	}
	return []byte{}
}

// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF as an error to be reported.
func ReadFile(path string) ([]byte, int, error) {
	if !IsAbs(path) {
		path = Abs(path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}

	k := paths.GetS(path)
	if content, ok := cache.Get(k); ok {
		if getModTime(info) == content.modTime {
			return content.b, k, nil
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	lastKey++
	k = lastKey
	paths.Set(path, k)
	cache.Set(k, CacheFile{b: content, modTime: getModTime(info)})
	return content, k, nil
}

// Possible errors:
// - PathError
// - EOF
func ReadTextFile(path string) (string, error) {
	bytes, _, err := ReadFile(path)
	return string(bytes), err
}

// WriteFile writes data to the named file, creating it if necessary.
// If the file does not exist, WriteFile creates it with permissions perm (before umask); otherwise WriteFile truncates it before writing, without changing permissions.
// Since WriteFile requires multiple system calls to complete, a failure mid-operation can leave the file in a partially written state.
func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o667)
}

// WriteFile writes data to the named file, creating it if necessary.
// If the file does not exist, WriteFile creates it with permissions perm (before umask); otherwise WriteFile truncates it before writing, without changing permissions.
// Since WriteFile requires multiple system calls to complete, a failure mid-operation can leave the file in a partially written state.
func WriteTextFile(path, data string) error {
	return WriteFile(path, []byte(data))
}

// Possible errors:
// - PathError
func OpenFile(path string) (File, error) {
	return os.Open(path)
}

// Possible errors:
// - PathError
func OpenFileWithFlags(path string, flag int) (File, error) {
	return os.OpenFile(path, flag, 0o644)
}

// Create creates or truncates the named file.
// If the file already exists, it is truncated.
// If the file does not exist, it is created with mode 0o666 (before umask).
// If successful, methods on the returned File can be used for I/O; the associated file descriptor has mode O_RDWR.
// The directory containing the file must already exist.
// If there is an error, it will be of type - PathError
func CreateFile(path string) (File, error) {
	return os.Create(path)
}

// FsStat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError and FsStat returns nil.
func FsStat(path string) FileInfo {
	return catch(os.Stat(path))
}

// GobEncode transmits the data item represented by the empty interface value, guaranteeing that all necessary type information has been transmitted first.
// Passing a nil pointer to Encoder could panic, as they cannot be transmitted by gob.
func GobEncode(w io.Writer, value any) error {
	return gob.NewEncoder(w).Encode(value)
}

// GobDecode reads the next value from the input stream and stores it in the data represented by the empty interface value.
// If e is nil, the value will be discarded. Otherwise, the value underlying e must be a pointer to the correct type for the next data item received.
// If the input is at EOF, Decode returns io.EOF and does not modify e.
func GobDecode(r io.Reader, e any) error {
	return gob.NewDecoder(r).Decode(&e)
}

// PathExists Reports whether path is a valid path to a local file or directory.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// OpenFileInDir opens the file name in the directory dir. It is equivalent to OpenRoot(dir) followed by opening the file in the root.
// OpenFileInDir returns nil if any component of the name references a location outside of dir.
func OpenFileInDir(dir, name string) File {
	return catch(os.OpenInRoot(dir, name))
}

// OpenRoot opens the named directory.
// It follows symbolic links in the directory name.
// If there is an error, it will be of type PathError and OpenRoot returns nil.
func OpenRoot(name string) *os.Root {
	return catch(os.OpenRoot(name))
}

func MkdirAll(path string, perm FileMode) error {
	return os.MkdirAll(path, perm)
}
