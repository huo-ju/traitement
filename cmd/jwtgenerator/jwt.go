package main


import (
	"flag"
    "time"
    "fmt"
	"path/filepath"
	"github.com/spf13/viper"
    jwt "github.com/dgrijalva/jwt-go"
)

var (
    jwtSecret   string
)

func loadconf() {
	viper.AddConfigPath(filepath.Dir("./configs/"))
	viper.AddConfigPath(filepath.Dir("."))
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.ReadInConfig()
	jwtSecret = viper.GetString("JWT_SECRET")
}

func main() {

    var name = flag.String("name", "defaultname", "Jwt user name")
    var isAdmin = flag.Bool("admin", false, "is Admin?")
	flag.Parse()
	loadconf()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = *name
	claims["admin"] = *isAdmin
	claims["exp"] = time.Now().Add(time.Hour * 24*365*100).Unix()
    fmt.Println(claims)
	t, err := token.SignedString([]byte(jwtSecret))
    if err!=nil{
        fmt.Println(err)
    }else {
        fmt.Println("JWT Token:")
        fmt.Println(t)
    }
}
