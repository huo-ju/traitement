package api

import (
    "fmt"
	"net/http"
	"github.com/labstack/echo/v4"
    jwt "github.com/dgrijalva/jwt-go"
    "git.a.jhuo.ca/huoju/traitement/pkg/types"
)


func AddUrl(c echo.Context) (err error) {
    c.Echo().Logger.Info("AddUrl")
    user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)

    c.Echo().Logger.Info("api save by user ", name)


    var urlMetaList []types.UrlMeta
    //msg := new(type.UrlMeta)
    if err = c.Bind(&urlMetaList); err != nil {
        return
    }
    fmt.Println(urlMetaList)
    //r, err := database.DBConn.AddURLTask(url string)
    //insertID, err := msg.Save()
    //if err != nil {
    //    c.Echo().Logger.Error("Insert Error", err)
    //}
    //return c.JSON(http.StatusOK, map[string]int64{"result_id": insertID, })
	return c.String(http.StatusOK, "recive url from "+name+"!")

}
