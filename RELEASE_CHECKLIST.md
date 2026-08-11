# 3m-ui Release Checklist

Recommended first release-candidate tag: `v0.1.0-rc.1`.

## Verification

- [ ] Run `go fmt ./...` from `backend/` and confirm no uncommitted formatting changes remain.
- [ ] Run `go test ./...` from `backend/`.
- [ ] Run `go vet ./...` from `backend/`.
- [ ] Run `pnpm build` from `frontend/`.
- [ ] Build a release binary and verify `3m-ui --version` prints the tag, git commit, and UTC build time.

## Deployment Audit

- [ ] Fresh install on systemd: installer creates `/etc/3m-ui/config.yaml`, `/var/lib/3m-ui/`, installs `/usr/local/bin/3m-ui`, enables and starts `3m-ui.service`.
- [ ] Fresh install on OpenRC: installer creates the same directories, installs `/etc/init.d/3m-ui`, adds it to the default runlevel, and starts it.
- [ ] Upgrade on systemd and OpenRC: updater backs up `/etc/3m-ui`, replaces only `/usr/local/bin/3m-ui`, restarts the service, and keeps `/var/lib/3m-ui/3m-ui.db` intact.
- [ ] Uninstall default mode: removes service definitions, binary, config directory, and log directory while keeping `/var/lib/3m-ui`.
- [ ] Uninstall purge mode: `uninstall.sh --purge` removes `/var/lib/3m-ui` only after confirmation.

## GitHub Actions Audit

- [ ] Push a `v*` tag to trigger `.github/workflows/release.yml`.
- [ ] Confirm frontend assets are copied from `frontend/dist` to `backend/cmd/server/web/dist` before Go builds.
- [ ] Confirm release artifacts include `3m-ui-linux-amd64`, `3m-ui-linux-arm64`, and `3m-ui-linux-armv7`.
- [ ] Confirm release artifacts include `install.sh`, `update.sh`, and `uninstall.sh`.
- [ ] Download each binary artifact and run `--version` where architecture access is available.

## Release Notes

- [ ] Document supported Linux init systems: systemd and OpenRC.
- [ ] Document supported architectures: amd64, arm64, armv7.
- [ ] Document installation, upgrade, uninstall, and purge commands.
- [ ] List remaining known issues and operational caveats.
