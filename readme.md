# ReplaceURLRegex a Traefik plugin

[Traefik](https://traefik.io) plugins are developed using the [Go language](https://golang.org).

A [Traefik](https://traefik.io) middleware plugin is just a [Go package](https://golang.org/ref/spec#Packages) that provides an `http.Handler` to perform specific processing of requests and responses.

Rather than being pre-compiled and linked, however, plugins are executed on the fly by [Yaegi](https://github.com/traefik/yaegi), an embedded Go interpreter.

## Usage

For a plugin to be active for a given Traefik instance, it must be declared in the static configuration.

Plugins are parsed and loaded exclusively during startup, which allows Traefik to check the integrity of the code and catch errors early on.
If an error occurs during loading, the plugin is disabled.

For security reasons, it is not possible to start a new plugin or modify an existing one while Traefik is running.

Once loaded, middleware plugins behave exactly like statically compiled middlewares.
Their instantiation and behavior are driven by the dynamic configuration.

Plugin dependencies must be [vendored](https://golang.org/ref/mod#tmp_25) for each plugin.
Vendored packages should be included in the plugin's GitHub repository. ([Go modules](https://blog.golang.org/using-go-modules) are not supported.)

### Configuration

For each plugin, the Traefik static configuration must define the module name (as is usual for Go packages).

The following declaration (given here in YAML) defines an plugin:

```yaml
# Static configuration
pilot:
  token: xxxxx

experimental:
  plugins:
    replacepathfromurlregex:
      moduleName: github.com/schackoa/replacepathfromurlregex
      version: v0.0.4
```

Here is an example of a file provider dynamic configuration (given here in YAML), where the interesting part is the `http.middlewares` section:

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`, `{subdomain:[a-z]+}.nginx.roc`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - my-plugin

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:8080
  
  middlewares:
    my-plugin:
      plugin:
        replacepathfromurlregex:
          regex: "^https?://([a-z]+).nginx.roc(:[0-9]+)?/prefix/(.*?)$"
          replacement: /${1}/${2}
```

### Dev Mode

Traefik also offers a developer mode that can be used for temporary testing of plugins not hosted on GitHub.
To use a plugin in dev mode, the Traefik static configuration must define the module name (as is usual for Go packages) and a path to a [Go workspace](https://golang.org/doc/gopath_code.html#Workspaces), which can be the local GOPATH or any directory.

```yaml
# Static configuration
pilot:
  token: xxxxx

experimental:
  devPlugin:
    goPath: /plugins/go
    moduleName: github.com/schackoa/replacepathfromurlregex
```

(In the above example, the `replacepathfromurlregex` plugin will be loaded from the path `/plugins/go/src/github.com/schackoa/replacepathfromurlregex`.)

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - my-plugin

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    my-plugin:
      plugin:
        dev:
          regex: "^https?://([a-z]+).nginx.roc(:[0-9]+)?/prefix(.*?)$"
          replacement: /${1}${2}
```
