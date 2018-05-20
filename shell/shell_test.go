package shell

import (
	"strings"
	"testing"
	"time"

	"github.com/cortesi/termlog"
)

type cmdTest struct {
	cmd     string
	bufferr bool

	logHas  string
	buffHas string
	err     bool
	procerr bool
	kill    bool
}

func testCmd(t *testing.T, shell string, ct cmdTest) {
	lt := termlog.NewLogTest()
	exec, err := NewExecutor(shell, ct.cmd)
	if err != nil {
		t.Error(err)
		return
	}
	type result struct {
		err     error
		procerr error
		buff    string
	}

	ch := make(chan result)
	go func() {
		err, procerr, buff := exec.Run(lt.Log.Stream(""), ct.bufferr)
		ch <- result{err: err, procerr: procerr, buff: buff}
	}()

	if ct.kill {
		for {
			if exec.Running() {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		err := exec.Stop()
		if err != nil {
			t.Errorf("Error stopping: %s", err)
			return
		}
	}

	res := <-ch
	if (res.err != nil) != ct.err {
		t.Errorf("Unexpected invocation error: %s", err)
	}
	if (res.procerr != nil) != ct.procerr {
		t.Errorf("Unexpected process error: %s", res.procerr)
	}
	if ct.buffHas != "" && !strings.Contains(res.buff, ct.buffHas) {
		t.Errorf("Unexpected buffer return: %s", res.buff)
	}
	if ct.logHas != "" && !strings.Contains(lt.String(), ct.logHas) {
		t.Errorf("Unexpected log return: %s", lt.String())
	}
}

var bashTests = []cmdTest{
	{
		cmd:    "echo moddtest; true",
		logHas: "moddtest",
	},
	{
		cmd:     "echo moddtest; false",
		logHas:  "moddtest",
		procerr: true,
	},
	{
		cmd:     "definitelynosuchcommand",
		procerr: true,
	},
	{
		cmd:     "echo moddstderr >&2",
		bufferr: true,
		buffHas: "moddstderr",
	},
	{
		cmd:     "echo moddtest; sleep 999999",
		logHas:  "moddtest",
		kill:    true,
		procerr: true,
	},
}

func TestBash(t *testing.T) {
	if _, err := getBash(); err != nil {
		t.Skip("skipping bash tests")
		return
	}
	for _, tc := range bashTests {
		testCmd(t, "bash", tc)
	}
}
