package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"fortio.org/terminal/ansipixels/tcolor"
)

func main() {
	Main()
}

var (
	GREEN = tcolor.Green.Foreground()
	WHITE = tcolor.White.Foreground()
	RED   = tcolor.Red.Foreground()
)

func Main() {
	if len(os.Args) < 2 {
		log.Fatal(("gorep requires a pattern to search"))
	}
	lines := [][2]int{{}}
	noTrim := flag.Bool("no-trim", false,
		"disable trimming leading indentation in each line when printed")
	fileFlag := flag.String("f", "", "take input from a file or directory")
	flag.Parse()
	args := flag.Args()
	var re *regexp.Regexp
	var err error
	re, err = regexp.Compile(args[0])
	if err != nil {
		log.Fatal("invalid input")
	}
	var str string
	if len(args) > 1 {
		str = strings.Join(args[1:], " ")
	}
	switch {
	case *fileFlag != "":
		info, err := os.Stat(*fileFlag)
		if err != nil {
			log.Fatalf("can't open given file or directory")
		}
		if info.IsDir() {
			files := children(*fileFlag)
			matchAllChildren(re, *noTrim, files)
			return
		}
		content, err := os.ReadFile(*fileFlag)
		str = string(content)
		if err != nil {
			log.Fatal("can't open given file")
		}
	case len(args) < 2:
		scanner := bufio.NewScanner(os.Stdin)
		var builder strings.Builder
		index := 0
		for scanner.Scan() {
			_, err := builder.Write(scanner.Bytes())
			builder.WriteByte('\n')
			lines[index][1] = builder.Len()
			index++
			lines = append(lines, [2]int{builder.Len()})
			if err != nil {
				log.Fatal("invalid input")
			}
		}
		str = builder.String()
	}
	match(re, *noTrim, str, "")
}

func matchAllChildren(re *regexp.Regexp, noTrim bool, children [][2]string) {
	for _, file := range children {
		preString := fmt.Sprintf("%s%s: \n", RED, file[0])
		match(re, noTrim, file[1], preString)
	}
}

func match(re *regexp.Regexp, noTrim bool, str string, preString string) {
	i := 0
	emptyCount := 0
	printString := ""
	for line := range strings.Lines(str) {
		matches := re.FindAllString(line, -1)
		indices := re.FindAllStringIndex(line, -1)
		if len(matches) == 0 {
			emptyCount++
			i++
			continue
		}
		printString = fmt.Sprintf("%s%s%d. %s", printString, RED, i+1, WHITE)
		matchBuilder := strings.Builder{}
		curI := 0
		for j, m := range matches {
			pre := line[curI:indices[j][0]]
			post := line[indices[j][1]:]
			if !noTrim {
				pre = strings.Trim(pre, "\t ")
			}
			matchBuilder.WriteString(pre)
			matchBuilder.WriteString(GREEN)
			matchBuilder.WriteString(m)
			matchBuilder.WriteString(WHITE)
			curI = indices[j][1]
			if j == len(matches)-1 {
				matchBuilder.WriteString(post)
			}
		}
		matchString := matchBuilder.String()
		if !noTrim {
			matchString = strings.Trim(matchString, " ")
		}
		printString = fmt.Sprintf("%s%s", printString, matchString)

		i++
	}
	if emptyCount < i {
		printString = fmt.Sprintf("%s%s", preString, printString)
	}
	fmt.Printf("%s", printString)
}

func children(path string) [][2]string {
	files := make([][2]string, 0) // {name, contents}
	entries, err := os.ReadDir(path)
	if err != nil {
		return files
	}
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(path + " is not a directory")
	}
	for _, e := range entries {
		err := os.Chdir(pwd)
		if err != nil {
			log.Fatal(err)
		}
		if e.IsDir() {
			err := os.Chdir(e.Name())
			if err != nil {
				continue
			}
			files = append(files, children(e.Name())...)
			continue
		}
		contents, err := os.ReadFile(e.Name())
		if err != nil {
			continue
		}
		files = append(files, [2]string{e.Name(), string(contents)})
	}
	return files
}
