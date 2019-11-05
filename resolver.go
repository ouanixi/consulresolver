package resolver

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

const Scheme = "consul"

type Target struct {
	serviceName string
	tags        []string
}

type ConsulResolverBuilder struct {
	ConsulClientConfig *api.Config
}

func NewConsulBuilder(cc *api.Config) resolver.Builder {
	return &ConsulResolverBuilder{ConsulClientConfig: cc}
}

func (b *ConsulResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	tgt, err := buildTarget(target)
	if err != nil {
		return nil, err
	}
	consul, err := api.NewClient(b.ConsulClientConfig)

	if err != nil {
		return nil, err
	}

	r := NewConsulResolver(cc, consul, tgt)
	return r, nil
}

func buildTarget(target resolver.Target) (*Target, error) {
	rawTarget := fmt.Sprintf("%s:/%s/%s", target.Scheme, target.Authority, target.Endpoint)
	parsed, err := url.Parse(rawTarget)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != Scheme {
		return nil, fmt.Errorf(fmt.Sprintf("wrong scheme passed. Must be `%s`", Scheme))
	}
	query := parsed.Query()

	tags := []string{}
	rawTags := query.Get("tags")

	if rawTags != "" {
		ts := strings.Split(rawTags, ",")
		for _, tag := range ts {
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}
	return &Target{
		serviceName: parsed.Host,
		tags:        tags,
	}, nil
}

func (b *ConsulResolverBuilder) Scheme() string {
	return Scheme
}

type ConsulResolver struct {
	target    *Target
	lastIndex uint64
	cc        resolver.ClientConn
	consul    *api.Client
	done      bool
}

func NewConsulResolver(cc resolver.ClientConn, consul *api.Client, tgt *Target) *ConsulResolver {
	r := ConsulResolver{
		target: tgt,
		cc:     cc,
		consul: consul,
		done:   false,
	}
	go r.updateConnState()
	return &r
}

func (r *ConsulResolver) ResolveNow(resolver.ResolveNowOption) {}

func (r *ConsulResolver) updateConnState() {
	for {
		if r.done {
			return
		}
		// following call to consul will block until there's a new index value
		services, meta, err := r.consul.Health().ServiceMultipleTags(
			r.target.serviceName,
			r.target.tags,
			true,
			&api.QueryOptions{WaitIndex: r.lastIndex},
		)
		if err != nil {
			grpclog.Errorf("error retrieving instances from Consul: %v", err)
			return
		}
		r.lastIndex = meta.LastIndex
		addrs := []resolver.Address{}
		for _, s := range services {
			addr := resolver.Address{
				ServerName: r.target.serviceName,
				Addr:       s.Service.Address + ":" + strconv.Itoa(s.Service.Port),
			}
			addrs = append(addrs, addr)
		}
		grpclog.Info("updating balancer state")
		r.cc.UpdateState(resolver.State{Addresses: addrs})
	}

}

func (r *ConsulResolver) Close() {
	r.done = true
}
