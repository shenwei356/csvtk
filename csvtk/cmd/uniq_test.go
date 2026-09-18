package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

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
