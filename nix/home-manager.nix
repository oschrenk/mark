# home-manager module for mark.
#
#   inputs.mark.url = "github:oschrenk/mark";
#   imports = [ inputs.mark.homeModules.mark ];
#
#   programs.mark = {
#     enable = true;
#     targets.homelab = {
#       default = true;
#       url = "http://pi-2:9090/api/v1/write";
#     };
#   };
#
# Generates $XDG_CONFIG_HOME/mark/config.toml, the path config.Path() reads.
self:
{
  config,
  lib,
  pkgs,
  ...
}:

let
  inherit (lib)
    literalExpression
    mkEnableOption
    mkIf
    mkOption
    types
    ;

  cfg = config.programs.mark;

  tomlFormat = pkgs.formats.toml { };

  targetType = types.submodule {
    options = {
      url = mkOption {
        type = types.str;
        example = "http://pi-2:9090/api/v1/write";
        description = ''
          Prometheus remote write endpoint. Expanded by mark through
          `os.ExpandEnv`, so `$HOME` and friends may be used literally.
        '';
      };

      metric = mkOption {
        type = types.str;
        # Mirrors config.DefaultMetric. Written out explicitly so the file
        # states the series name rather than relying on the Go fallback.
        default = "homelab_event";
        description = "Name of the Prometheus series events are written to.";
      };

      default = mkOption {
        type = types.bool;
        default = false;
        description = ''
          Use this target when `mark add` is run without `--target`. At most one
          target may set it.
        '';
      };
    };
  };

  settings = {
    targets = lib.mapAttrs (_: target: {
      inherit (target) url metric default;
    }) cfg.targets;
  };
in
{
  options.programs.mark = {
    enable = mkEnableOption "mark, which records a timestamped event in Prometheus";

    package = mkOption {
      type = types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.mark;
      defaultText = literalExpression "mark.packages.\${system}.mark";
      description = "The mark package to install.";
    };

    targets = mkOption {
      type = types.attrsOf targetType;
      default = { };
      example = literalExpression ''
        {
          homelab = {
            default = true;
            url = "http://pi-2:9090/api/v1/write";
          };
          lab.url = "http://pi-4:9090/api/v1/write";
        }
      '';
      description = ''
        Targets keyed by the name `mark add --target` selects.
      '';
    };
  };

  config = mkIf cfg.enable {
    assertions = [
      {
        assertion = cfg.targets != { };
        message = "programs.mark: at least one target must be defined.";
      }
      {
        # config.Load() rejects a file with more than one default target, so
        # catch it at build time rather than on the first `mark add`.
        assertion = lib.count (t: t.default) (lib.attrValues cfg.targets) <= 1;
        message = "programs.mark: at most one target may set `default = true`.";
      }
    ];

    home.packages = [ cfg.package ];

    # Written unconditionally: config.Load() fails without this file, so an
    # enabled mark with no config could not record anything.
    xdg.configFile."mark/config.toml".source = tomlFormat.generate "mark-config.toml" settings;
  };
}
