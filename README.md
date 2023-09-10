# mermaid-lambda

Mermaid.js as a service — see https://thekevinwang.com/wiki/meta/mermaid for examples.

## Intended usage

This repository is meant to build an OCI image, using Docker.
The image is intended to be run as a stateless container, inside
AWS Lambda.

### Minimal

At a bare minimum, you can run use the image inside AWS lambda,
and create a Lambda Function URL.

- [x] Serverless
- [ ] ~~Custom Domain Name~~
- [ ] ~~Edge Caching~~

```mermaid
sequenceDiagram
    participant CL as Client
    participant FU as Function URL
    participant LM as Lambda
```

### Full

You can also leave add an API Gateway proxy integration, along
with CloudFront CDN. This opens up the door to edge caching,
and custom domain names via the Route53 → CloudFront alias
integration.

- [x] Serverless
- [x] Custom Domain Name (via Route53, ACM, CloudFront)
- [x] Edge Caching

```mermaid
sequenceDiagram
    participant CL as Client
    participant CF as CloudFront
    participant AGW as API Gateway
    participant LM as Lambda

    CL->>CF: GET
    activate CF
    alt Cache Hit
        CF->>CL: Cached Response
    else Cache Miss
        CF->>+AGW: Origin Request
        AGW->>+LM: Proxy Integration
        Note over LM: Run Container via ECR image
        LM->>-AGW: (Content-Type: text/plain)
        AGW->>-CF: Origin Response
        CF->>CL: (Content-Type: image/svg+xml)
    end
    deactivate CF
```

## Todo

**Endpoint Documentation**

**Error handling**

**Security**

> **Warning**
> This is not yet considered a _secure_ implementation.
