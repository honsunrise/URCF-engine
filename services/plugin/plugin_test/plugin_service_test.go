package plugin_test

import (
	"github.com/zhsyourai/URCF-engine/repositories/account"
	"fmt"
	"math/rand"
	"testing"
	"github.com/zhsyourai/URCF-engine/services/plugin/protocol"
)

var testID = "__test" + fmt.Sprint(rand.Int())
var testPassword = "password" + fmt.Sprint(rand.Int())
var repo = account.NewAccountRepository()

func TestPluginService(t *testing.T) {
	stub := protocol.NewPluginStub()
}