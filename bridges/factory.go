package bridges

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"

	"github.com/svera/sackson-server/interfaces"
)

var bridges map[string]interface{}

// Error messages returned from bridge factory
const (
	BridgeNotFound = "bridge_not_found"
)

func init() {
	bridges = make(map[string]interface{})
}

// Create returns a new instance of the bridge struct specified
func Create(name string) (interfaces.Bridge, error) {
	if plug, ok := bridges[name]; ok {
		var bridge interfaces.Bridge
		castable := plug.(func() interface{})()
		bridge, ok := castable.(interfaces.Bridge)
		if !ok {
			fmt.Println("Unexpected type from module symbol")
			os.Exit(1)
		}
		return bridge, nil
	}
	return nil, errors.New(BridgeNotFound)
}

func Load() {
	dir := "/usr/lib/sackson"
	files, _ := ioutil.ReadDir(dir)
	if len(files) == 0 {
		fmt.Printf("No files found in %s\n", dir)
		return
	}

	for _, f := range files {
		plug, err := plugin.Open(dir + "/" + f.Name())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := bridgeName(f)
		bridge, err := plug.Lookup("New")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			bridges[name] = bridge
		}
		fmt.Printf("Loaded bridge %s\n", name)
	}
}

func bridgeName(file os.FileInfo) string {
	var extension = filepath.Ext(file.Name())
	return file.Name()[0 : len(file.Name())-len(extension)]
}
