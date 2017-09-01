title: References in Ponzu

References in Ponzu allow you to create relationships between your Content types.
Ponzu uses an embedded database, rather than a more traditional relational database 
with SQL support. This may seem unnatural since there is no native concept of
"foreign keys" or "joins" like you may be used to. Instead, Ponzu wires up your
data using references, which are simply URL paths, like `/api/content?type=Post&id=1`

A foreign key as a URL path?! Am I crazy? No! For the purpose Ponzu serves, 
this structure works quite well, especially given its creation was specifically 
tuned for HTTP/2 features such as "Request/Response Multiplexing" and "Server Push." 

There is a deeper dive into the HTTP/2 concepts [below](/References/Overview/#designed-for-http2), but first we'll walk through
a quick tutorial on Ponzu's references. 

To generate references from the CLI, please [read through the documentation](/CLI/Generating-References). 
The example below assumes you understand the syntax. 

---

### Create Your Content Types

Here we are creating two Content types, `Author` and `Book`. A `Book` will keep
a reference to an `Author` in the sense that an author wrote the book.

```bash
$ ponzu gen c author name:string photo:string:file bio:string:textarea
$ ponzu gen c book title:string author:@author,name pages:int year:int
```

The structs generated for each look like:

`content/author.go`
```go
type Author struct {
	item.Item

	Name  string `json:"name"`
	Photo string `json:"photo"`
	Bio   string `json:"bio"`
}
```

`content/book.go`
```go
type Book struct {
	item.Item

	Title  string `json:"title"`
	Author string `json:"author"`
	Pages  int    `json:"pages"`
	Year   int    `json:"year"`
}
```

Notice how the `Author` field within the `Book` struct is a `string` type, not
an `Author` type. This is because the `Author` is stored as a `string` in our
database, as a reference to the `Author`, instead of embedding the `Author` data 
inside the `Book`. 

Some example JSON data for the two structs looks like:

<kbd>GET</kbd> `/api/content?type=Author&id=1` (`Author`)
```json
{
    "data": [
        {
            "uuid": "024a5797-e064-4ee0-abe3-415cb6d3ed18",
            "id": 1,
            "slug": "item-id-024a5797-e064-4ee0-abe3-415cb6d3ed18",
            "timestamp": 1493926453826,
            "updated": 1493926453826,
            "name": "Shel Silverstein",
            "photo": "/api/uploads/2017/05/shel-silverstein.jpg",
            "bio": "Sheldon Allan Silverstein was an American poet..."
        }
    ]
}
```

<kbd>GET</kbd> `/api/content?type=Book&id=1` (`Book`)
```json
{
    "data": [
        {
            "uuid": "024a5797-e064-4ee0-abe3-415cb6d3ed18",
            "id": 1,
            "slug": "item-id-024a5797-e064-4ee0-abe3-415cb6d3ed18",
            "timestamp": 1493926453826,
            "updated": 1493926453826,
            "title": "The Giving Tree",
            "author": "/api/content?type=Author&id=1",
            "pages": 57,
            "year": 1964
        }
    ]
}
```

As you can see, the `Author` is a reference as the `author` field in the JSON
response for a `Book`. When you're building your client, you need to make a second
request for the `Author`, to the URL path found in the `author` field of the `Book`
response. 

For example, in pseudo-code: 
```bash
# Request 1: 
$book = GET /api/content?type=Book&id=1

# Request 2: 
$author = GET $book.author # where author = /api/content?type=Author&id=1
```

Until recently, this would be considered bad practice and would be costly to do
over HTTP. However, with the wide availability of HTTP/2 clients, including all
modern web browsers, mobile devices, and HTTP/2 libraries in practically every 
programming language, this pattern is fast and scalable. 

---

### Designed For HTTP/2

At this point, you've likely noticed that you're still making two independent 
HTTP requests to your Ponzu server. Further, if there are multiple references or more
than one item, you'll be making many requests -- _how can that be efficient?_ 

There are two main concepts at play: Request/Response Multiplexing and Server Push.

#### Request/Response Multiplexing

With HTTP/2, a client and server (peers) transfer data over a single TCP connection, 
and can send data back and forth at the same time. No longer does a request need
to wait to be sent until after an expected response is read. This means that HTTP 
requests can be sent much faster and at the _same time_ on a single connection. 
Where previously, a client would open up several TCP connections, the re-use of a 
single connection reduces CPU overhead and makes the server more efficient.

This feature is automatically provided to you when using HTTP/2 - the only 
requirement is that you connect via HTTPS and have active TLS certificates, which 
you can get for free by running Ponzu with the `--https` flag and configuring it 
with a properly set, active domain name of your own. 

#### Server Push

Another impactful feature of HTTP/2 is "Server Push": the ability to preemptively
send a response from the server to a client without waiting for a request. This
is where Ponzu's reference design really shows it's power. Let's revisit the
example from above:

```bash
# Request 1: 
$book = GET /api/content?type=Book&id=1

# Request 2: 
$author = GET $book.author # where author = /api/content?type=Author&id=1
```

Instead of waiting for the server to respond with the data for `$book.author`, 
the response data is already in the client's cache before we even make the request!
Now there is no round-trip made to the server and back, and the client reads the 
pushed response from cache in fractions of a millisecond. 

But, how does the server know which response to push and when? You'll need to 
specify which fields of the type you've requested should be pushed. This is done
by implementing the [`item.Pushable` interface](/Interfaces/Item#itempushable). 
See the example below which demonstrates a complete implementation on the `Book`
struct, which has a reference to an `Author`.

##### Example

`content/book.go`
```go
...
type Book struct {
	item.Item

	Title  string `json:"title"`
	Author string `json:"author"`
	Pages  int    `json:"pages"`
	Year   int    `json:"year"`
}


func (b *Book) Push(res http.ResponseWriter, req *http.Request) ([]string, error) {
    return []string{
        // the json struct tag is used to tell the server which
        // field(s) it should push - only URL paths originating
        // from your server can be pushed!
        "author", 
    }, nil
}
...
```

Now, whenever a single `Book` is requested, the server will preemptively push the
`Author` referenced by the book. The response for the `Author` will _already be
on the client_ and will remain there until a request for the referenced `Author` 
has been made.

!!! note "What else can I Push?"
    Only fields that are URL paths originating from your server can be pushed. 
    This means that you could also implement `item.Pushable` on the `Author`
    type, and return `[]string{"photo"}, nil` to push the Author's image!

---

### Other Considerations

HTTP/2 Server Push is a powerful feature, but it can be abused just like anything
else. To try and help mitigate potential issues, Ponzu has put some "stop-gaps"
in place. Server Push is only activated on **single item** API responses, so you
shouldn't expect to see references or files pushed from the `/api/contents` endpoint.
An exception to this is the `/api/search` endpoint, which only the **first** 
result is pushed (if applicable) no matter how many items are in the response. 

You should take advantage of HTTP/2 in Ponzu and get the most out of the system. 
With the automatic HTTPS feature, there is no reason not to and you gain the 
additional benefit of encrypting your traffic - which your users will appreciate!
