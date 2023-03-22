package main

type Storage interface {
	Shorten(string, int64) (string, error)
	UnShorten(string) (string, error)
	//ShortLinkInfo(string2 string) (URLDetail, error)
}
