package main

import (
	"flag"
	"fmt"
	"github.com/howeyc/fsnotify"
	"github.com/remogatto/application"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"
)

const (
	// Multiple events that occur for the same file in this
	// time windows will be discarded.
	DISCARD_TIME = 1 * time.Second
	RERUN_TIME   = 2 * time.Second
)

var (
	events  map[string]*eventOnFile
	rwMutex sync.RWMutex
)

// eventOnFile stores informations about events occured on a file
type eventOnFile struct {
	fsnotifyEvent *fsnotify.FileEvent
	time          time.Time
}

func addEvent(event *eventOnFile) *eventOnFile {
	rwMutex.Lock()
	events[event.fsnotifyEvent.Name] = event
	rwMutex.Unlock()
	return event
}

func getEvent(filename string) *eventOnFile {
	rwMutex.RLock()
	event, ok := events[filename]
	rwMutex.RUnlock()
	if ok {
		return event
	}
	return nil
}

// sigterm is a type for handling a SIGTERM signal.
type sigterm struct {
	hitCounter byte
	paths   []string
}

func (h *sigterm) HandleSignal(s os.Signal) {
	switch ss := s.(type) {
	case syscall.Signal:
		switch ss {
		case syscall.SIGTERM, syscall.SIGINT:
			if h.hitCounter > 0 {
				application.Exit()
				return
			}
			application.Printf("Hit CTRL-C again to exit, otherwise tests will run again in %s.", RERUN_TIME)
			h.hitCounter++
			go func() {
				time.Sleep(RERUN_TIME)
				execGoTest(h.paths)
				h.hitCounter = 0
			}()
		}
	}
}

// watchLoop watches for changes in the folder
type watcherLoop struct {
	pause, terminate chan int
	initialPath string
	paths         []string
}

func newWatcherLoop(initialPath string) *watcherLoop {
	return &watcherLoop{make(chan int), make(chan int), initialPath, folders(initialPath)}
}

func (l *watcherLoop) Pause() chan int {
	return l.pause
}

func (l *watcherLoop) Terminate() chan int {
	return l.terminate
}

func (l *watcherLoop) Run() {
	// Run the tests for the first time.
	execGoTest(l.paths)

	watcher, err := fsnotify.NewWatcher()

	for _, path := range l.paths {
		err = watcher.Watch(path)
		if err != nil {
			application.Fatal(err.Error())
		}
		application.Printf("Start watching path %s", path)
	}

	for {
		select {
		case <-l.pause:
			l.pause <- 0
		case <-l.terminate:
			watcher.Close()
			l.terminate <- 0
			return
		case ev := <-watcher.Event:
			if ev.IsModify() {
				if matches(ev.Name, ".*\\.go$") {
					if application.Verbose {
						application.Logf("Event %s occured for file %s", ev, ev.Name)
					}
					// check if the same event was
					// registered for the same
					// file in the acceptable
					// TIME_DISCARD time window
					event := getEvent(ev.Name)
					if event == nil {
						event = addEvent(&eventOnFile{ev, time.Now()})
						application.Logf("Run the tests")
						execGoTest(l.paths)
					} else if time.Now().Sub(event.time) > DISCARD_TIME {
						event.time = time.Now()
						application.Logf("Run the tests")
						execGoTest(l.paths)
					} else {
						if application.Verbose {
							application.Logf("Event %s was discarded for file %s", ev, ev.Name)
						}
					}
				}
			}
		case err := <-watcher.Error:
			application.Fatal(err.Error())
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "PrettyAutoTest continously watches for changes in folder and re-run the tests\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n\n")
	fmt.Fprintf(os.Stderr, "\tprettyautotest [options]\n\n")
	fmt.Fprintf(os.Stderr, "Options are:\n\n")
	flag.PrintDefaults()
}

// Returns whether 's' matches 'pattern'
func matches(s, pattern string) bool {
	return regexp.MustCompile(pattern).MatchString(s)
}

func execGoTest(paths []string) {
	for _, path := range paths {
		// Execute go test -i first
		cmd := exec.Command("go", "test", "-i")
		cmd.Dir = path
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err)
		}
		cmd = exec.Command("go", "test")
		cmd.Dir = path
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err)
		}
		fmt.Print(string(out))
	}
}

func init() {
	events = make(map[string]*eventOnFile, 0)
}

// Returns a slice of subfolders (recursive), including the folder passed in
func folders(path string) (paths []string) {
	filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			// skip folders that begin with a dot and
			// folders without test files
			hidden := filepath.HasPrefix(name, ".") && name != "." && name != ".."
			testFiles, _ := filepath.Glob(filepath.Join(name, "*_test.go"))
			if hidden || len(testFiles) == 0 {
				return filepath.SkipDir
			} else {
				paths = append(paths, newPath)
			}
		}
		return nil
	})
	return paths
}

func main() {
	verbose := flag.Bool("verbose", false, "Verbose mode")
	help := flag.Bool("help", false, "Show usage")
	flag.Usage = usage
	flag.Parse()

	initialPath := flag.Arg(0)
	if initialPath == "" {
		initialPath = "./"
	}

	if _, err := os.Stat(initialPath); err != nil {
		application.Fatal(err.Error())
	}

	application.Verbose = *verbose
	if *help {
		usage()
		return
	}
	watcherLoop := newWatcherLoop(initialPath)
	application.Register("Watcher Loop", watcherLoop)
	application.InstallSignalHandler(&sigterm{paths: watcherLoop.paths})
	exitCh := make(chan bool)
	application.Run(exitCh)
	<-exitCh
}
