.PHONY: build release release-fast test run clean install uninstall

build:
	zig build

release:
	zig build -Doptimize=ReleaseSmall

release-fast:
	zig build -Doptimize=ReleaseFast

test:
	zig build test

run:
	zig build run -- $(ARGS)

clean:
	rm -rf zig-out .zig-cache

install: release
	install -Dm755 zig-out/bin/nanogen $(HOME)/.local/bin/nanogen

uninstall:
	rm -f $(HOME)/.local/bin/nanogen
