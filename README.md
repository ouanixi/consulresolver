# Lightweight GRPC consul resolver

Simple Consul resolver that works alongside gRPC client side load balancing. It listens in to changes in Consul and updates the State accordingly and asynchronously.

## Usage

```go
import (
    "github.com/ouanixi/consulresolver"

    "github.com/hashicorp/consul/api"
    "google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"

)
func main() {
	builder := consulresolver.NewConsulBuilder(api.DefaultConfig())
	resolver.Register(builder)

	conn, _ = grpc.Dial(
			"consul:///email?tags=grpc",
			grpc.WithBalancerName(roundrobin.Name),
        )
    defer conn.Close()

}
```