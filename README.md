## Simple file storage service demo
Stores given files in gorm handled db (currently sqlite) with memory caching and API for 
retrieving files metadata, one file at a time.

## Running
To run default server <br>
```
  go run .
```

To run all tests <br>
```
  go test -v ./...
```

### Access main page UI to download or list files:
`http://localhost:1323/` <br>
There is a screenshot of it at `./screenshot.png`


## API docs
### Upload file using post method with data base64 encoded:
```
curl -u user:pass -X POST localhost:1323/api/v1/file -d  '{
    "name": "file.txt",
    "uploaded_by": "File Owner",
    "email": "test@test.com",
    "data": "Zm9vIGJhciB0ZXN0Cg=="
  }'

```

### Alternatively upload using multipart form:
```
curl -u user:pass -F name="Test User" email=testuesr@gmail.com -F file=@file.txt localhost:1323/upload
```

### List files meta-data:
```
curl -u user:pass -X GET localhost:1323/api/v1/file
```
### Sample response:
```
[
  {
    "ID": 1,
    "CreatedAt": "0001-01-01T00:00:00Z",
    "UpdatedAt": "0001-01-01T00:00:00Z",
    "DeletedAt": null,
    "name": "file.txt",
    "uploaded_by": "TestName",
    "email": "test@test.com"
  },
  {
    "ID": 2,
    "CreatedAt": "0001-01-01T00:00:00Z",
    "UpdatedAt": "0001-01-01T00:00:00Z",
    "DeletedAt": null,
    "name": "file.txt",
    "uploaded_by": "Some User",
    "email": "test@test.com"
  }
]

```

### Download file by id
This will return given file if found as blob application/octet-stream.
```
curl -X GET -u user:pass localhost:1323/api/v1/file/download/1 
```
