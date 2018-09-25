package golang

import (
	"errors"
	"os"
	"os/exec"
)

// Compile the source code
func Compile(src string, dst string, env map[string]string) error {
	// lookup gobin
	gobin, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	// executable
	cmd := exec.Command(gobin, "build", "-o", dst, src)
	for key, val := range env {
		cmd.Env = append(cmd.Env, key+"="+val)
	}

	// execute and return the output
	b, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(b))
	}

	return nil
}

// cleanup our function
func rmrf(filename string) error {
	if err := os.RemoveAll(filename); err != nil {
		return err
	}
	return nil
}