build: mu8

mu8:
    go build -race -o mu8 Mu8.go cpu.go memory.go screen.go grafix.go
    ./mu8

screen:
    go build -o screen sc.go
    ./screen

spin:
    go build -o spinner spinner.go
    ./spinner

tcell:
    go build -o gfx screen.go grafix.go
    ./gfx

resize:
    go build -race -o res resize.go
    ./res

examples: build
    ./mu8 examples/randomnumber/random_number_test.ch8
    ./mu8 examples/hex-to-dec.ch8
    ./mu8 examples/ibm-logo.ch8
