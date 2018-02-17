package plugin_test

import (
	"github.com/zhsyourai/URCF-engine/repositories/account"
	"fmt"
	"math/rand"
)

var testID = "__test" + fmt.Sprint(rand.Int())
var testPassword = "password" + fmt.Sprint(rand.Int())
var repo = account.NewAccountRepository()
