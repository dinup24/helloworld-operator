### Helloworld operator

#### Pre-requisites (MacOS)
```
1. Install Operator SDK

brew install operator-sdk
```

```
operator-sdk init --domain=helloworld.ibm.com --repo=github.com/dinup24/helloworld-operator

operator-sdk create api --group=apps --version=v1alpha1 --kind=HelloWorld

make generate

make manifests

make install

export USERNAME=dinup24

make docker-build IMG=docker.io/$USERNAME/helloworld-operator:v0.0.1

docker login docker.io
dinup24/******

make docker-push IMG=docker.io/$USERNAME/helloworld-operator:v0.0.1

cd config/default/ && /root/go/bin/kustomize edit set namespace "default" && cd ../..

make deploy IMG=docker.io/$USERNAME/helloworld-operator:v0.0.1
```
