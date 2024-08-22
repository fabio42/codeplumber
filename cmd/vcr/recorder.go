package vcr

// This awesome code is shamelessly taken from https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

var lock sync.Mutex

// Marshal is a function that marshals the object into an
// io.Reader.
// By default, it uses the JSON marshaller.
var Marshal = func(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Unmarshal is a function that unmarshals the data from the
// reader into the specified value.
// By default, it uses the JSON unmarshaller.
var Unmarshal = func(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// Recorder is a struct that handle all recorder settings
type Recorder struct {
	Mode               string
	Recording, Playing bool
	RecordDir          string
}

// NewRecorder returns a new Recorder
func NewRecorder(record, play bool, dir string) *Recorder {
	var r Recorder
	// create dir if not exist
	if record || play {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Debug().Str("model", "vcr").Str("func", "NewRecorder").Msgf("creating record directory %s", dir)
			os.MkdirAll(dir, os.ModePerm)
		}
	}
	r.Recording = record
	r.Playing = play
	r.RecordDir = dir
	return &r
}

// Record saves the interface i to a file f
func (r *Recorder) Record(f string, i interface{}) error {
	var err error
	if r.Recording {
		err = Save(fmt.Sprintf("%s/%s", r.RecordDir, f), i)
	}
	return err
}

// Play loads the interface i from a file f
func (r *Recorder) Play(f string, i interface{}) (interface{}, error) {
	var err error
	var out interface{}
	var b []byte
	file := fmt.Sprintf("%s/%s", r.RecordDir, f)
	if r.Playing {
		err = Load(file, i)
	}
	if err.Error() == "json: Unmarshal(non-pointer string)" {
		b, err = os.ReadFile(file)
		out = strings.ReplaceAll(string(b), "\"", "")
	}

	return out, err
}

// Save saves a representation of v to the file at path.
func Save(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// Load loads the file at path into v.
// Use os.IsNotExist() to see if the returned error is due
// to the file being missing.
func Load(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Unmarshal(f, v)
}
