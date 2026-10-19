package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestUniqEmitsBeforeEOF checks output while the producer keeps stdin open,
// then sends more records to verify deduplication across successive writes.
func TestUniqEmitsBeforeEOF(t *testing.T) {
	// Default buffering must still stream once enough output is available.
	large := strings.Repeat("x", 80*1024)
	for _, tc := range []struct {
		name         string
		args         []string
		input, want  string
		later, final string
	}{
		{"buffered", nil, "key,value\nA," + large + "\n", "key,value\nA,", "A,duplicate\nB,last\n", large + "\nB,last\n"},
		{"immediate", []string{"--immediate-output"}, "key,value\nA,first\nA,duplicate\n", "key,value\nA,first\n", "A,again\nB,last\n", "B,last\n"},
		{"keep two", []string{"--immediate-output", "-n", "2", "-i"}, "key,value\nA,first\na,second\n", "key,value\nA,first\na,second\n", "A,third\nB,last\n", "B,last\n"},
		{"no header", []string{"--immediate-output", "-H"}, "A,first\n", "A,first\n", "A,again\nB,last\n", "B,last\n"},
		{"omit header", []string{"--immediate-output", "-U"}, "key,value\nA,first\n", "A,first\n", "A,again\nB,last\n", "B,last\n"},
		{"quote all tabs", []string{"--immediate-output", "--quote-all", "-t"}, "key\tvalue\nA\t\"first\nline\"\n", "\"key\"\t\"value\"\n\"A\"\t\"first\nline\"\n", "A\tagain\nB\tlast\n", "\"B\"\t\"last\"\n"},
		{"header only", []string{"--immediate-output"}, "key,value\n", "key,value\n", "A,first\n", "A,first\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertCLIStreamsBeforeEOF(t, append([]string{"uniq"}, tc.args...), tc.input, tc.want, tc.later, tc.final)
		})
	}
}

func TestUniqFrequencyModes(t *testing.T) {
	input := writeAccuracyInput(t, "uniq.csv", "key,value\nA,first\nb,first-b\na,second\nc,only\nB,second-b\nA,third\n")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"default", nil, "key,value\nA,first\nb,first-b\na,second\nc,only\nB,second-b\n"},
		{"keep two", []string{"-i", "-n", "2"}, "key,value\nA,first\nb,first-b\na,second\nc,only\nB,second-b\n"},
		{"repeated", []string{"-d"}, "key,value\nA,first\n"},
		{"repeated long", []string{"--repeated", "-i"}, "key,value\nA,first\nb,first-b\n"},
		{"unique", []string{"-u"}, "key,value\nb,first-b\na,second\nc,only\nB,second-b\n"},
		{"unique long case insensitive", []string{"--unique", "-i"}, "key,value\nc,only\n"},
		{"repeated immediate", []string{"-d", "-i", "--immediate-output"}, "key,value\nA,first\nb,first-b\n"},
		{"unique immediate", []string{"-u", "--immediate-output"}, "key,value\nb,first-b\na,second\nc,only\nB,second-b\n"},
		{"no header", []string{"-H", "-i", "-d"}, "A,first\nb,first-b\n"},
		{"delete header", []string{"-U", "-i", "-u"}, "c,only\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"uniq", "-f", "key"}, tc.args...)
			if tc.name == "no header" {
				args = append([]string{"uniq", "-f", "1"}, tc.args...)
			}
			args = append(args, input)
			got, _ := runAccuracyCLI(t, args...)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUniqFrequencyModesCompositeFieldsAndDelimiter(t *testing.T) {
	input := writeAccuracyInput(t, "groups.txt", "a;b;value\nx;y;first\nz;y;only\nx;y;second\nx;z;other\n")
	for _, tc := range []struct {
		flag string
		want string
	}{
		{"-d", "a,b,value\nx,y,first\n"},
		{"-u", "a,b,value\nz,y,only\nx,z,other\n"},
	} {
		got, _ := runAccuracyCLI(t, "uniq", "--delimiter", ";", "-f", "a,b", tc.flag, input)
		if got != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.flag, got, tc.want)
		}
	}
	got, _ := runAccuracyCLI(t, "--delimiter", ";", "uniq", "-f", "a,b", "-d", input)
	if got != "a,b,value\nx,y,first\n" {
		t.Fatalf("delimiter before subcommand: got %q", got)
	}

	headerOnly := writeAccuracyInput(t, "header.csv", "a,b\n")
	for _, flag := range []string{"-d", "-u"} {
		got, _ := runAccuracyCLI(t, "uniq", flag, headerOnly)
		if got != "a,b\n" {
			t.Fatalf("%s with header only: got %q", flag, got)
		}
	}
}

func TestUniqFrequencyModesRejectKeepNAndEachOther(t *testing.T) {
	input := writeAccuracyInput(t, "uniq.csv", "key\nA\nA\n")
	for _, flags := range [][]string{{"-d", "-u"}, {"-d", "-n", "2"}, {"-u", "-n", "1"}} {
		args := append([]string{"-test.run=^TestCLIAccuracyHelper$", "--", "uniq"}, flags...)
		args = append(args, input)
		cmd := exec.Command(os.Args[0], args...)
		cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil || !strings.Contains(stderr.String(), "if any flags in the group") {
			t.Errorf("uniq %v: expected mutually exclusive error, got %v: %s", flags, err, stderr.String())
		}
	}
}

func TestUniqFrequencyModesFromStdin(t *testing.T) {
	for _, tc := range []struct {
		flag string
		want string
	}{
		{"-d", "key,value\nx,first\n"},
		{"-u", "key,value\ny,only\n"},
	} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "uniq", "-f", "key", tc.flag, "-")
		cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
		cmd.Stdin = strings.NewReader("key,value\nx,first\ny,only\nx,last\n")
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil || stdout.String() != tc.want {
			t.Errorf("uniq %s on stdin: got %q, want %q (error: %v; stderr: %s)", tc.flag, stdout.String(), tc.want, err, stderr.String())
		}
	}
}
