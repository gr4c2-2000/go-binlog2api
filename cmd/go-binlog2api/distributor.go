package main

import (
	"strings"
	"sync"

	sc "github.com/gr4c2-2000/go-binlog2api/internal/go-binlog2api"
	"github.com/siddontang/go-mysql/canal"
)

const from = "FROM"
const to = "TO"

//Distribute comperse with config and send data to proper source
func Distribute(e *canal.RowsEvent) error {

	RConfigs := sc.RConfig()
	var wg sync.WaitGroup
	for _, c := range RConfigs.RConfig {
		if c.Db != e.Table.Schema || c.Table != e.Table.Name {
			continue
		}

		switch e.Action {
		case "update":
			if configSet(c.ReciverAPI.UpdateURL) == true && send(e, c.OnField) == true {
				wg.Add(1)
				go func(c sc.ReciverConfig) {
					defer wg.Done()
					sendToAPI(e, c.ReciverAPI.Critical, c.ReciverAPI.Format, c.ReciverAPI.UpdateURL, c.ReciverAPI.UpdateMethod)
				}(c)
			}
			if configSet(c.ReciverMQ.UpdateQueueProducer) == true && send(e, c.OnField) == true {
				wg.Add(1)
				go func(c sc.ReciverConfig) {
					defer wg.Done()
					sendToQueue(e, c.ReciverMQ.Critical, c.ReciverMQ.Format, c.ReciverMQ.UpdateQueueProducer)
				}(c)
			}
		case "insert":
			if configSet(c.ReciverAPI.InsertURL) == true {
				wg.Add(1)
				go func(c sc.ReciverConfig) {
					defer wg.Done()
					sendToAPI(e, c.ReciverAPI.Critical, c.ReciverAPI.Format, c.ReciverAPI.InsertURL, c.ReciverAPI.InsertMethod)
				}(c)

			}
			if configSet(c.ReciverMQ.InsertQueueProducer) == true {
				wg.Add(1)
				go func(c sc.ReciverConfig) {
					defer wg.Done()
					sendToQueue(e, c.ReciverMQ.Critical, c.ReciverMQ.Format, c.ReciverMQ.InsertQueueProducer)
				}(c)
			}
		case "delete":
			if configSet(c.ReciverAPI.DeleteURL) == true {
				wg.Add(1)
				go func(c sc.ReciverConfig) {
					defer wg.Done()
					sendToAPI(e, c.ReciverAPI.Critical, c.ReciverAPI.Format, c.ReciverAPI.DeleteURL, c.ReciverAPI.DeleteMethod)
				}(c)
			}
			if configSet(c.ReciverMQ.DeleteQueueProducer) == true {
				wg.Add(1)
				go func(c sc.ReciverConfig) {
					defer wg.Done()
					sendToQueue(e, c.ReciverMQ.Critical, c.ReciverMQ.Format, c.ReciverMQ.DeleteQueueProducer)
				}(c)
			}
		default:
			sc.Logger.Error(e, "unsupported Type")
		}
	}

	wg.Wait()
	return nil

}

func rowToMap(e *canal.RowsEvent) map[int]map[string]interface{} {
	mp := make(map[int]map[string]interface{})

	for inx, row := range e.Rows {
		rmp := make(map[string]interface{})
		for in, cel := range row {
			rmp[e.Table.Columns[in].Name] = cel
		}

		mp[inx] = rmp
	}
	return mp
}

func rowToMapForUpdate(e *canal.RowsEvent) map[int]map[string]map[string]interface{} {

	mp := make(map[int]map[string]map[string]interface{})
	var group map[string]map[string]interface{}

	for inx, row := range e.Rows {
		var action string
		ind := (inx - (inx % 2)) / 2
		rmp := make(map[string]interface{})
		if (inx % 2) == 0 {
			action = from
			group = make(map[string]map[string]interface{})
			mp[ind] = group
		} else {
			action = to
		}
		for in, cel := range row {
			rmp[e.Table.Columns[in].Name] = cel
		}
		group[action] = rmp
	}
	return mp
}

func configSet(s string) bool {
	if strings.TrimSpace(s) != "" {
		return true
	}
	return false
}

func sendToAPI(e *canal.RowsEvent, critical bool, format string, URL string, method string) {
	var rQ interface{}
	if format == "RAW" {
		rQ = e
	} else {
		rQ = createJSONRequest(e)
	}
	err := RequestAPI(rQ, method, URL)
	if err != nil && critical == true {
		sc.Logger.Fatalln("Error in critical reciver ", " ", URL)
		panic("Error in critical reciver")

	} else if err != nil {
		sc.Logger.Error("Error in reciver ", " ", URL)
	}
	return

}
func sendToQueue(e *canal.RowsEvent, critical bool, format string, queProd string) {
	var rQ interface{}
	if format == "RAW" {
		rQ = e
	} else {
		rQ = createJSONRequest(e)
	}
	err := Produce(queProd, rQ)
	if err != nil && critical == true {
		sc.Logger.Fatalln("Error in critical reciver ", " ", queProd)
		panic("Error in critical reciver")
	} else if err != nil {
		sc.Logger.Error("Error in reciver ", " ", queProd)
	}
	return

}
func comper(m map[int]map[string]map[string]interface{}, f string) bool {
	for _, val := range m {
		if val["FROM"][f] != val["TO"][f] {
			return true
		}
	}
	return false
}

func send(e *canal.RowsEvent, of []string) bool {
	if len(of) != 0 {
		ump := rowToMapForUpdate(e)
		for _, fl := range of {
			if comper(ump, fl) == true {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

func createJSONRequest(e *canal.RowsEvent) interface{} {
	req := make(map[string]interface{})
	req["table"] = e.Table.Name
	req["db"] = e.Table.Schema
	if e.Action == "update" {
		req["data"] = rowToMapForUpdate(e)
	} else {
		req["data"] = rowToMap(e)
	}
	req["logPos"] = e.Header.LogPos
	req["timestamp"] = e.Header.Timestamp
	req["serverId"] = e.Header.ServerID
	req["flags"] = e.Header.Flags

	return req
}
