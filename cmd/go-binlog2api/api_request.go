package main

import (
	"encoding/json"

	"github.com/fatih/structs"
	sc "github.com/gr4c2-2000/go-binlog2api/internal/go-binlog2api"
	"github.com/gr4c2-2000/go-lite-http-client/client"
	"github.com/gr4c2-2000/go-lite-http-client/entity"
)

//RequestAPI for API
func request(rqs interface{}, mth string, URL string) error {
	var rM interface{}
	requestStruct := entity.NewHttpRequestStruct()
	requestStruct.SendJson()
	requestStruct.Method = mth
	requestStruct.Expect200()
	requestStruct.SetAddress(URL)

	switch t := rqs.(type) {
	case map[string]interface{}:
		rM = t
	default:
		rM = structs.Map(t)

	}
	jsonBytes, err := json.Marshal(rM)

	if err != nil {
		return err
	}
	requestStruct.Query = jsonBytes

	_, err2 := client.HttpRequestClient(requestStruct)
	if err2 != nil {
		return err2
	}
	return nil
}

//RequestAPI send request
func RequestAPI(rqs interface{}, mth string, URL string) error {
	var err error

	for i := 0; i < 5; i++ {
		err = request(rqs, mth, URL)
		if err == nil {
			return nil
		}
		sc.Logger.Errorln("try:", i+1, err)
	}
	return err
}
