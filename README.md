# Go ShortLink - A URL shortening web service demo based on Golang

## Requisites
- Golang
- Redis
- Docker


## API

there are three simple apis

- shorten url
- get shorten url info
- visit short url and redirect

### shorten url

```
API：/api/shorten
METHOD：POST
PARAMS: { "url": "https:www.example.com", "expire_in_minutes": 60 }
```

### get shorten url info

```
API: /api/info/{link}
METHOD: GET
```

### visit short url and redirect

visit link will return status code 307 and redirect to the origin url

```
API: /{link}
METHOD: GET
```


## Reference 
TODO
