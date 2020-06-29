# feph

File Exist Probes Helper for Azure Application Gateway Ingress Controller on Kubernetes.

## background

Azure [AGIC](https://github.com/Azure/application-gateway-kubernetes-ingress)(Application Gateway Ingress Controller) on AKS(Azure Kubernetes Service) is only support httpGet probes now. This small binary is server to provide endpoint check wheather file is exist or not.

## How to use

This project can use set Dockerfile like below. 

p.s Thank you for wonderful project [glare](https://github.com/Contextualist/glare).

```
## Dockerfile
FROM debian

...  ## any of your code

ARG FEPH_VER=v0.0.14
ENV FEPH_PORT=4000
ENV TARGET_PORT=3000
ENV CHECK_DIR=./
RUN apt-get update && apt-get install -y curl \
    && curl -L https://glare.now.sh/mrchypark/feph@{$FEPH_VER}/feph-{$FEPH_VER}-linux-amd64.tar.gz -o feph.tar.gz \
    && tar -zxvf feph.tar.gz \
    && rm feph.tar.gz \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get purge -y --auto-remove

...  ## any of your code

ENTRYPOINT [""]
CMD ["sh","-c","<User CMD> | ./feph"]

```

### Env

feph has default env like below.

```
ENV FEPH_PORT=4000
ENV TARGET_PORT=5005
ENV CHECK_DIR=./

```

### Support Endpoint

- ext
```
localhost:4000/ext/:ext
```

ext is extention name of file. If you want to check file what is named end of `.txt` is exist, you can set health check on deployment.yaml like below.

```
## deployment.yaml
...
  livenessProbe:
    httpGet:
      path: /ext/text
      port: 4000
    periodSeconds: 10
    timeoutSeconds: 10
...
```

- filename
```
localhost:4000/filename/:name
```

filename means full file name. you can use this endpoint like below.

```
## deployment.yaml
...
  livenessProbe:
    httpGet:
      path: /filename/checkfile.log
      port: 4000
    periodSeconds: 10
    timeoutSeconds: 10
...
```


- contain
```
localhost:4000/contain/:string
```

contain means contain query on file name. if you want to check file is exist that contain `logging` text on file name, you can use this endpoint like below.

```
## deployment.yaml
...
  livenessProbe:
    httpGet:
      path: /contain/logging
      port: 4000
    periodSeconds: 10
    timeoutSeconds: 10
...
```
