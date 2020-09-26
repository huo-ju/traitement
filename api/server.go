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

var DenyDomains map[string]int

func AddUrl(amqpQueue *rabbitmq.Queue) echo.HandlerFunc {
    // ... but it returns a echo handler
    return func(c echo.Context) error {
		c.Echo().Logger.Info("AddUrl")
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		name := claims["name"].(string)

		c.Echo().Logger.Info("api save by user ", name)
		c.Echo().Logger.Info(DenyDomains)

		var urlMetaList []types.UrlMeta
        if err := c.Bind(&urlMetaList); err != nil {
		    //return
            c.String(http.StatusInternalServerError , "error json format")

		}
		r, err := task.AddURLMetaTasks(urlMetaList, &DenyDomains, amqpQueue)
        fmt.Println(r)
        fmt.Println(err)
		return c.String(http.StatusOK, "urls recived.")
    }
}
