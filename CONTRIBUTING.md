# Contributing

Issues and pull requests are welcome.

## Setup

    go build
    go vet ./... && go test ./...

The tests don't need a Razer device. To try changes on real hardware you need Linux with OpenRazer (see the README); note your settings first, since changes apply right away.

## Adding a setting

Mice and keyboards go through the OpenRazer daemon over D-Bus (`internal/razer/daemon.go`), headsets through the driver's sysfs files (`internal/razer/headset.go`). A setting only shows up when its `Get` works, so a device without it simply doesn't list it. Set `Linked` when a write can change other rows.

## Style

- `gofmt`, small functions, return errors instead of panicking.
- Comments explain why, not what.
- No new dependencies for what a few lines can do.
- This project is MIT: don't copy code from GPL projects like OpenRazer or Polychromatic. D-Bus method names and sysfs file names are fine.

## Commits and branches

- [Conventional Commits](https://www.conventionalcommits.org): `feat(#12): add mouse lift-off distance`, or `fix: …` without an issue. Lowercase, imperative, no period.
- Branches: `<type>/<issue>-<short-desc>`, e.g. `feat/12-lift-off-distance`.
