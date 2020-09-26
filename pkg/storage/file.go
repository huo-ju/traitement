package storage

import (
    "path/filepath"
    "io"
    "os"
    "time"
    "strings"
	"git.a.jhuo.ca/huoju/traitement/pkg/types"
    "fmt"
)


//FileBucket file storage bucket
type FileBucket struct {
    RootPath string
}


//Save save string into a file
func (bucket *FileBucket) Save(pagecontent *types.PageContent, filename string) (string, error) {
    now := time.Now()
    subPath := now.Format("2006-01-02")
    path := filepath.Join(bucket.RootPath, subPath)
    filepath := filepath.Join(bucket.RootPath, subPath, filename)

    _, err := os.Stat(path)
    if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
        return filepath,err
    }

    file, err := os.Create(filepath)
    if err != nil {
        return filepath,err
    }
    defer file.Close()
    content := fmt.Sprintf("Url:%s\nTitle:%s\nAuthor:%s\nPublishTime:%s\nContent:%s", pagecontent.Url, strings.TrimSpace(pagecontent.Title), strings.TrimSpace(pagecontent.Author), pagecontent.PublishTime, strings.TrimSpace(pagecontent.Content))
    _, err = io.WriteString(file, content)
    if err != nil {
        return filepath,err
    }
    file.Sync()
    return filepath, nil
}
