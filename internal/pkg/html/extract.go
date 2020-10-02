package html

import (
    "fmt"
    "strings"
    "net/url"
    "github.com/antchfx/htmlquery"
	"github.com/virushuo/textractor"
	"git.a.jhuo.ca/huoju/traitement/pkg/types"
)


func FindLink(siteprefix string, html string) []types.UrlMeta {
    minpathlen := 8

    doc, err := htmlquery.Parse(strings.NewReader(html))
    fmt.Println(err)
    list := htmlquery.Find(doc, "//a/@href")

    urlmeta_list := make([]types.UrlMeta,len(list))
    idx := 0
    for _ , n := range list{
        href := htmlquery.SelectAttr(n, "href")
        var link string
        if strings.HasPrefix(href, "/") { //link without domain
            if len(href) >= minpathlen {
                link = fmt.Sprintf("%s%s",siteprefix, href)
            }
        } else if strings.HasPrefix(strings.ToLower(href), "http:") || strings.HasPrefix(strings.ToLower(href), "https:"){
            u, err := url.Parse(href)
            if err == nil {
                if len(href)-len(u.Scheme)-2-len(u.Host) >= minpathlen {
                    link = href
                }
            }

        }
        if len(link)>0{
            urlmeta_list[idx] = types.UrlMeta{Url:strings.TrimSpace(link), GatherLink:false, Uniq:true, SavePage:true}
            idx++
        }
    }
    return urlmeta_list[:idx]
}

func FindContent(url string, html string) (*types.PageContent, error ){
    var err error
    defer func() {
        if r := recover(); r != nil {
            err = r.(error)
        }
    }()

    r, err := textractor.Extract(html)
    if err == nil {
        pagecontent := &types.PageContent{url, r.Title, r.Author, r.PublishTime , r.Content}
        return pagecontent,nil
    }
    return nil, err
}
