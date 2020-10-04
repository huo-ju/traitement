package storage

import (
	"net/http"
    "encoding/json"
    "time"
    "bytes"
    "fmt"
	"git.a.jhuo.ca/huoju/traitement/pkg/types"
)

type WebApi struct {
    Endpoint string
    Token string
}

func (webapi *WebApi) SaveUrls(data *[]types.UrlMeta) (error) {
    output, err := json.Marshal(data)
    if err != nil {
        return err
    }
    fmt.Println(err)
    fmt.Println(string(output))
    var httpclient = &http.Client{
        Timeout: time.Second * 60,
    }
    var bearer = "Bearer " + webapi.Token
    apiendpoint := webapi.Endpoint + "/addurl"
    req, err := http.NewRequest("POST", apiendpoint, bytes.NewBuffer(output))
    req.Header.Set("Authorization", bearer)
    req.Header.Add("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	_, httperr := httpclient.Do(req)

	if httperr != nil {
		apierr := fmt.Errorf("WEBAPI result error: %w\n", httperr)
        fmt.Println(apierr)
        return httperr
    } else {
		//bodyBytes, err1 := ioutil.ReadAll(resp.Body)
		//bodyString := string(bodyBytes)
        //fmt.Println(bodyString)
		return nil
    }
}
