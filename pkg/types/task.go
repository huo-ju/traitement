package types

type UrlMeta struct {
    Url string `json:"Url"`
    GatherLink bool `json:"GatherLink"`
    Uniq bool `json:"Uniq"`
    SavePage bool `json:"SavePage"`
}

type Task struct {
    ID string
    Type string
    Meta string
}
