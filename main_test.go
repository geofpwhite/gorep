package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func TestInvalidArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldPanic bool
	}{
		{
			name:        "NoArguments",
			args:        []string{"gorep"},
			shouldPanic: true,
		},
		{
			name:        "InvalidRegex",
			args:        []string{"gorep", "[invalid", "valid", "test.txt"},
			shouldPanic: true,
		},
		{
			name:        "InvalidFlag",
			args:        []string{"gorep", "-invalid", "test", "replacement"},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic for test case: %s", tt.name)
					}
				}()
			}
			Configure() // Call the main function directly
		})
	}
}

func TestValidArgs(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("can't get pwd")
	}
	os.Args = []string{"gorep", "test", "this is a test"}
	c := newConfig(regexp.MustCompile("const[a-z]{3}"), true, ".", t.TempDir()+"/result", nil)
	if c.Main() != 0 {
		t.Fail()
	}
	t.Chdir(pwd)
	c.file = "main_test.go"
	c.outputPath = ""
	if c.Main() != 0 {
		t.Fail()
	}
}

func TestWalk(t *testing.T) {
	c := newConfig(regexp.MustCompile("test[a-z]{3}"), false, ".", "", nil)
	c.Main()
}

func BenchmarkMain(b *testing.B) {
	fmt.Println(b.N)
	c := newConfig(nil, false, "", "", []string{" this is a testing string"})
	for b.Loop() {
		c.re = regexp.MustCompile("test[a-z]{3}")
		c.Main()
	}
}
