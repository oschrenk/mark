{
  description = "Mark - record a timestamped event in Prometheus";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  # Offer prebuilt binaries from the Cachix cache so `nix profile install`
  # downloads instead of compiling. Consumers are prompted to trust these.
  nixConfig = {
    extra-substituters = [ "https://oschrenk.cachix.org" ];
    extra-trusted-public-keys = [
      "oschrenk.cachix.org-1:3JOMfkq2vFiLw4UsCVwzu8kWFBkuS/3DD5AojcO9pks="
    ];
  };

  outputs =
    { self, nixpkgs }:
    let
      # The commit is the release. There are no tags and no VERSION file: a
      # consumer pins this repo by rev in their own flake.lock, so the rev is
      # the only identifier that means anything. Unset on a dirty tree.
      version = self.shortRev or "dirty";

      # Pure HTTP client with no platform-specific code, so it builds anywhere.
      # aarch64-linux is here because the pis are the natural place to record an
      # event from a script.
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        mark = pkgs.buildGoModule {
          pname = "mark";
          inherit version;
          src = self;

          # Regenerate after changing go.mod/go.sum: set to lib.fakeHash,
          # run `nix build`, then paste the expected hash from the error.
          vendorHash = "sha256-z37KqCKGM6kXU9LzPwRGBbBezSciDSztRePpxXm1un4=";

          # Version and Commit are the same string by construction. BuildDate is
          # a constant rather than a timestamp: nix builds are reproducible, and
          # a real date would make every rebuild a different derivation, which
          # would miss the binary cache.
          ldflags = [
            "-s"
            "-w"
            "-X github.com/oschrenk/mark/internal/cli.Version=${version}"
            "-X github.com/oschrenk/mark/internal/cli.Commit=${version}"
            "-X github.com/oschrenk/mark/internal/cli.BuildDate=nix"
          ];

          # `completion` is hidden (internal/cli/root.go) but not disabled, and
          # it never reaches the config loader or the network, so it is safe to
          # run in the sandbox.
          nativeBuildInputs = [ pkgs.installShellFiles ];
          postInstall = ''
            installShellCompletion --cmd mark \
              --bash <($out/bin/mark completion bash) \
              --zsh <($out/bin/mark completion zsh) \
              --fish <($out/bin/mark completion fish)
          '';

          meta = {
            description = "Record a timestamped event in Prometheus";
            homepage = "https://github.com/oschrenk/mark";
            mainProgram = "mark";
          };
        };
        default = mark;
      });

      apps = forAllSystems (pkgs: rec {
        mark = {
          type = "app";
          program = "${self.packages.${pkgs.stdenv.hostPlatform.system}.mark}/bin/mark";
        };
        default = mark;
      });

      homeModules = rec {
        mark = import ./nix/home-manager.nix self;
        default = mark;
      };

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go # go, language
            golangci-lint # go, linter runner
            gopls # go, lsp
            go-tools # go, staticcheck for `task check`
          ];
        };
      });
    };
}
