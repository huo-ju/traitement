package storage

import (
    "path/filepath"
    "io"
    "os"
    "time"
    "fmt"
)


//FileBucket file storage bucket
type FileBucket struct {
    RootPath string
}

//Save save string into a file
func (bucket *FileBucket) Save(content string, filename string) (string, error) {
    now := time.Now()
    subPath := now.Format("2006-02-01")
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

    _, err = io.WriteString(file, content)
    if err != nil {
        return filepath,err
    }
    file.Sync()
    return filepath, nil
}
