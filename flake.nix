{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, treefmt-nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        treefmtEval = treefmt-nix.lib.evalModule pkgs {
          projectRootFile = "flake.nix";
          programs.nixpkgs-fmt.enable = true;
          programs.prettier = {
            enable = true;
            includes = [ "*.md" "*.yaml" "*.yml" ];
          };
          programs.gofmt.enable = true;
        };

        schlag-o-meter = pkgs.buildGoModule {
          name = "schlag-o-meter";
          version = "0.0.1";
          vendorHash = "sha256-gA1t0HFOeXH1D7VRw5XCXE4oQ/DKdgVbQVJU3sm1wIs=";
          src = ./.;
        };

        containerImage = pkgs.dockerTools.buildLayeredImage {
          name = "ghcr.io/rubenhoenle/schlag-o-meterer";
          tag = "unstable";
          config = {
            Entrypoint = [ "${schlag-o-meter}/bin/schlag-o-meter" ];
          };
        };
      in
      {
        formatter = treefmtEval.config.build.wrapper;
        checks.formatter = treefmtEval.config.build.check self;

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
          ];
        };

        packages = flake-utils.lib.flattenTree {
          default = schlag-o-meter;
          containerimage = containerImage;
        };
      }
    );
}
