package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"fortio.org/terminal/ansipixels/tcolor"
)

func main() {
	Main()
}

var (
	GREEN = tcolor.Green.Foreground()
	WHITE = tcolor.White.Foreground()
	RED   = tcolor.Red.Foreground()
	BLUE  = tcolor.Blue.Foreground()
)

func Main() {
	if len(os.Args) < 2 {
		log.Fatal(("gorep requires a pattern to search"))
	}
	lines := [][2]int{{}}
	noTrim := flag.Bool("no-trim", false,
		"disable trimming leading indentation in each line when printed")
	fileFlag := flag.String("f", "", "take input from a file or directory")
	outputFile := flag.String("o", "", "save the matches to a file")
	flag.Parse()
	args := flag.Args()
	var re *regexp.Regexp
	var err error
	err = flag.CommandLine.Parse(args[1:])
	if err != nil {
		log.Fatal("bad input")
	}
	var opf *os.File
	if *outputFile != "" {
		_, err = os.ReadFile(*outputFile)
		if err == nil {
			log.Fatal("output file already exists")
		}
		opf, err = os.Create(*outputFile)
		if err != nil {
			log.Fatal("output file couldn't be created")
		}
		defer opf.Close()
	}

	re, err = regexp.Compile(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	var str string
	if len(args) > 1 {
		str = strings.Join(args[1:], " ")
	}
	switch {
	case *fileFlag != "":
		info, err := os.Stat(*fileFlag)
		if err != nil {
			fmt.Println("can't open given file or directory")
			return
		}
		if info.IsDir() {
			files := recursiveFileSearch(*fileFlag)
			matchAllChildren(re, *noTrim, files, opf)
			return
		}
		content, err := os.ReadFile(*fileFlag)
		str = string(content)
		if err != nil {
			fmt.Println("can't open given file")
			return
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
				fmt.Println("invalid input")
				return
			}
		}
		str = builder.String()
	}
	match(re, *noTrim, str, "", opf)
}

func matchAllChildren(re *regexp.Regexp, noTrim bool, children [][2]string, outputFile *os.File) {
	for _, file := range children {
		preString := fmt.Sprintf("%s%s: \n", BLUE, file[0])
		match(re, noTrim, file[1], preString, outputFile)
	}
}

func match(re *regexp.Regexp, noTrim bool, str string, preString string, output *os.File) {
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
		lengthMatches := len(matches)
		for j, m := range matches {
			pre := line[curI:indices[j][0]]
			if !noTrim {
				pre = strings.TrimLeft(pre, "\t")
			}
			matchBuilder.WriteString(fmt.Sprintf("%s%s%s%s", pre, GREEN, m, WHITE))
			curI = indices[j][1]
			if j != lengthMatches-1 {
				continue
			}
			post := line[indices[j][1]:]
			if !noTrim {
				post = strings.TrimRight(post, "\t\n")
			}
			matchBuilder.WriteString(post)
		}
		matchString := matchBuilder.String()
		if !noTrim {
			matchString = strings.Trim(matchString, " ")
		}
		printString = fmt.Sprintf("%s%s\n", printString, matchString)

		i++
	}
	if emptyCount < i {
		printString = fmt.Sprintf("%s%s", preString, printString)
	}
	fmt.Printf("%s", printString)
	if output != nil {
		forOutputFile := printString
		for _, toRemove := range []string{GREEN, RED, BLUE, WHITE} {
			forOutputFile = strings.ReplaceAll(forOutputFile, toRemove, "")
		}
		// forOutputFile := strings.ReplaceAll(printString, GREEN, "")
		// forOutputFile = strings.ReplaceAll(forOutputFile, RED, "")
		// forOutputFile = strings.ReplaceAll(forOutputFile, BLUE, "")
		// forOutputFile = strings.ReplaceAll(forOutputFile, WHITE, "")
		_, err := output.WriteString(forOutputFile)
		if err != nil {
			fmt.Println("couldn't write output")
		}
	}
}

func recursiveFileSearch(path string) [][2]string {
	files := make([][2]string, 0) // {name, contents}
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("can't open pwd")
	}
	if path != "." && path != "./" && path != ".\\" && pwd[len(pwd)-len(path):] != path {
		pwd += "/" + path
	}
	entries, err := os.ReadDir(pwd)
	if err != nil {
		return files
	}
	for _, e := range entries {
		err := os.Chdir(pwd)
		if err != nil {
			log.Fatal(err)
		}
		if e.IsDir() {
			err = os.Chdir(e.Name())
			if err != nil {
				log.Fatal(err)
			}
			files = append(files, recursiveFileSearch(e.Name())...)
			continue
		}
		contents, err := os.ReadFile(e.Name())
		if err != nil || !utf8.Valid(contents) {
			fmt.Println(err)
			continue
		}
		files = append(files, [2]string{e.Name(), string(contents)})
	}
	return files
}
