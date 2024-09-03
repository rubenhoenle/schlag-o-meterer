{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, treefmt-nix }:
    let
      system = "x86_64-linux";
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
    in
    {
      formatter.${system} = treefmtEval.config.build.wrapper;
      checks.${system}.formatter = treefmtEval.config.build.check self;

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          go
        ];
      };

      packages.${system} = {
        default = schlag-o-meter;
      };
    };
}
