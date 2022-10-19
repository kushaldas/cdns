_default:
  @just --list --unsorted

clean:
  rm ./catchthedns

build:
  go build github.com/kushaldas/cdns/cmd/cdns
  sudo setcap cap_net_bind_service=+ep ./cdns
