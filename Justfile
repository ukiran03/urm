# Variable definition
binary_name := "app"

# Show available commands
default:
    @just --list

# Build the binary
build:
    mkdir -p build
    go build -o build/{{binary_name}} .

# Build and run the app
# Usage: just run "--my-flag=value"
run args="": build
    ./build/{{binary_name}} {{args}}

# Remove the binary and build folder
clean:
    rm -rf build/
    go clean
