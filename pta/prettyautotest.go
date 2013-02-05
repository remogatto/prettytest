package main

import (
	"os/exec"
	"regexp"
	"os"
	"fmt"
	"log"
	"flag"
	"syscall"
	"sync"
	"time"
	"github.com/remogatto/application"
	"github.com/howeyc/fsnotify"
)

const (
	// Multiple events that occur for the same file in this
	// time windows will be discarded.
	DISCARD_TIME = 1 * time.Second
)

var (
	events map[string]*eventOnFile
	rwMutex sync.RWMutex
)

// eventOnFile stores informations about events occured on a file
type eventOnFile struct {
	fsnotifyEvent *fsnotify.FileEvent
	time time.Time
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
			application.Printf("%s\n", "Hit CTRL-C again to exit from PrettyAutoTest")
			h.hitCounter++
		}
	}
}

// watchLoop watches for changes in the folder
type watcherLoop struct {
	pause, terminate chan int
	watchDir string
}

func (l *watcherLoop) Pause() chan int {
	return l.pause
}

func (l *watcherLoop) Terminate() chan int {
	return l.terminate
}

func (l *watcherLoop) Run() {
	watcher, err := fsnotify.NewWatcher()
	err = watcher.Watch(l.watchDir)
	if err != nil {
		application.Fatal(err.Error())
	}
	application.Printf("Start watching path %s", l.watchDir)
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
						execGoTest(l.watchDir)
					} else if time.Now().Sub(event.time) > DISCARD_TIME {
						event.time = time.Now()
						application.Logf("Run the tests")
						execGoTest(l.watchDir)
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

func newWatcherLoop(watchDir string) *watcherLoop {
	return &watcherLoop{make(chan int), make(chan int), watchDir}
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

func execGoTest(path string) {
	cmd := exec.Command("go", "test")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	fmt.Print(string(out))
}

func init() {
	events = make(map[string]*eventOnFile, 0)
}

func main() {
	watchDir := flag.String("watchdir", "./", "Directory to watch for changes")
	verbose  := flag.Bool("verbose", false, "Verbose mode")
	help := flag.Bool("help", false, "Show usage")
	flag.Usage = usage
	flag.Parse()
	application.Verbose = *verbose
	if *help {
		usage()
		return
	}
	application.Register("Watcher Loop", newWatcherLoop(*watchDir))
	application.InstallSignalHandler(&sigterm{})
	exitCh := make(chan bool)
	application.Run(exitCh)
	<-exitCh
}

