package drivers

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plugin"

	"github.com/svera/sackson-server/api"
)

var drivers map[string]plugin.Symbol

// Error messages returned from driver factory
const (
	DriverNotFound = "driver_not_found"
	DriverNotValid = "driver_not_valid"
)

func init() {
	drivers = make(map[string]plugin.Symbol)
}

// Create returns a new instance of the driver struct specified
func Create(name string) (api.Driver, error) {
	var driver api.Driver
	if name == "test" {
		driver = NewMock()
		return driver, nil
	}
	if plug, ok := drivers[name]; ok {
		if driverConstructor, ok := plug.(func() api.Driver); ok {
			return driverConstructor(), nil
		}
		return nil, errors.New(DriverNotValid)
	}
	return nil, errors.New(DriverNotFound)
}

// Exist return true if a driver with the passed name can be instantiated, false otherwise
func Exist(name string) bool {
	if _, exist := drivers[name]; exist {
		return true
	}
	return false
}

// Load reads all libraries from the game drivers directory and stores them in the drivers map
// if they implement a method called "New"
func Load() {
	dir := "/usr/lib/sackson-server"
	files, _ := ioutil.ReadDir(dir)
	if len(files) == 0 {
		log.Printf("No files found in %s\n", dir)
		return
	}

	for _, f := range files {
		plug, err := plugin.Open(dir + "/" + f.Name())
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		name := driverName(f)
		driver, err := plug.Lookup("New")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		} else {
			drivers[name] = driver
		}
		log.Printf("Loaded driver \"%s\"\n", name)
	}
}

func driverName(file os.FileInfo) string {
	var extension = filepath.Ext(file.Name())
	return file.Name()[0 : len(file.Name())-len(extension)]
}
