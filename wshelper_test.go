package wshelper

import (
	"fmt"
	"testing"
)
func TestConfig(t *testing.T) {
	Assertf(v.Get("MaxOnlineConnPerPool") != nil, t, "no such field 'MaxOnlineConnPerPool', add it in config.yaml")
}

func TestBKBMBGB(t *testing.T) {
	Assertf(B == 1, t, "want 1024 but got %d", B)
	Assertf(KB == 1024, t, "want 1024 but got %d", KB)
	Assertf(MB == 1024*KB, t, "want 1024 but got %d", MB)
	Assertf(GB == 1024*MB, t, "want 1024 but got %d", GB)
}

func TestWebSocketHelper_SetCommands_ListCommands(t *testing.T) {
	const (
		command1 = 1 + iota
		command2
		command3
		command4
	)

	ws := NewWsHelper(nil)
	ws.SetCommands(command1, command2, command3, command4)

	b, e := ws.Marshal(ws.ListCommandsHash())
	if e != nil {
		fmt.Println(e.Error())
		t.Fatal()
	}
	hit := `{"2C9C6216CD8D8CC373FB0FF43A1599F2":1,"999D9957950FA1CE356BE807C0AD6BDA":2,"F0B3957C732669AB5D8644A78DF75230":3,"F37AB071927EF3E23D1E82CDA8D6869E":4}`
	if string(b) != hit {
		t.Fatalf("want '%s' but got %s", hit, b)
	}
	for k, _ := range ws.ListCommandsHash() {
		if len([]byte(k)) != 32 {
			t.Fatalf("want length 32 but got %d", len([]byte(k)))
		}
	}
}

