package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

type mu8Config struct {
	rompath     string
	profileMode bool
}

type Mu8 struct {
	cpu       Cpu
	video     Display
	romsz     int
	profiling bool
}

func InitMu8(config mu8Config) Mu8 {
	mu8 := Mu8{}

	if config.profileMode {
		mu8.profiling = true
		mu8.StartProfiling()
	}

	mu8.cpu = NewCpu()
	mu8.video = NewDisplay()

	mu8.load_rom(config.rompath)

	log.Println("Mu8! go!")

	return mu8
}

func (mu8 *Mu8) StartProfiling() {
	mu8prof, err := os.Create("mu8.prof")
	if err != nil {
		log.Fatal(err)
	}

	pprof.StartCPUProfile(mu8prof)
	log.Println("======== started profiling of the app ========")

}

func (Mu8 *Mu8) load_rom(rompath string) {
	rom, err := os.ReadFile(rompath)
	if err != nil {
		log.Fatalf("load_rom: %s\n", err.Error())
	}

	copy(Mu8.cpu.M[PROGRAM_ADDRESS_OFFSET:], rom)
	Mu8.romsz = len(rom)
}

func handle_flags() mu8Config {
	profiling := flag.Bool("profile", false, "enable CPU profiling")

	flag.Parse()

	posArgv := flag.Args()
	if len(posArgv) == 0 {
		fmt.Fprintf(os.Stderr, ""+
			"no path provided\n"+
			"usage: ./mu8 [-h] [-profile] <rom-path>\n")
		os.Exit(1)
	}

	config := mu8Config{}
	if *profiling {
		config.profileMode = true
	}
	config.rompath = posArgv[0]

	return config
}

func (Mu8 *Mu8) Quit() {
	Mu8.video.handleQuit()
	if Mu8.profiling {
		pprof.StopCPUProfile()
		// let's assume it closes the file
		log.Println("========= stop profiling of the app =========")
	}
}

const (
	debug  = true
	CPU_HZ = 44 * time.Microsecond
)

func main() {
	config := handle_flags()

	Mu8 := InitMu8(config)
	video := Mu8.video
	cpu := Mu8.cpu

	go video.startRenderLoop()
	defer Mu8.Quit()

cycle:
	for cpu.ip < MEMORY_SIZE {
		// collect key input for the next instruction that needs it
		cpu.cycle(&Mu8, video.collide, video.key, video.echan)

		select {
		case ev := <-video.echan:
			if ev.Kind == EventQuit {
				logmsg("[cycle]: closing")
				break cycle
			}
		default:
			<-time.After(CPU_HZ)
		}
	}
}

func logmsg(format string, args ...any) {
	if !debug {
		return
	}

	fmt.Fprintf(os.Stderr, format, args...)
}
