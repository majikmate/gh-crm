package crm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type assignment struct {
	Id   int
	Name string
}

func IsAssignmentFolder() (bool, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get current directory: %v", err)
	}

	p := filepath.Join(currentDir, crmFolder, assigmentFile)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

func NewAssignment() *assignment {
	return &assignment{}
}

func LoadAssignment() (*assignment, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}

	p := filepath.Join(currentDir, crmFolder, assigmentFile)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil, fmt.Errorf("No assigment found. Run `gh crm clone` to clone an assignment or chnage to a folder that contains an assignment.")
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s file: %v", p, err)
	}
	defer f.Close()

	j, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %v", p, err)
	}

	a := NewAssignment()
	err = json.Unmarshal(j, &a)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s file: %v", p, err)
	}

	return a, nil
}

func (a *assignment) Set(id int, name string) {
	a.Id = id
	a.Name = name
}

func (a *assignment) Save(path string) error {
	var err error
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	f := filepath.Join(path, crmFolder)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		err := os.Mkdir(f, 0755)
		if err != nil {
			return fmt.Errorf("failed to create %s directory: %v", f, err)
		}
	}

	j, err := json.MarshalIndent(a, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal classroom: %v", err)
	}

	p := filepath.Join(f, assigmentFile)
	file, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("failed to create %s file: %v", p, err)
	}
	defer file.Close()

	_, err = file.Write(j)
	if err != nil {
		return fmt.Errorf("failed to write %s file: %v", p, err)
	}

	return nil
}
