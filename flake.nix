{
  description = "Bridge for various implementations to publish to ntfy.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    treefmt-nix.url = "github:numtide/treefmt-nix";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        inputs.treefmt-nix.flakeModule
      ];
      systems = ["x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin"];
      perSystem = {
        config,
        self',
        inputs',
        pkgs,
        system,
        ...
      }: {
        treefmt = {
          projectRootFile = ".git/config";
          programs = {
            alejandra.enable = true;
            gofumpt.enable = true;
          };
        };

        devShells.default =
          pkgs.mkShell {buildInputs = with pkgs; [go gopls];};
      };
      flake = {
      };
    };
}
