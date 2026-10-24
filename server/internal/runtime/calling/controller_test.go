package calling

import "testing"

func TestNewControllerRequiresFreeSWITCHClient(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewController(nil) must reject a missing FreeSWITCH client")
		}
	}()
	NewController(nil)
}
