# mHub - simple message hub

---

### Use Case

![usecase](usecase.png)

### Usage

#### Requirements

A running redis instance [check here how to start](https://redis.io/topics/quickstart)

#### Install

`go get -u github.com/jquiterio/mhub`

#### Generate Certificates

```bash
mkdir ~/go-spb_certificates
cd ~/go-spb_certificates

cp $GOPATH/src/github.com/jquiterio/go-spb/gencert.sh .
./gencert.sh localhost
```

#### Start the message Hub Server

##### Option 1: Build

```bash
go build github.com/jquiterio/mhub -o mhub
./mhub #for default values
# OR
HUB_ADDR=localhost HUB_PORT=8083 REDIS_ADDR=localhost:6379 REDIS_PASSWD="" ./mhub
```

##### Option 2: Pre-build Server Binary

Bla bla bla

##### Option 2: Docker

bla bla bla

##### Use client

```go
func main() {
  	cli, err := client.NewHubClient("localhost:8083")
	if err != nil {
		panic(err)
	}
	cli.MessageHandler = func(msg interface{}) error {
		fmt.Println(msg)
		return nil
	}
	fmt.Println(cli.ClientID)
	if ok := cli.AddTopic([]string{"test"}); !ok {
		panic("failed to add topic")
	}
	fmt.Println("local Topics: ", cli.Topics)
	if ok := cli.Subscribe(); !ok {
		panic("failed to subscribe")
	}
	cli.Publish("test", "Yay!")
	cli.Me()
	cli.GetMessages()
}
```
