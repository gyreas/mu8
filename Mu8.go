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
	Rompath     string
	ProfileMode bool
}

type Mu8 struct {
	Cpu       Cpu
	Video     Display
	Romsz     int
	Profiling bool
}

func InitMu8(config mu8Config) Mu8 {
	mu8 := Mu8{}

	if config.ProfileMode {
		mu8.Profiling = true
		mu8.StartProfiling()
	}

	mu8.Cpu = NewCpu()
	mu8.Video = NewDisplay()

	mu8.LoadRom(config.Rompath)

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

func (Mu8 *Mu8) LoadRom(rompath string) {
	rom, err := os.ReadFile(rompath)
	if err != nil {
		log.Fatalf("load_rom: %s\n", err.Error())
	}

	copy(Mu8.Cpu.M[PROGRAM_ADDRESS_OFFSET:], rom)
	Mu8.Romsz = len(rom)
}

func handleFlags() mu8Config {
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
		config.ProfileMode = true
	}
	config.Rompath = posArgv[0]

	return config
}

func (Mu8 *Mu8) Quit() {
	Mu8.Video.HandleQuit()
	if Mu8.Profiling {
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
	config := handleFlags()

	Mu8 := InitMu8(config)
	video := Mu8.Video
	cpu := Mu8.Cpu

	go video.StartRenderLoop()
	defer Mu8.Quit()

cycle:
	for cpu.Ip < MEMORY_SIZE {
		// collect key input for the next instruction that needs it
		cpu.Cycle(&Mu8, video.Collide, video.Key, video.Echan)

		select {
		case ev := <-video.Echan:
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
