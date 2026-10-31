package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func CloneRepository(ctx context.Context, cloneURL, path string) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	gitExe, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed: %w", err)
	}

	_, err = os.Stat(path)

	switch {
	case err == nil:
		return fmt.Errorf("destination already exist: %s", path)
	case os.IsNotExist(err):
		//valid
	default:
		return err
	}

	cmd := exec.CommandContext(ctx, gitExe, "clone", cloneURL, path)

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

func CheckoutRepository(ctx context.Context, path, branch, commitSHA string) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	gitExe, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed: %w", err)
	}

	file, err := os.Stat(path)

	switch {
	case os.IsNotExist(err):
		return fmt.Errorf("workspace does not exist: %w", err)
	case err != nil:
		return fmt.Errorf("failed to check workspace: %w", err)
	case !file.IsDir():
		return fmt.Errorf("workspace path is not a directory: %s", path)
	}

	var cmd *exec.Cmd

	if commitSHA != "" {
		if err := validateCommitSHA(commitSHA); err != nil {
			return err
		}
		cmd = exec.CommandContext(ctx, gitExe, "-C", path, "checkout", "--detach", commitSHA)
	} else {
		if err := validateBranch(ctx, gitExe, branch); err != nil {
			return err
		}
		cmd = exec.CommandContext(ctx, gitExe, "-C", path, "checkout", branch)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %s: %w", output, err)
	}

	return nil
}

func validateCommitSHA(commitSHA string) error {
	if len(commitSHA) < 7 || len(commitSHA) > 64 {
		return fmt.Errorf("invalid commit SHA length")
	}

	for _, char := range commitSHA {
		if !((char >= '0' && char <= '9') ||
			(char >= 'a' && char <= 'f') ||
			(char >= 'A' && char <= 'F')) {
			return fmt.Errorf("commit SHA must contain only hexadecimal characters")
		}
	}

	return nil
}

func validateBranch(ctx context.Context, gitExe, branch string) error {
	if branch == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	cmd := exec.CommandContext(ctx, gitExe, "check-ref-format", "--branch", branch)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("invalid branch name %q: %s", branch, output)
	}

	return nil
}
