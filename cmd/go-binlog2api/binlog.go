package main

import (
	"runtime/debug"

	sc "github.com/gr4c2-2000/go-binlog2api/internal/go-binlog2api"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
)

type binlogHandler struct {
	canal.DummyEventHandler
}

var i int = 0
var Position = sc.Position{}

func (h *binlogHandler) OnRow(e *canal.RowsEvent) error {
	defer func() {
		if r := recover(); r != nil {
			println(string(debug.Stack()))
			sc.Logger.Info()
			sc.Logger.Info(r, "  ", string(debug.Stack()))
		}
	}()
	err := Distribute(e)
	if err == nil {
		saveCoords()
	} else {
		return err
	}

	return nil
}

//OnXID hope to see binlog position
func (h *binlogHandler) OnXID(nextPos mysql.Position) error {
	Position.BinlogPos = nextPos.Pos
	Position.BinlogFilename = nextPos.Name
	return nil
}

func (h *binlogHandler) String() string {
	return "binlogHandler"
}

func binlogListener() {
	sc.Logger.Info("Starting...")
	c, err := getDefaultCanal()
	var coords mysql.Position
	if err == nil {
		err := sc.GetBinlogPosition(&coords)
		if err != nil {
			coords, err := c.GetMasterPos()
			if err == nil {
				c.SetEventHandler(&binlogHandler{})
				err = c.RunFrom(coords)
				if err != nil {
					sc.Logger.Fatalln(err)
				}
			}
		}
		c.SetEventHandler(&binlogHandler{})
		sc.Logger.Info("Start with:", coords)
		err = c.RunFrom(coords)
		if err != nil {
			sc.Logger.Fatalln(err)

		}
	}

}

func getDefaultCanal() (*canal.Canal, error) {
	cfg := canal.NewDefaultConfig()
	err := sc.GetDatabaseConfig(cfg)
	sc.Logger.Info("DbConfig:", cfg)
	if err != nil {
		sc.Logger.Fatalln(err)
		panic("Error On set Place in binlog")
	}
	return canal.NewCanal(cfg)
}

func saveCoords() {
	if i > 100 {
		Position.Save()
		i = 0
	}
	i++
}
