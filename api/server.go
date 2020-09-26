package api

import (
    "fmt"
	"net/http"
	"github.com/labstack/echo/v4"
    jwt "github.com/dgrijalva/jwt-go"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
    "git.a.jhuo.ca/huoju/traitement/pkg/types"
	"git.a.jhuo.ca/huoju/traitement/internal/pkg/task"
)



func AddUrl(amqpQueue *rabbitmq.Queue) echo.HandlerFunc {
    // ... but it returns a echo handler
    return func(c echo.Context) error {
		c.Echo().Logger.Info("AddUrl")
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		name := claims["name"].(string)

		c.Echo().Logger.Info("api save by user ", name)

		var urlMetaList []types.UrlMeta
		//msg := new(type.UrlMeta)
        //fmt.Println("===============")
        //fmt.Println(urlMetaList)
        if err := c.Bind(&urlMetaList); err != nil {
		    //return
            c.String(http.StatusInternalServerError , "error json format")

		}
		fmt.Println(urlMetaList)
		r, err := task.AddURLMetaTasks(urlMetaList, amqpQueue)
        fmt.Println(r)
		if err == nil {
            //atask := &types.Task{ID: uuid.New().String(), Type:"SPIDER", Meta: "{\"url\":\"http://google.com\"}"}
	        //body, err := json.Marshal(atask)
            //amqpQueue.Publish(body)
		    //publish to queue
		}else {
		}
		return c.String(http.StatusOK, "urls recived.")
    }
}
