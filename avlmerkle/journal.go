package avlmerkle

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
)

const fileJournal string = "residual.log"

type journalFile struct {
	file    *os.File
	scanner *bufio.Scanner
}

type Journal struct {
	sync.Mutex
	journals journalFile
}

// NewJournal creates a new Journal context. One can use a single journal
// context for many journals.
func NewJournal() (*Journal, error) {
	if _, err := os.Lstat(fileJournal); !os.IsNotExist(err) {
		f, err := os.OpenFile(fileJournal, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			return nil, err
		}
		return &Journal{
			journals: journalFile{
				file:    f,
				scanner: bufio.NewScanner(f),
			},
		}, nil
	}

	fileobj, _ := os.Create(fileJournal)
	return &Journal{
		journals: journalFile{
			file:    fileobj,
			scanner: bufio.NewScanner(fileobj),
		},
	}, nil
}

func FileExist() error {
	err := syscall.Access(fileJournal, syscall.F_OK)
	if !os.IsNotExist(err) {
		bytes, err := ioutil.ReadFile(fileJournal)
		if err != nil {
			return err
		}

		if len(bytes) == 0 {
			return errors.New("file has no content")
		}
		return nil
	}
	return err

}

// Journal writes content to a journal file. Note that content should not be
// bigger than bufio.Scanner can read per line. If the user does not provide
// "\n" at the end of content string, this function appends it.
func (j *Journal) JouWrite(content string) error {
	j.Lock()
	defer j.Unlock()

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	f, err := os.OpenFile(j.journals.file.Name(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		return err
	}
	defer f.Sync()
	defer f.Close()

	_, err = f.Write([]byte(content))
	return err
}

func (j *Journal) Close() error {
	j.Lock()
	defer j.Unlock()
	return j.journals.file.Close()
}

// // Replay reads a single line from the journal file and calls the replay
// // function that was provided. If the scanner encounters EOF it returns EOF,
// // unlike the scanner API.
func (j *Journal) Replay() (*bufio.Scanner, error) {
	j.Lock()
	defer j.Unlock()

	// We can run unlocked from here

	if !j.journals.scanner.Scan() {
		if j.journals.scanner.Err() == nil {
			return nil, io.EOF
		}
		return nil, j.journals.scanner.Err()
	}
	for j.journals.scanner.Scan() {
	}
	return j.journals.scanner, nil
}

func (j *Journal) StrToMap(ScanText string, root string) map[string]interface{} {
	result := make(map[string]interface{}, 1)
	flysnowRegexp := regexp.MustCompile(`root:.(.*).key:.(.*).bs:.(.*)$`)
	params := flysnowRegexp.FindStringSubmatch(ScanText)

	if params[1] != root {
		return nil
	}

	result["root"] = params[1]
	result["key"] = params[2]
	result["bs"] = params[3]

	return result

}

// Delete successfully submitted data
func (j *Journal) DelWellDate() error {
	return j.journals.file.Truncate(0)
}
