package bridges

/*
import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/svera/sackson-server/interfaces"
)

var bridges map[string]interfaces.Bridge

// Error messages returned from bridge factory
const (
	BridgeNotFound = "bridge_not_found"
)

func init() {
	bridges = make(map[string]interfaces.Bridge)
}

// Create returns a new instance of the bridge struct specified
func Create(name string) (interfaces.Bridge, error) {
	if bridge, ok := bridges[name]; ok {
		return bridge, nil
	}
	return nil, errors.New(BridgeNotFound)
}

func Load() {
	dir := "/usr/lib/sackson"
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		plug, err := plugin.Open(dir + "/" + f.Name())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := bridgeName(f)
		bridgeSymbol, err := plug.Lookup(name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var bridge interfaces.Bridge
		bridge, ok := bridgeSymbol.(interfaces.Bridge)
		bridges[name] = bridge
		if !ok {
			fmt.Println("unexpected type from module symbol")
			os.Exit(1)
		}
	}
}

func bridgeName(file os.FileInfo) string {
	var extension = filepath.Ext(file.Name())
	var name = file.Name()[0 : len(file.Name())-len(extension)]
	return strings.Title(name)
}
*/
