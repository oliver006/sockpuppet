# SockPuppet
#### Having fun with WebSockets, Python, Golang and nytimes.com <br>
<img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> <img src ="http://upload.wikimedia.org/wikipedia/commons/a/a7/Sock-puppet.jpg" height="50px"> 


<br>
### What's this all about? 
Did you ever wonder how **nytimes.com** pushes breaking news articles to the front page while you have it open in your browser? Well, I used my browser's developer tools to look at what's going one and it turns out, they don't periodically reload JSON data but use websockets to push new events directly to your browser ([see here](https://developer.mozilla.org/en-US/docs/WebSockets) for more information about websockets).<br>
It's a system called `nyt-fabrik`, here are a few talks and presentations where they give some insight into the architecture: [search google for "nytimes fabrik websockets"](https://www.google.com/search?q=nytimes+fabrik+websockets). 

There is example code, see [here for the Python code](sockpuppet.py) and [here for the Golang example](sockpuppet.go).

<br>
### Cool, so how does it work?

When you go to **nytimes.com**, your browser will establish a websocket connection with the NYT fabrik server and, after a little login dance, will start listening for news events.
Your browser opens a websocket TCP connection to e.g. `ws://blablabla.fabrik.nytimes.com./123/abcde123/websocket` and the server sends a one-character frame `o` which is a request to provide some sort of login identification.<br>
The client (your browser) responds with `["{\"action\":\"login\",\"client_app\":\"hermes.push\",\"cookies\":{\"nyt-s\":\"SOME_COOKIE_VALUE_HERE\"}}"]` and next thing you know you, you either receive a `h` every 20-30 seconds which is some sort of keep-alive or a frame that starts with `a` and has all sorts of data encoded as JSON.

If we receive a message starting with `a`, we can strip the first character and JSON decode the rest. 

```json
{
    "body": "{\"status\":\"updated\",\"version\":1,\"links\":[{\"url\":\"http://www.nytimes.com/2015/05/26/us/cleveland-police.html\",\"count\":0,\"content_id\":\"100000003702598\",\"content_type\":\"article\",\"offset\":0}],\"title\":\"Cleveland Is Said to Settle Justice Department Lawsuit Over Policing\",\"start_time\":1432581057,\"display_duration\":null,\"label\":\"Breaking News\",\"last_modified\":1432581057,\"display_type_id\":1,\"end_time\":1432581057,\"id\":34931339,\"sub_type\":\"BreakingNews\"}",
    "timestamp": "2015-05-21T11:21:11.123456Z",
    "hash_key": "34131339",
    "uuid": "1234",
	...
    "account": "nyt1",
    "type": "feeds_item"
}
```

If the decoded message has field "body", we can decode it. In case of a breaking news item it looks something like this: 

```json
{"status": "updated", "sub_type": "BreakingNews", 
"links": [{"url": "http://www.nytimes.com/2015/05/26/us/cleveland-police.html", "count": 0, "content_id": "100000003702598", "content_type": "article", "offset": 0}], 
"title": "Cleveland Is Said to Settle Justice Department Lawsuit Over Policing", 
"start_time": 1432581057, "display_duration": null, "label": "Breaking News",
"version": 1, "display_type_id": 1, "end_time": 1432581057, 
"last_modified": 1432581057, "id": 34131339}
```
<br>
### Neat but how do I access the feed programmatically?

Good question, let's see, we need about 3-4 things to get this to work, easy. For the Python example, I'll be using the [Tornado websocket framework](http://tornado.readthedocs.org/en/latest/websocket.html) and for the Golang example I'll be using the [Golang.org websocket package](https://godoc.org/golang.org/x/net/websocket).

#### Connect to the websocket

In Python, this is easy:

```python
url = "ws://blablabla.fabrik.nytimes.com./123/abcdef123/websocket"
try:
    w = yield tornado.websocket.websocket_connect(url, connect_timeout=5)
    logging.info("Connected to %s", url)
except Exception as ex:
    logging.error("couldn't connect, err: %s", ex)
``` 

In Golang, it looks about the same:

```go
addr := "ws://blablabla.fabrik.nytimes.com./123/abcdef123/websocket"
ws, err := websocket.Dial(addr, "", "http://www.nytimes.com/")
if err != nil {
	log.Fatal(err)
}
log.Printf("Connected to %s", addr)
```
That was easy, wasn't it?

#### Listen for incoming messages 
Good, we now are connected and have a websocket object/struct we can work with, let's listen for incoming messages.<br>

Python:

```python
while True:
    payload = yield w.read_message()
    if payload is None:
        logging.error("uh oh, we got disconnected")
        return
```
and in Golang:

```go
var msgBuf = make([]byte, 4096)
for {
	bufLen, err := ws.Read(msgBuf)
	if err != nil {
		log.Printf("read err: %s", err)
		return
	}
```
One caveat here, the Golang version can't handle messages longer than 4k (it'll chunk them into 4k pieces) but for our purposes that's not an issue.

#### Send the login message 

If we receive `o` we need to send the login message. We need a cookie value so let's make one up:

```python
if payload[0] == "o":
    cookie = ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(32))
    msg = json.dumps(['{"action":"login", "client_app":"hermes.push", "cookies":{"nyt-s":"%s"}}' % cookie])
    w.write_message(msg.encode('utf8'))
    logging.info("sent cookie: %s", cookie)
```

In Golang this is a bit more verbose:

```go
if msgBuf[0] == 'o' {
	// reply to the login request
	cookie := randCookie()
	msg := fmt.Sprintf(`["{\"action\":\"login\", \"client_app\":\"hermes.push\", \"cookies\":{\"nyt-s\":\"%s\"}}"]`, cookie)
	_, err := ws.Write([]byte(msg))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Sent cookie: %s\n", cookie)
}
```
and `randCookie()` lookslike this:

```go
func randCookie() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, 30)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
```

#### Patiently wait; and (mostly) ignore the `h` messages
Nothing much to do here, whenever we get a `h` message we can simply write `ping` to the console.

```python
elif payload[0] == 'h':
    logging.info('ping')
```
and

```go
if payload[0] == "o" {
	log.Println("ping")
}
```
 

#### Decode the news alert message when we receive one

Messages from the server that start with `a` contain JSON encoded data that we can decode. 
Python first:

```go
elif payload[0] == 'a':
    frame = json.loads(payload[1:])
	if 'body' in frame:
	    body = json.loads(frame['body'])
```	
Now you can for check `if body['sub_type'] == "BreakingNews"` or whatever else you plan on doing with this.

In Golang everything is a bit more verbose but roughly works the same (inlined and shortened for brevity).

```python
if payload[0] == "o" {

	frame := []struct {
		UUID        string `json:"uuid"`
		Product     string `json:"product"`
		Project     string `json:"project"`
		...
		Body        string `json:"body,omitempty"`
	}{}

	// [1:] as we want to skip the leading character `a`
	err = json.Unmarshal(payload[1:], &frame)
	if err != nil {
		return
	}
	if len(frame.Body) > 1 {
		// here we should try to JSON unmarshal frame.Body
	}
}

```
`frame.Body` can now be unmarshaled in the same way as `payload[1:]` earlier.
The resulting struct for it looks something like this:

```go
type MessageBody struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Version      int    `json:"version"`
	SubType      string `json:"sub_type"`
	Label        string `json:"label"`
	StartTime    int    `json:"start_time"`
	EndTime      int    `json:"end_time"`
	LastModified int    `json:"last_modified"`
	Links []struct {
		URL         string `json:"url"`
		ContentID   string `json:"content_id"`
	} `json:"links"`
}

``` 

<br>
### Sweet but what do I do with this?

Totally up to you. Send yourself an email or txt msg using Twilio or Plivo every time something happens. 


### Cool, how do I run the examples?

Python

```
python sockpuppet.py --ws_addr="ws://<<ADDRESS HERE>>"
```

Go

```
go run sockpuppet.go --ws_addr="ws://<<ADDRESS HERE>>"
```

You can find a valid websocket host by using the Developer Console of your favorite browser and visit [nytimes.com](nytimes.com) and look for websocket connections in the network tab.





