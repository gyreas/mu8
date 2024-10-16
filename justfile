build:
    go build .

examples: build
    ./mu8 examples/randomnumber/random_number_test.ch8
    ./mu8 examples/hex-to-dec.ch8
