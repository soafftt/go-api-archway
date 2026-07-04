package utils

import (
	"bytes"
	json2 "encoding/json"
	"io"
	"net/http"
	"sync"
)

var httpResSyncPool = sync.Pool{
	New: func() any {
		return new(make([]byte, 0, 4096))
	},
}

func ToStruct[T any](httpResponse *http.Response) (T, error) {
	bodyBufferPtr := httpResSyncPool.Get().(*[]byte)
	bodyBuffer := *bodyBufferPtr
	defer httpResSyncPool.Put(bodyBufferPtr)

	readBuffer := bytes.NewBuffer(nil)
	if _, err := io.CopyBuffer(readBuffer, httpResponse.Body, bodyBuffer); err != nil {
		// TODO ERROR 처리.
	}

	var tResult T
	if err := json2.Unmarshal(bodyBuffer, &tResult); err != nil {
		// TOTO Error 처리.
	}

	return tResult, nil
}
