{
  description = "Nix flakes";

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: 
    flake-utils.lib.eachDefaultSystem (system: 
      let
        pythonEnv = pkgs.python312.withPackages (ps: []);
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
	    sqlite.dev
	    poetry
	    inkscape
          ];

	  LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath [
	    pkgs.sqlite.out
	  ];
        };
      });
}
