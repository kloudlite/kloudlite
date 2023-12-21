{
  description = "kloudlite operator development setup";

  inputs = {
    nixpkgs = {
        url = "nixpkgs/nixos-unstable";
      };
  };

  outputs = { self, nixpkgs, flake-utils }: 
    flake-utils.lib.eachDefaultSystem(system: 
      let 
        pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          hardeningDisable = [ "all" ];
          buildInputs = with pkgs; [
            # cli tools
            curl
            jq
            yq
            go-task

            # source version control
            git
            pre-commit

            # programming tools
            go_1_21
            operator-sdk
            mongosh

            # kubernetes specific tools
            k9s
            kubectl
            kubernetes-helm

            # grpc tools
            protobuf
            protoc-gen-go
            protoc-gen-go-grpc

            # build tools
            gnumake
            podman
            upx
          ];

          shellHook = ''
            echo "You are using nix flakes"
          '';
          };
        }
      );
}
