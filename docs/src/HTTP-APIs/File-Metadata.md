title: File Metadata HTTP API

Ponzu provides a read-only HTTP API to get metadata about the files that have been uploaded to your system. As a security and bandwidth abuse precaution, the API is only queryable by "slug" which is the normalized filename of the uploaded file. 

---

### Endpoints

#### Get File by Slug (single item)
<kbd>GET</kbd> `/api/uploads?slug=<Slug>`

##### Sample Response
```javascript
{
  "data": [
    {
        "uuid": "024a5797-e064-4ee0-abe3-415cb6d3ed18",
        "id": 6,
        "slug": "filename.jpg",
        "timestamp": 1493926453826, // milliseconds since Unix epoch
        "updated": 1493926453826,
        "name": "filename.jpg",
        "path": "/api/uploads/2017/05/filename.jpg",
        "content_length": 357557,
        "content_type": "image/jpeg",
    }
  ]
}
```