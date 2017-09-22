package drivers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"

	"github.com/svera/sackson-server/interfaces"
	"github.com/svera/sackson-server/mocks"
)

var drivers map[string]interface{}

// Error messages returned from driver factory
const (
	DriverNotFound = "driver_not_found"
	DriverNotValid = "driver_not_valid"
)

func init() {
	drivers = make(map[string]interface{})
}

// Create returns a new instance of the driver struct specified
func Create(name string) (interfaces.Driver, error) {
	var driver interfaces.Driver
	if name == "test" {
		driver = &mocks.Driver{}
		return driver, nil
	}
	if plug, ok := drivers[name]; ok {
		castable := plug.(func() interface{})()
		driver, ok := castable.(interfaces.Driver)
		if !ok {
			fmt.Printf("Module \"%s\" does not implement Driver interface\n", name)
			return nil, errors.New(DriverNotValid)
		}
		return driver, nil
	}
	return nil, errors.New(DriverNotFound)
}

func Load() {
	dir := "/usr/lib/sackson-server"
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

		name := driverName(f)
		driver, err := plug.Lookup("New")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			drivers[name] = driver
		}
		fmt.Printf("Loaded driver \"%s\"\n", name)
	}
}

func driverName(file os.FileInfo) string {
	var extension = filepath.Ext(file.Name())
	return file.Name()[0 : len(file.Name())-len(extension)]
}
