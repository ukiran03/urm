// This files deals with moving and writing trash files
package main

type TrashImpl interface {
	TrashDir() (string, error)
	SourceDir() string

	// [05-03-2026] TODO: this should hold logic for each trash implementation
	//   - if it is HomeTrash, hopefully os.Rename will be enough
	//   - if it is SpecialTrash, copy+delete should be implemented
	MoveToTrash() error
}

func (ht *HomeTrash) SourceDir() string {
	return ht.homeDir
}

func (st *SpecialTrash) SourceDir() string {
	return st.rootDir
}

func Rename(target string, trash TrashImpl) error {
	return nil
}
