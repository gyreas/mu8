build: mu8

mu8:
    go build -o mu8 Mu8.go cpu.go memory.go screen.go grafix.go

test: build
    ./mu8 examples/chip8-test-rom/test_opcode.ch8 2> mu8.log

examples: build
    ./mu8 examples/randomnumber/random_number_test.ch8
    ./mu8 examples/hex-to-dec.ch8
    ./mu8 examples/ibm-logo.ch8
