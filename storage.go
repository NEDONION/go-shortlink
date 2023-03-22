package main

type Storage interface {
	Shorten(url string, expiration int64) (string, error)
	UnShorten(encodeId string) (string, error)
	ShortLinkInfo(encodeId string) (ShortLinkInfo, error)
}
