# Perses

## Drawing the Events

Perses annotations arrived after v0.54.0, so this needs the next release.

```cue
annotations: [{
    display: {name: "Events", color: "#fe8019"}
    plugin: {
        kind: "PrometheusPromQLAnnotation"
        spec: {
            expr:  "homelab_event"
            title: "{{description}}"
            tags:  ["tags"]
        }
    }
}]
```
