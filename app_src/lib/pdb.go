package lib

import (
// "os"
)

type PDB struct {
	// file    *os.File
	entries  map[int]string
	keyCount int
}

func NewPDB(name string) *PDB {
	v, e := GetEnvVar(ARE_CACHE_DIR)
	if !e {
		v = JoinPaths(DirOf(ExecPath()), "cache")
		SetEnvVar(ARE_CACHE_DIR, v)
	}
	// file, err := os.Create(JoinPaths(v, name))
	// if err != nil {
	// 	Println("Error creating pdb.", err)
	// }
	return &PDB{
		// file,
		map[int]string{}, 0}
}

//
// func (pdb *PDB) Write(key, path int, loc Loc) (n int) {
// 	var real_path string = AnonymousPath
// 	func() {
// 		path2 := pathFromKey(path)
// 		real_path = path2
// 		defer recover()
// 	}()
// 	s := SourceAtPosition(real_path, loc)
// 	n, err := pdb.file.WriteString(Sprintf("\x00%d%d%s", key, len(s), s))
// 	if err != nil && DEBUG_MODE {
// 		Panic(err)
// 	}
// 	return
// }

func (pdb *PDB) Write(path int, loc Loc) (key int) {
	key = pdb.keyCount
	pdb.keyCount++
	defer func() {
		recover()
		pdb.entries[key] = SourceLog(-1, loc)
	}()
	pdb.entries[key] = SourceLog(path, loc)
	return
}

func (pdb *PDB) Read(key int) string {
	return pdb.entries[key]
}
