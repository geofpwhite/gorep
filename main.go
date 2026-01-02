package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

type config struct {
	trim        bool
	file        string
	outputPath  string
	re          *regexp.Regexp
	args        []string
	fileChannel chan [2]string
}

func newConfig(re *regexp.Regexp, trim bool, file string, outputPath string, args []string) *config {
	return &config{
		trim:        trim,
		file:        file,
		outputPath:  outputPath,
		re:          re,
		args:        args,
		fileChannel: make(chan [2]string),
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
			log.Println("output file already exists")
			return 1
		}
		opf, err = os.Create(c.outputPath)
		if err != nil {
			log.Println("output file couldn't be created")
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
			log.Println("can't open given file or directory")
			return 1
		}
		if info.IsDir() {
			files := c.collect(c.file)
			c.matchAllChildren(files, opf)
			return 0
		}
		content, err := os.ReadFile(c.file)
		str = string(content)
		if err != nil {
			log.Println("can't open given file")
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
				log.Println("invalid input")
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
		indices := c.re.FindAllStringIndex(line, -1)
		if len(indices) == 0 {
			emptyCount++
			i++
			continue
		}
		printString = fmt.Sprintf("%s%s%d. %s", printString, RED, i+1, WHITE)
		matchBuilder := strings.Builder{}
		curI := 0
		lengthMatches := len(indices)
		for j, ary := range indices {
			m := line[ary[0]:ary[1]]
			pre := line[curI:ary[0]]
			if c.trim {
				pre = strings.TrimLeft(pre, "\t")
			}
			matchBuilder.WriteString(fmt.Sprintf("%s%s%s%s", pre, GREEN, m, WHITE))
			curI = ary[1]
			if j != lengthMatches-1 {
				continue
			}
			post := line[ary[1]:]
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
	fmt.Print(printString)
	if output != nil {
		forOutputFile := printString
		cleanedBytes, _ := ansipixels.AnsiClean([]byte(forOutputFile))
		forOutputFile = string(cleanedBytes)
		_, err := output.WriteString(forOutputFile)
		if err != nil {
			log.Println("couldn't write output")
		}
	}
}

func walk(path string, returnChan chan [2]string) error {
	var walkFunc func(path string, d fs.DirEntry, err error) error
	visited := make(map[string]bool)
	wg := &sync.WaitGroup{}
	walkFunc = func(newPath string, d fs.DirEntry, _ error) error {
		if visited[newPath] {
			return nil
		}
		visited[newPath] = true

		if d.IsDir() {
			return filepath.WalkDir(newPath, walkFunc)
		}
		wg.Add(1)
		go func() {
			read(returnChan, newPath, newPath)
			wg.Done()
		}()
		return nil
	}
	err := filepath.WalkDir(path, walkFunc)
	wg.Wait()
	close(returnChan)
	return err
}

func read(returnChannel chan [2]string, path string, name string) {
	contents, err := os.ReadFile(path)
	if err != nil || !utf8.Valid(contents) {
		return
	}
	returnChannel <- [2]string{name, string(contents)}
}

func (c *config) collect(path string) [][2]string {
	return collect(path, c.fileChannel)
}

func collect(file string, fileChannel chan [2]string) [][2]string {
	returning := make([][2]string, 0, 100)
	done := make(chan bool)
	go func() {
		for file := range fileChannel {
			returning = append(returning, file)
		}
		done <- true
	}()
	err := walk(file, fileChannel)
	if err != nil {
		return nil
	}
	<-done
	return returning
}

func main() {
	c := Configure()
	c.Main()
}
