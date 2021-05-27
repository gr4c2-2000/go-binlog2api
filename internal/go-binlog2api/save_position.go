package scripts

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

//Position current position of binlog
type Position struct {
	BinlogPos      uint32
	BinlogFilename string
}

var positionPath = "../configs/dbPossition.txt"

//Save current position to file
func (p *Position) Save() {
	b, err := json.Marshal(p)
	if err != nil {
		Logger.Fatalln("Problem with struct to JSON on saving position")
		return
	}

	f, err := os.OpenFile(positionPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		Logger.Fatalln(err)
	}
	defer f.Close()

	f.Write(b)
}

func (p *Position) Read() error {
	jsonFile, err := os.Open(positionPath)
	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	json.Unmarshal(byteValue, &p)
	if p.BinlogFilename == "" {
		Logger.Info("Bo Data in File")
		return errors.New("No Data in File")
	}
	return nil
}
