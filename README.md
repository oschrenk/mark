# `mark`

Record a timestamped event in Prometheus.

`mark` writes the event as a Prometheus series, which Perses 0.55+ then can overlay as a panel through its `PrometheusPromQLAnnotation` plugin.

## Usage

```sh
mark add "swapped the 120mm fan" --tag hardware --tag fan
mark add "ISP outage, homelab down" --at 2026-09-07T16:41:00-06:00 --tag outage
mark add "test" --dry-run
```

`--dry-run` prints the request and sends nothing:

```text
would POST to http://pi-2:9090/api/v1/write

  __name__     homelab_event
  description  swapped the 120mm fan
  tags         hardware,fan

  value        1
  timestamp    1788884381151  (2026-09-08T10:19:41-06:00)

  payload      107 bytes protobuf, before snappy

  as promql: homelab_event{description="swapped the 120mm fan", tags="hardware,fan"}
```

## Install

There are no tags and no release artifacts.
The commit is the release, so pin the rev in your own `flake.lock` and `nix flake update` when you want a newer one.

```nix
inputs.mark.url = "github:oschrenk/mark";
```

Prebuilt `aarch64-darwin` binaries come from the `oschrenk` Cachix cache, which the flake offers as a substituter.
The linux systems build from source.

### home-manager

The flake ships a home-manager module that installs mark and generates
`~/.config/mark/config.toml`, so targets are declared in Nix rather than
hand-edited.

```nix
{
  inputs.mark.url = "github:oschrenk/mark";

  # in your home-manager configuration:
  imports = [ inputs.mark.homeModules.mark ];

  programs.mark = {
    enable = true;

    targets.homelab = {
      default = true;
      url = "http://pi-2:9090/api/v1/write";
    };
  };
}
```

`metric` defaults to `homelab_event`.
At most one target may set `default = true`, which the module checks at build time.

## Configuration

`$XDG_CONFIG_HOME/mark/config.toml`, or `~/.config/mark/config.toml`.
See `config.example.toml`.

```toml
[targets.homelab]
default = true
url = "http://pi-2:9090/api/v1/write"
metric = "homelab_event"
```

Define as many targets as you need.
Pick one with `--target`, or mark one `default`.
