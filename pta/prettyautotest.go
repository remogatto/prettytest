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
    watchDir   string
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
            application.Printf("Hit CTRL-C again to exit otherwise tests will be re-runned in %s.", RERUN_TIME)
            h.hitCounter++
            go func() {
                time.Sleep(RERUN_TIME)
                execGoTest(h.watchDir)
                h.hitCounter = 0
            }()
        }
    }
}

// watchLoop watches for changes in the folder
type watcherLoop struct {
    pause, terminate chan int
    watchDir         string
}

func newWatcherLoop(watchDir string) *watcherLoop {
    return &watcherLoop{make(chan int), make(chan int), watchDir}
}

func (l *watcherLoop) Pause() chan int {
    return l.pause
}

func (l *watcherLoop) Terminate() chan int {
    return l.terminate
}

func (l *watcherLoop) Run() {
    // Run the tests for the first time.
    execGoTest(l.watchDir)

    watcher, err := fsnotify.NewWatcher()

    for _, folder := range folders(l.watchDir) {
        err = watcher.Watch(folder)
        if err != nil {
            application.Fatal(err.Error())
        }
        application.Printf("Start watching path %s", folder)
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
    cmd := exec.Command("go", "test", "./...")
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

// returns a slice of subfolders (recursive), including the folder passed in
func folders(path string) (paths []string) {
    filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            name := info.Name()
            // skip folders that begin with a dot
            hidden := filepath.HasPrefix(name, ".") && name != "." && name != ".."
            if hidden {
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
    watchDir := flag.String("watchdir", "./", "Directory to watch for changes")
    verbose := flag.Bool("verbose", false, "Verbose mode")
    help := flag.Bool("help", false, "Show usage")
    flag.Usage = usage
    flag.Parse()

    application.Verbose = *verbose
    if *help {
        usage()
        return
    }
    application.Register("Watcher Loop", newWatcherLoop(*watchDir))
    application.InstallSignalHandler(&sigterm{watchDir: *watchDir})
    exitCh := make(chan bool)
    application.Run(exitCh)
    <-exitCh
}
