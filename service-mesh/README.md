## istio tracing 概述

istio官方的介绍为:
> Istio makes it easy to create a network of deployed services with load balancing, service-to-service authentication, monitoring, and more, without any changes in service code.

istio在使用时，不对代码做任何处理即可进行服务治理，但是实际使用过程中，不修改服务代码，istio的调用链总是断开的。

在 Istio 中，所有的治理逻辑的执行体都是和业务容器一起部署的 Envoy 这个 Sidecar，不管是负载均衡、熔断、流量路由还是安全、可观察性的数据生成都是在 Envoy 上。Sidecar 拦截了所有的流入和流出业务程序的流量，根据收到的规则执行执行各种动作。实际使用中一般是基于 K8S 提供的 InitContainer 机制，用于在 Pod 中执行一些初始化任务. InitContainer 中执行了一段 Iptables 的脚本。正是通过这些 Iptables 规则拦截 pod 中流量，并发送到 Envoy 上。Envoy 拦截到 Inbound 和 Outbound 的流量会分别作不同操作，执行上面配置的操作，另外再把请求往下发，对于 Outbound 就是根据服务发现找到对应的目标服务后端上；对于 Inbound 流量则直接发到本地的服务实例上。

Envoy的埋点规则为:
- Inbound 流量：对于经过 Sidecar 流入应用程序的流量，如果经过 Sidecar 时 Header 中没有任何跟踪相关的信息，则会在创建一个根 Span，TraceId 就是这个 SpanId，然后再将请求传递给业务容器的服务；如果请求中包含 Trace 相关的信息，则 Sidecar 从中提取 Trace 的上下文信息并发给应用程序。
- Outbound 流量：对于经过 Sidecar 流出的流量，如果经过 Sidecar 时 Header 中没有任何跟踪相关的信息，则会创建根 Span，并将该跟 Span 相关上下文信息放在请求头中传递给下一个调用的服务；当存在 Trace 信息时，Sidecar 从 Header 中提取 Span 相关信息，并基于这个 Span 创建子 Span，并将新的 Span 信息加在请求头中传递。

根据这个规则，对于一个api->A-这个简单调用，我们有如下分析:
- 当一个请求进入api时，该请求头中没有任何trace相关的信息,对于这个inbound流量，istio会创建一个根span，并向请求头注入span信息。
- 当api向A创建并发送rpc或http请求时，这个请求对于api的envoy来说时outbound流量，如果请求头中没有trace信息，会创建根span信息填入请求头
- 这种情况下，在istio的jaeger页面上我们可以看到两段断裂的trace记录

**结论：埋点逻辑是在 Sidecar 代理中完成，应用程序不用处理复杂的埋点逻辑，但应用程序需要配合在请求头上传递生成的 Trace 相关信息。**

istio使用jaeger作为trace系统，格式为zipkin format。在请求头中有如下headers:

- x-request-id
- x-b3-traceid
- x-b3-spanid
- x-b3-parentspanid
- x-b3-sampled
- x-b3-flags
- x-ot-span-context

注意: 在http请求中，比如使用gin框架时，这些header中的key应是首字母大写的，例如:X-Request-Id

## 部署运行

```shell
kubectl label namespace default istio-injection=enabled --overwrite
kubectl apply -f ./k8s/k8s.yaml
```


## 参考

[微服务使用istio分布式追踪](https://hanamichi.wiki/posts/go-micro-istio/)