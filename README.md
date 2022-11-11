# ratify-verifier-plugin

This is a sample verifier plugin for [Ratify](https://github.com/deislabs/ratify), written in Go

For more details on how plugins work, please visit the [verifier specification](https://github.com/deislabs/ratify/blob/main/docs/verifier.md)

## Usage

### Build

```shell
# Build the plugin
go build -o sample .
```

### Standalone testing

Ratify plugins use a combination of environment variables and STDIN to run plugins. This example sets the environment variables in the shell and uses the data from `hack/stdin.json` to provide configuration

```shell
# Run the plugin standalone
export RATIFY_VERIFIER_VERSION=1.0.0
export RATIFY_VERIFIER_COMMAND=VERIFY
export RATIFY_VERIFIER_SUBJECT=wabbitnetworks.azurecr.io/test/net-monitor:signed
cat hack/stdin.json | ./sample
```

### Debugging in VS Code

You can debug your verifier using VS Code

- Press `F5` to start the `Debug` launch configuration
- You'll be prompted for the subject (defaults to the sample image)
- At this point, the debugger is active but waiting for input. You'll have the plugin running in a terminal pane
- Copy the contents of `hack/stdin.json` and paste it into the terminal, then **press Ctrl+D to send EOF** to the input stream, which will trigger the plugin to execute

### Local usage with Ratify

After it has been built, the binary is ready be used with Ratify

First, copy it to the plugins dir to make it available for use

```shell
# Copy to the default Ratify plugins dir
mkdir ~/.ratify/plugins
cp ./sample ~/.ratify/plugins/sample
```

Next, add an entry to `verifier.plugins` in the Ratify config to activate your verifier plugin

```json
{
  "executor": {},
  "store": {
    "version": "1.0.0",
    "plugins": [
      {
        "name": "oras"
      }
    ]
  },
  "policy": {
    "version": "1.0.0",
    "plugin": {
      "name": "configPolicy",
      "artifactVerificationPolicies": {
        "application/vnd.cncf.notary.v2.signature": "all"
      }
    }
  },
  "verifier": {
    "version": "1.0.0",
    "plugins": [
      {
        "name": "sample",
        "artifactTypes": "application/vnd.cncf.notary.v2.signature"
      },
      {
        "name": "notaryv2",
        "artifactTypes": "application/vnd.cncf.notary.v2.signature"
      }
    ]
  }
}
```


### Deploy with Ratify to Kubernetes

Ratify ships a [distroless](https://github.com/GoogleContainerTools/distroless) image, so your plugin must be built with `CGO_ENABLED=0`, ex:

```shell
CGO_ENABLED=0 go build -o sample .
```

Regardless of how you build and distribute your plugin, users need to have it accessible within their Ratify container. Ex:

```Dockerfile
# See note on CRDs below; this version won't work as-is yet
FROM ghcr.io/deislabs/ratify:v1.0.0-alpha.3 AS ratify

COPY ./sample /.ratify/plugins/sample
```

You'll need to use this image, which contains your plugin, in your Ratify chart deployment. Ex:

```shell
# See note on CRDs below; this version of Ratify won't work as-is yet
docker build -t myregistry.azurecr.io/ratify-with-plugins:v1.0.0-alpha.3 .
docker push myregistry.azurecr.io/ratify-with-plugins:v1.0.0-alpha.3
```

And in your Ratify [chart](https://github.com/deislabs/ratify/tree/main/charts/ratify) values:

```yaml
image:
  repository: myregistry.azurecr.io/ratify-with-plugins
  tag: v1.0.0-alpha.3
  pullPolicy: IfNotPresent
# /snip...
```

#### Temporary workaround: v1.0.0-alpha.3

This gets you a Ratify deployment with your plugin available. The final step is to activate it by adding updating your `ratify-configuration` ConfigMap

#### Future

> Note: Ratify CRD support [just landed](https://github.com/deislabs/ratify/pull/349/files), but it hasn't been published yet, so this doesn't actually work unless you build all of Ratify yourself

Create a `Verifier` resource to register your custom plugin

```yaml
apiVersion: config.ratify.deislabs.io/v1alpha1
kind: Verifier
metadata:
  name: verifier-sample
spec:
  name: sample
  artifactTypes: application/vnd.cncf.notary.v2.signature
  parameters: {}
```
