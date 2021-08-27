# Native Go Zookeeper Client Library

This library enables users to communicate with Zookeeper servers over the Zookeeper wire protocol and provides an API for easy usage. It leverages the [Jute compiler](https://github.com/go-zookeeper/jute) for simple serialization and deserialization of protocol messages.

Message formats are defined in a [Jute file](https://github.com/apache/zookeeper/blob/master/zookeeper-jute/src/main/resources/zookeeper.jute) which can be compiled to Go code similarly to Thrift or Protobuf. 
The Jute compiler can be ran from the `internal` directory by running `go generate`, which will run the compiler against the `zookeeper.jute` definition file.

The project is still in active development and therefore may not be ready for production use.

## Setup

The recommended procedure is to download the `zk` client via `go get`:

```
`go get -u github.com/facebookincubator/zk`
```

After successfully running this command, the library will be available in your `GOPATH`, and you can start using it to communicate with your Zookeeper infrastructure.

## Usage

The default way library users can communicate with a Zookeeper server is by using the `Client` abstraction and its methods. It provides a functionality of retryable calls as well as additional configuration parameters.
Upon reconnecting, the client simply creates a new session instead of attempting to reuse the old `sessionID` if one was generated.

It is also possible to use a raw connection via `DialContext` for more fine-tuned control. This call returns a `Conn` instance which can be used for manual RPCs, and does not offer any additional functionalities such as reconnects.

Connections are kept alive by the client and should be closed after usage by calling `Client.Reset()` or `Conn.Close()` depending on the API.

### Default Client

```
client := &Client{
    Network:    "tcp",
    Ensemble:   "127.0.0.1:2181",
}
defer client.Reset()

data, err := client.GetData(context.Background(), "/")
log.Println(string(data))

children, err := client.GetChildren(context.Background(), "/")
log.Println(children)

```

### Retryable client

When connection problems or timeouts are encountered, the client will try to re-establish the connection and retry the operation. Some errors are non-retryable, for example if the znode specified does not exist.

```
client := &Client{
    Network:        "tcp",
    Ensemble:       "127.0.0.1:2181",
    MaxRetries:     5,
    SessionTimeout: time.Second,
}
defer client.Reset()

data, err := client.GetData(context.Background(), "/")
log.Println(string(data))
```

### Custom dialers

Should library users require custom discovery mechanisms, for example for connecting to multiple nodes, they can add a custom `Dialer` to the Client.

```
client := &Client{
    Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
        // custom logic here
    }
}
```

### Using lower-level primitives

Client users can get access to the lower-level `Conn` primitive by Calling `zk.DialContext`. The connection’s lifetime can then be handled by passing a `Context` to the call.

```
conn, err := DialContext(context.Background(), "tcp", "127.0.0.1:2181")
if err != nil {
    log.Println("unexpected error: ", err)
}
defer conn.Close()

data, err := conn.GetData("/")
if err != nil {
    log.Println("unexpected error calling GetData: ", err)
}
```


See also [example.go](https://github.com/facebookincubator/zk/blob/master/example/main.go) for an example CLI which uses this library.

## TODO

* Support for [watches](https://zookeeper.apache.org/doc/current/zookeeperProgrammers.html#ch_zkWatches)
* Support for [locks](https://zookeeper.apache.org/doc/r3.1.2/recipes.html#sc_recipes_Locks) and other recipes
* Extension of [the API](https://zookeeper.apache.org/doc/r3.4.6/api/org/apache/zookeeper/ZooKeeper.html) so that all of Zookeeper’s client commands are supported
* Digest-based and SASL authentication

## Contributing

Want to contribute? Great! See the `CONTRIBUTING.md` file for more information on how to help out.
