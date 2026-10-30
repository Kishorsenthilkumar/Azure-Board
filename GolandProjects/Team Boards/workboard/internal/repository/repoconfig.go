package repository

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func CloneRepository(cloneURL, path string) error {

	gitExe, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed: %w", err)
	}

	_, err = os.Stat(path)

	switch {
	case err == nil:
		return fmt.Errorf("destination already exist: %s", err)
	case os.IsNotExist(err):
		//valid
	default:
		return err
	}

	cmd := exec.Command(gitExe, "clone", cloneURL, path)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if cleanupErr := os.RemoveAll(path); cleanupErr != nil {
			return errors.Join(
				fmt.Errorf("git clone failed: %s: %w", output, err),
				fmt.Errorf("cleanup failed: %w", cleanupErr),
			)
		}
		return fmt.Errorf("git clone failed: %s: %w", output, err)
	}
	return nil

}

func CheckoutRepository(path, branch, commitSHA string) error {
	gitExe, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed: %w", err)
	}

	file, err := os.Stat(path)

	switch {
	case os.IsNotExist(err):
		return fmt.Errorf("workspace does not exist: %w", err)
	case err == nil && !file.IsDir():
		return fmt.Errorf("Path exists but isn't a directory: %w", err)
	default:
		return err
	}

	cmd := exec.Command(gitExe, "-C", path, "checkout", branch)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %s: %w", output, err)
	}

	return nil
}
