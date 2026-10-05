// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux && !android) || freebsd

package notify

import (
	"bufio"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

type notifyCall struct {
	replaces      uint32
	summary, body string
	actions       []string
	hints         map[string]dbus.Variant
}

type fakeServer struct {
	mu    sync.Mutex
	calls []notifyCall
	next  uint32
}

func (f *fakeServer) GetCapabilities() ([]string, *dbus.Error) {
	return []string{"body", "body-markup", "actions"}, nil
}

func (f *fakeServer) Notify(app string, replaces uint32, icon, summary, body string, actions []string, hints map[string]dbus.Variant, timeout int32) (uint32, *dbus.Error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, notifyCall{replaces, summary, body, actions, hints})
	if replaces != 0 {
		return replaces, nil
	}
	f.next++
	return f.next, nil
}

func (f *fakeServer) wait(t *testing.T, n int) []notifyCall {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		f.mu.Lock()
		calls := append([]notifyCall(nil), f.calls...)
		f.mu.Unlock()
		if len(calls) >= n {
			return calls
		}
	}
	t.Fatalf("fewer than %d notifications", n)
	return nil
}

// privateBus starts a session bus of the test's own, or skips it.
func privateBus(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("dbus-daemon"); err != nil {
		t.Skip("no dbus-daemon")
	}
	cmd := exec.Command("dbus-daemon", "--session", "--nofork", "--print-address")
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	address, err := bufio.NewReader(out).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", strings.TrimSpace(address))
}

// Notifications go to the desktop's service with their text escaped for
// its markup and the sound asked for; one about the same thing replaces
// the last, and a click opens it with the activation token.
func TestDBusNotifications(t *testing.T) {
	privateBus(t)
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	server := &fakeServer{}
	if err = conn.ExportMethodTable(map[string]any{
		"GetCapabilities": server.GetCapabilities,
		"Notify":          server.Notify,
	}, busPath, busIface); err != nil {
		t.Fatal(err)
	}
	if reply, err := conn.RequestName(busName, dbus.NameFlagDoNotQueue); err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		t.Fatal("name", reply, err)
	}

	n := New("KomaruGram", nil)
	defer n.Close()
	opened := make(chan string, 1)
	n.Show(Notification{Title: "Team", Body: "Alice: <b>&</b>", Sound: true, Tag: "a/1", Open: func(token string) { opened <- token }})
	calls := server.wait(t, 1)
	c := calls[0]
	if c.summary != "Team" || c.body != "Alice: &lt;b&gt;&amp;&lt;/b&gt;" || c.replaces != 0 || len(c.actions) != 2 || c.actions[0] != "default" {
		t.Fatalf("first: %+v", c)
	}
	if v, ok := c.hints["sound-name"]; !ok || v.Value() != "message-new-instant" {
		t.Errorf("sound hints: %v", c.hints)
	}
	n.Show(Notification{Title: "Team", Body: "again", Tag: "a/1"})
	n.Show(Notification{Title: "Bob", Body: "hi", Tag: "a/2"})
	calls = server.wait(t, 3)
	if calls[1].replaces != 1 || calls[2].replaces != 0 {
		t.Fatalf("replaces: %d, %d", calls[1].replaces, calls[2].replaces)
	}
	if _, ok := calls[1].hints["suppress-sound"]; !ok {
		t.Errorf("silent hints: %v", calls[1].hints)
	}

	// The replacement has no Open; the first one's is gone with it.
	if err = conn.Emit(busPath, busIface+".ActivationToken", uint32(1), "token-1"); err != nil {
		t.Fatal(err)
	}
	if err = conn.Emit(busPath, busIface+".ActionInvoked", uint32(1), "default"); err != nil {
		t.Fatal(err)
	}
	select {
	case token := <-opened:
		t.Fatal("opened a replaced notification", token)
	case <-time.After(200 * time.Millisecond):
	}
	n.Show(Notification{Title: "Carol", Body: "click", Tag: "a/3", Open: func(token string) { opened <- token }})
	server.wait(t, 4)
	conn.Emit(busPath, busIface+".ActivationToken", uint32(3), "token-3")
	conn.Emit(busPath, busIface+".ActionInvoked", uint32(3), "default")
	select {
	case token := <-opened:
		if token != "token-3" {
			t.Fatal("token", token)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("click not handled")
	}
}
