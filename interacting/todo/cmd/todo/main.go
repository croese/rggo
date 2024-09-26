package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"rggo/interacting/todo"
	"strings"
	"time"
)

var todoFileName = ".todo.json"

type listStringer struct {
	verbose       bool
	skipCompleted bool
	list          *todo.List
}

func (s listStringer) String() string {
	formatted := ""

	// handling counter manually since we might skip items in the loop
	index := 0
	for _, t := range *s.list {
		prefix := "  "
		if t.Done && s.skipCompleted {
			continue
		}

		if t.Done {
			prefix = "X "
		}

		if s.verbose {
			formatted += fmt.Sprintf("%s%d: %s [created: %s]\n", prefix,
				index+1, t.Task, t.CreatedAt.Format(time.DateTime))
		} else {
			formatted += fmt.Sprintf("%s%d: %s\n", prefix, index+1, t.Task)
		}

		index++
	}
	return formatted
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"%s tool. Developed by croese\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Copyright 2024\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage information:\n")
		flag.PrintDefaults()
	}
	add := flag.Bool("add", false, "Task to be included in the todo list. Task description may be provided after this argument or from STDIN.")
	list := flag.Bool("list", false, "List all tasks")
	complete := flag.Int("complete", 0, "Item to be completed")
	delete := flag.Int("del", 0, "Item to be deleted")
	verbose := flag.Bool("v", false, "Enable verbose output")
	showCompleted := flag.Bool("s", false, "Suppress completed tasks")
	flag.Parse()

	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}

	l := &todo.List{}

	if err := l.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case *list:
		s := listStringer{*verbose, *showCompleted, l}
		fmt.Print(s)
	case *complete > 0:
		if err := l.Complete(*complete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *add:
		t, err := getTask(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		l.Add(t)

		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *delete > 0:
		if err := l.Delete(*delete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid option")
		os.Exit(1)
	}
}

func getTask(r io.Reader, args ...string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	s := bufio.NewScanner(r)
	s.Scan()
	if err := s.Err(); err != nil {
		return "", err
	}
	if len(s.Text()) == 0 {
		return "", fmt.Errorf("Task cannot be blank")
	}
	return s.Text(), nil
}
