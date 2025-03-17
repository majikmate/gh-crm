package crm

import (
	"fmt"
	"os"
)

const (
	crmFolder = ".crm"

	classroomFile = "classroom.json"
	assigmentFile = "assignment.json"
)

func Fatal(v ...any) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
