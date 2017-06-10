title: Content HTTP API


Ponzu provides a read & write HTTP API to access and interact with content on a
system. By default, write access (including create, update and delete) and search 
are disabled. See the section on Ponzu's [API Interfaces](/Interfaces/API) to learn
more about how to enable these endpoints.

---

## Endpoints

### Get Content by Type
<kbd>GET</kbd> `/api/content?type=<Type>&id=<ID>`

##### Sample Response
```javascript
{
  "data": [
    {
        "uuid": "024a5797-e064-4ee0-abe3-415cb6d3ed18",
        "id": 6,
        "slug": "item-id-024a5797-e064-4ee0-abe3-415cb6d3ed18" // customizable
        "timestamp": 1493926453826, // milliseconds since Unix epoch
        "updated": 1493926453826,
        // your content data...,
    }
  ]
}
```

---

### Get Contents by Type
<kbd>GET</kbd> `/api/contents?type=<Type>`

  - optional params:
    1. `order` (string: ASC / DESC, default: DESC)
    2. `count` (int: -1 - N, default: 10, -1 returns all)
    3. `offset` (int: 0 - N, default: 0)
##### Sample Response
```javascript
{
  "data": [
    {
        "uuid": "024a5797-e064-4ee0-abe3-415cb6d3ed18",
        "id": 6,
        "slug": "item-id-024a5797-e064-4ee0-abe3-415cb6d3ed18", // customizable
        "timestamp": 1493926453826, // milliseconds since Unix epoch
        "updated": 1493926453826,
        // your content data...,
    },
    {
        "uuid": "5a9177c7-634d-4fb1-88a6-ef6c45de797c",
        "id": 7,
        "slug": "item-id-5a9177c7-634d-4fb1-88a6-ef6c45de797c", // customizable
        "timestamp": 1493926453826, // milliseconds since Unix epoch
        "updated": 1493926453826,
        // your content data...,
    },
    // more objects...
  ]
}
```

---

### Get Content by Slug
<kbd>GET</kbd> `/api/content?slug=<Slug>`

##### Sample Response
```javascript
{
  "data": [
    {
        "uuid": "024a5797-e064-4ee0-abe3-415cb6d3ed18",
        "id": 6,
        "slug": "item-id-024a5797-e064-4ee0-abe3-415cb6d3ed18", // customizable
        "timestamp": 1493926453826, // milliseconds since Unix epoch
        "updated": 1493926453826,
        // your content data...,
    }
  ]
}
```

---

### New Content
<kbd>POST</kbd> `/api/content/create?type=<Type>`

  - Type must implement [`api.Createable`](/Interfaces/API#apicreateable) interface
!!! note "Request Data Encoding" 
    Request must be `multipart/form-data` encoded. If not, a `400 Bad Request` 
    Response will be returned.

##### Sample Response
```javascript
{
  "data": [
    {
        "id": 6, // will be omitted if status is pending
        "type": "Review",
        "status": "public"
    }
  ]
}
```

---

### Update Content
<kbd>POST</kbd> `/api/content/update?type=<Type>&id=<id>`

  - Type must implement [`api.Updateable`](/Interfaces/API#apiupdateable) interface
!!! note "Request Data Encoding" 
    Request must be `multipart/form-data` encoded. If not, a `400 Bad Request` 
    Response will be returned.
  
##### Sample Response
```javascript
{
  "data": [
    {
        "id": 6,
        "type": "Review",
        "status": "public"
    }
  ]
}
```

---

### Delete Content
<kbd>POST</kbd> `/api/content/delete?type=<Type>&id=<id>`

  - Type must implement [`api.Deleteable`](/Interfaces/API#apideleteable) interface
!!! note "Request Data Encoding" 
    Request must be `multipart/form-data` encoded. If not, a `400 Bad Request` 
    Response will be returned.

##### Sample Response
```javascript
{
  "data": [
    {
        "id": 6,
        "type": "Review",
        "status": "deleted"
    }
  ]
}
```

---

### Additional Information

All API endpoints are CORS-enabled (can be disabled in configuration at run-time) and API requests are recorded by your system to generate graphs of total requests and unique client requests within the Admin dashboard.

#### Response Headers
The following headers are common across all Ponzu API responses. Some of them can be modified
in the [system configuration](/System-Configuration/Settings) while your system is running.

##### HTTP/1.1
```
HTTP/1.1 200 OK
Access-Control-Allow-Headers: Accept, Authorization, Content-Type
Access-Control-Allow-Origin: *
Cache-Control: max-age=2592000, public
Content-Encoding: gzip
Content-Type: application/json
Etag: MTQ5Mzk0NTYzNQ==
Vary: Accept-Encoding
Date: Fri, 05 May 2017 01:15:49 GMT
Content-Length: 199
```

##### HTTP/2
```
access-control-allow-headers: Accept, Authorization, Content-Type
access-control-allow-origin: *
cache-control: max-age=2592000, public
content-encoding: gzip
content-length: 199
content-type: application/json
date: Fri, 05 May 2017 01:38:11 GMT
etag: MTQ5Mzk0ODI4MA==
status: 200
vary: Accept-Encoding
```

#### Helpful links
[Typewriter](https://github.com/natdm/typewriter)
Generate & sync front-end data structures from Ponzu content types. ([Ponzu example](https://github.com/natdm/typewriter/blob/master/EXAMPLES.md#example-use-in-a-package-like-ponzu))
