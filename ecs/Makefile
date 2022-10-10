.PHONY=bench

all:
	cargo build --release --manifest-path ./rust/Cargo.toml --target-dir rust-bin
	cargo build --manifest-path ./rust/Cargo.toml --target-dir rust-bin
	go build -o go-bin/ github.com/unitoftime/ecs/bench

run:
	mkdir -p results
	bash run.sh
