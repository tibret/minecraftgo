package wrapper

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	"github.com/looplab/fsm"
)

func JavaExecCmd(serverPath string, initialHeapSize, maxHeapSize int) *exec.Cmd {
	initialHeapFlag := fmt.Sprintf("-Xms%dM", initialHeapSize)
	maxHeapFlag := fmt.Sprintf("-Xmx%dM", maxHeapSize)
	return exec.Command("java", initialHeapFlag, maxHeapFlag, "-jar", serverPath, "nogui")
}

type Console struct {
	cmd    *exec.Cmd
	stdout *bufio.Reader
	stdin  *bufio.Writer
}

func NewConsole(cmd *exec.Cmd) *Console {
	c := &Console{
		cmd: cmd,
	}

	stdout, _ := cmd.StdoutPipe()
	c.stdout = bufio.NewReader(stdout)

	stdin, _ := cmd.StdinPipe()
	c.stdin = bufio.NewWriter(stdin)

	return c
}

func (c *Console) Start() error {
	return c.cmd.Start()
}

func (c *Console) WriteCmd(cmd string) error {
	wrappedCmd := fmt.Sprintf("%s\r\n", cmd)
	_, err := c.stdin.WriteString(wrappedCmd)
	if err != nil {
		return err
	}
	return c.stdin.Flush()
}

func (c *Console) ReadLine() (string, error) {
	return c.stdout.ReadString('\n')
}

func (c *Console) Kill() error {
	return c.cmd.Cancel()
}

var logRegex = regexp.MustCompile(`(\[[0-9:]*\]) \[([A-z(-| )#0-9]*)\/([A-z #]*)\]: (.*)`)

type LogLine struct {
	timestamp  string
	threadName string
	level      string
	output     string
}

func (ll *LogLine) Match(regex *regexp.Regexp) bool {
	return regex.Match([]byte(ll.output))
}

func ParseToLogLine(line string) *LogLine {
	fmt.Println("Parsing line", line)
	matches := logRegex.FindAllStringSubmatch(line, 4)
	if len(matches) < 1 {
		return &LogLine{
			timestamp:  "",
			threadName: "",
			level:      "",
			output:     "",
		}
	}

	return &LogLine{
		timestamp:  matches[0][1],
		threadName: matches[0][2],
		level:      matches[0][3],
		output:     matches[0][4],
	}
}

type Event string

const (
	EmptyEvent   Event = "empty"
	StartedEvent       = "started"
	StoppedEvent       = "stopped"
	StartEvent         = "start"
	StopEvent          = "stop"
)

var eventToRegexp = map[Event]*regexp.Regexp{
	StartedEvent: regexp.MustCompile(`Done (?s)(.*)! For help, type "help"`),
	StartEvent:   regexp.MustCompile(`Starting minecraft server version (.*)`),
	StopEvent:    regexp.MustCompile(`Stopping (.*) server`),
}

func LogParser(line string) Event {
	ll := ParseToLogLine(line)
	for e, r := range eventToRegexp {
		if ll.Match(r) {
			return e
		}
	}

	return EmptyEvent
}

const (
	ServerOffline  = "offline"
	ServerOnline   = "online"
	ServerStarting = "starting"
	ServerStopping = "stopping"
)

type Wrapper struct {
	console    *Console
	machine    *fsm.FSM
	LastLine   string
	OutputChan chan string
}

func (w *Wrapper) Start() error {
	go w.processLogEvents()
	return w.console.Start()
}

func (w *Wrapper) Stop() error {
	return w.console.WriteCmd("stop")
}

func NewWrapper(c *Console) *Wrapper {
	return &Wrapper{
		console: c,
		machine: fsm.NewFSM(
			ServerOffline,
			fsm.Events{
				fsm.EventDesc{
					Name: StopEvent,
					Src:  []string{ServerOnline},
					Dst:  ServerStopping,
				},
				fsm.EventDesc{
					Name: StoppedEvent,
					Src:  []string{ServerStopping},
					Dst:  ServerOffline,
				},
				fsm.EventDesc{
					Name: StartEvent,
					Src:  []string{ServerOffline},
					Dst:  ServerStarting,
				},
				fsm.EventDesc{
					Name: StartedEvent,
					Src:  []string{ServerStarting},
					Dst:  ServerOnline,
				},
			},
			nil,
		),
	}
}

func (w *Wrapper) processLogEvents() {
	for {
		line, err := w.console.ReadLine()
		w.LastLine = line
		if w.OutputChan != nil {
			w.OutputChan <- w.LastLine
		}
		if err == io.EOF {
			w.updateState(StoppedEvent)
			continue
		}

		event := LogParser(line)
		fmt.Println("Processing Event", string(event))
		w.updateState(event)
		fmt.Println("Current state", w.machine.Current())
	}
}

func (w *Wrapper) updateState(ev Event) error {
	if ev == EmptyEvent {
		return nil
	}
	return w.machine.Event(context.Background(), string(ev))
}

func (w *Wrapper) SendCommand(cmd string) string {
	if w.machine.Is(ServerOnline) {
		ch := make(chan string)
		defer close(ch)
		w.OutputChan = ch
		w.console.WriteCmd(cmd)
		response := <-w.OutputChan
		w.OutputChan = nil
		return response
	}

	return "Server not online"
}
