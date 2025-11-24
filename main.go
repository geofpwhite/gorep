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

type config struct {
	trim       bool
	file       string
	outputPath string
	re         *regexp.Regexp
	args       []string
}

func newConfig(re *regexp.Regexp, trim bool, file string, outputPath string, args []string) *config {
	return &config{
		trim:       trim,
		file:       file,
		outputPath: outputPath,
		re:         re,
		args:       args,
	}
}

func Configure() *config {
	if len(os.Args) < 2 {
		panic("gorep needs a pattern to match")
	}
	noTrim := flag.Bool("no-trim", false,
		"disable trimming leading indentation in each line when printed")
	fileFlag := flag.String("f", "", "take input from a file or directory")
	outputFile := flag.String("o", "", "save the matches to a file")
	flag.Parse()
	args := flag.Args()
	var re *regexp.Regexp
	var err error
	_ = flag.CommandLine.Parse(args[1:])
	re, err = regexp.Compile(args[0])
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
	return newConfig(re, !*noTrim, *fileFlag, *outputFile, args)
}

func main() {
	c := Configure()
	c.Main()
}

var (
	GREEN = tcolor.Green.Foreground()
	WHITE = tcolor.White.Foreground()
	RED   = tcolor.Red.Foreground()
	BLUE  = tcolor.Blue.Foreground()
)

func (c *config) Main() int {
	lines := [][2]int{{}}
	var opf *os.File
	if c.outputPath != "" {
		_, err := os.ReadFile(c.outputPath)
		if err == nil {
			log.Print("output file already exists")
			return 1
		}
		opf, err = os.Create(c.outputPath)
		if err != nil {
			log.Print("output file couldn't be created")
			return 1
		}
		defer opf.Close()
	}

	var str string
	if len(c.args) > 1 {
		str = strings.Join(c.args[1:], " ")
	}
	switch {
	case c.file != "":
		info, err := os.Stat(c.file)
		if err != nil {
			fmt.Println("can't open given file or directory")
			return 1
		}
		if info.IsDir() {
			files := recursiveFileSearch(c.file)
			c.matchAllChildren(files, opf)
			return 0
		}
		content, err := os.ReadFile(c.file)
		str = string(content)
		if err != nil {
			fmt.Println("can't open given file")
			return 1
		}
	case len(c.args) < 2:
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
				return 1
			}
		}
		str = builder.String()
	}
	c.match(str, "", opf)
	return 0
}

func (c *config) matchAllChildren(children [][2]string, outputFile *os.File) {
	for _, file := range children {
		preString := fmt.Sprintf("%s%s: \n", BLUE, file[0])
		c.match(file[1], preString, outputFile)
	}
}

func (c *config) match(str string, preString string, output *os.File) {
	i := 0
	emptyCount := 0
	printString := ""
	for line := range strings.Lines(str) {
		matches := c.re.FindAllString(line, -1)
		indices := c.re.FindAllStringIndex(line, -1)
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
			if c.trim {
				pre = strings.TrimLeft(pre, "\t")
			}
			matchBuilder.WriteString(fmt.Sprintf("%s%s%s%s", pre, GREEN, m, WHITE))
			curI = indices[j][1]
			if j != lengthMatches-1 {
				continue
			}
			post := line[indices[j][1]:]
			if c.trim {
				post = strings.TrimRight(post, "\t\n")
			}
			matchBuilder.WriteString(post)
		}
		matchString := matchBuilder.String()
		if c.trim {
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

	if path != "." && path != "./" && path != ".\\" && len(pwd)-len(path) > 0 && pwd[len(pwd)-len(path):] != path {
		pwd += "/" + path
	}

	if path != "." && (path[:2] == "C:" || path[0] == '/') {
		pwd = path
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
			continue
		}
		files = append(files, [2]string{e.Name(), string(contents)})
	}
	return files
}
