{ pkgs, ... }:

{
  packages = [ pkgs.git pkgs.git-crypt pkgs.mosquitto pkgs.sqlc pkgs.ansible pkgs.ansible-lint];

  enterShell = ''
    echo "Welcome to your homelab"
  '';

  languages.go.enable = true;
}
