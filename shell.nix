let current= import <nixpkgs> {};
in
current.stdenv.mkDerivation {
  name = "sctf-dev";

  hardeningDisable = [ "stackprotector" "fortify" ];

  buildInputs = [
    current.docker current.docker_compose current.mysql current.jq current.go_1_14 current.ag
  ];
  shellHook = ''
  export GOPATH="$HOME/work/scylla/sctf-go"
  export GOBIN="$GOPATH/bin"
  mkdir -p "$GOBIN"
  export PATH="$GOPATH/bin":$PATH
  export GO111MODULE=on
  '';
}
