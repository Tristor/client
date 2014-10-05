
package libkb

import (
	"os"
	"fmt"
	"github.com/okcupid/jsonw"
	"encoding/json"
)

type JsonFile struct {
	filename string
	which string
	jw *jsonw.Wrapper
}

func (f *JsonFile) Load(warnOnNotFound bool) error {
	G.Log.Debug(fmt.Sprintf("+ loading %s file: %s", f.which, f.filename))
	file, err := os.Open(f.filename)
	if err != nil {
		if os.IsNotExist(err) {
			if warnOnNotFound {
				G.Log.Warning(fmt.Sprintf("No %s file found; tried %s", f.which, f.filename))
			}
			return nil
		} else if os.IsPermission(err) {
			G.Log.Warning(fmt.Sprintf("Permission denied opening %s file %s", f.which, f.filename))
			return nil
		} else {
			return err
		}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	obj := make(map[string]interface{})
	err = decoder.Decode(&obj)
	if err != nil {
		G.Log.Error(fmt.Sprintf("Error decoding %s file %s", f.which, f.filename))
		return err
	}
	f.jw = jsonw.NewWrapper(obj)
	G.Log.Debug(fmt.Sprintf("- successfully loaded %s file", f.which))
	return nil
}

func (f *JsonFile) Save(mode os.FileMode) (err error) {
	G.Log.Debug(fmt.Sprintf("+ saving %s file %s", f.which, f.filename))

	err = MakeParentDirs(f.filename)
	if err != nil {
		G.Log.Error("Failed to make parent dirs for %s", f.filename)
		return err
	}

	var dat interface{}

	if f.jw == nil {
		// Make a default Dictionary if none already exists
		dat = make(map[string]interface{})
		G.Log.Warning("No value for %s file; assuming empty value (i.e., {})", f.which)
	} else {
		dat, err = f.jw.GetData()
		if err != nil {
			G.Log.Error("Failed to encode data for %s file", f.which)
			return err
		}
	}
	var writer *os.File
	flags := (os.O_WRONLY | os.O_CREATE | os.O_TRUNC)
	if mode == 0 {
		mode = PERM_FILE // By default, secrecy
	}
	writer, err = os.OpenFile(f.filename, flags, mode)
	if err != nil {
		G.Log.Error("Failed to open %s file %s for writing: %s", f.which, f.filename, err.Error())
		return err
	}
	defer writer.Close()

	encoder := json.NewEncoder(writer)
	err = encoder.Encode(dat)
	if err != nil {
		G.Log.Error("Error encoding data to %s file %s: %s", f.which, f.filename, err.Error())
		return err
	}
	err = writer.Close()
	if err != nil {
		G.Log.Error("Error flushing %s file %s: %s", f.which, f.filename, err.Error())
		return err
	}

	G.Log.Debug(fmt.Sprintf("- saved %s file %s", f.which, f.filename))
	return
}
