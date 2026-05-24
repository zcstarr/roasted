package main

import (
	"flag"
	"log"

	"github.com/jacobsa/go-serial/serial"
	"github.com/zcstarr/roasted/pkg"
	"github.com/zcstarr/roasted/sr700"
)

func main() {
	var (
		device = flag.String("device", "/dev/tty.wchusbserial1410", "SR700 serial device")
		debug  = flag.Bool("debug", false, "enable debug logging")
	)
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("usage: roasted [-device PATH] [-debug] <recipe.json>")
	}

	recipe, err := pkg.LoadRecipe(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	// todo: validate program before running

	options := serial.OpenOptions{
		PortName:              *device,
		BaudRate:              9600,
		DataBits:              8,
		StopBits:              1,
		InterCharacterTimeout: 100,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	defer func() { _ = port.Close() }()

	roaster := sr700.New(port)
	roaster.SetDebug(*debug)

	pgm, err := roaster.Connect()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected.  Program: ", pgm)

	switch r := recipe.(type) {
	case *pkg.SimpleRecipe:
		err = pkg.RunSimpleRecipe(roaster, r)
	case *pkg.OpenRoastRecipe:
		err = pkg.RunOpenRoastRecipe(roaster, r)
	default:
		log.Fatalf("unsupported recipe format: %T", recipe)
	}
	if err != nil {
		log.Printf("error stopping: %v", err)
	}
}
