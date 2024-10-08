package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"rggo/interacting/todo"
	"runtime"
	"testing"
	"time"
)

var binName = "todo"

func TestMain(m *testing.M) {
	fmt.Println("Building tool...")

	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	build := exec.Command("go", "build", "-o", binName)

	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot build tool %s: %s", binName, err)
		os.Exit(1)
	}
	fmt.Println("Running tests...")
	result := m.Run()

	fmt.Println("Cleaning up...")
	os.Remove(binName)

	os.Exit(result)
}

func TestTodoCLI(t *testing.T) {
	task := "test task number 1"

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tf, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("error creating temp file: %s", err)
	}
	defer os.Remove(tf.Name())

	cmdPath := filepath.Join(dir, binName)

	t.Run("AddNewTaskFromArgs", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add", task)
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	task2 := "test task number 2"
	t.Run("AddNewTaskFromSTDIN", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		cmdStdIn, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(cmdStdIn, task2)
		cmdStdIn.Close()

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ListTasks", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf("  1: %s\n  2: %s\n", task, task2)

		if expected != string(out) {
			t.Errorf("expected %q, got %q\n", expected, string(out))
		}
	})

	t.Run("ListTasksVerbose", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list", "-v")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		l := todo.List{}
		l.Get(tf.Name())

		expected := fmt.Sprintf("  1: %s [created: %s]\n  2: %s [created: %s]\n",
			task, l[0].CreatedAt.Format(time.DateTime),
			task2, l[1].CreatedAt.Format(time.DateTime))

		if expected != string(out) {
			t.Errorf("expected %q, got %q\n", expected, string(out))
		}
	})

	t.Run("CompleteTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-complete", "1")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command(cmdPath, "-list")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf("X 1: %s\n  2: %s\n", task, task2)

		if expected != string(out) {
			t.Errorf("expected %q, got %q\n", expected, string(out))
		}
	})

	t.Run("ListTasksSuppressCompleted", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list", "-s")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf("  1: %s\n", task2)

		if expected != string(out) {
			t.Errorf("expected %q, got %q\n", expected, string(out))
		}
	})

	t.Run("DeleteTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-del", "1")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command(cmdPath, "-list")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TODO_FILENAME=%s", tf.Name()))

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf("  1: %s\n", task2)

		if expected != string(out) {
			t.Errorf("expected %q, got %q\n", expected, string(out))
		}
	})
}
