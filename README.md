# Go-BirdNET

<p align="center">
  <img src="doc/go-birdnet-logo.webp" />
</p>
<p align="center">
<a href="https://creativecommons.org/licenses/by-nc-sa/4.0/">
 <img src="https://badgen.net/badge/License/CC-BY-NC-SA%204.0/green">
 </a>
 <img src="https://badgen.net/badge/OS/Linux%2C%20Windows%2C%20macOS/blue">
</p>

Go-BirdNET is an application inspired by BirdNET Analyzer, developed by the K. Lisa Yang Center for Conservation Bioacoustics at the Cornell Lab of Ornithology and Chemnitz University of Technology. While the original BirdNET is based on Python, Go-BirdNET is built using Golang, aiming for simplified deployment across multiple platforms, from Windows PCs to single board computers like Raspberry Pi.

Currently, Go-BirdNET supports the analysis of WAV files only, support for other audio file formats is planned. Use of metadata model for location-based filtering is not yet implemented. This is very much of work in progress, and contributions and feedback are welcome.

## Dependencies

Go-BirdNET primarily relies on go-tflite by Yasuhiro Matsumoto (a.k.a. mattn). Go-tflite provides a Go binding for the TensorFlow Lite C API. Although go-tflite is documented to support only TensorFlow Lite version 2.2.0-rc3, I have successfully compiled it against version 2.14.0. As go-tflite is a C API binding compiled binary is not statically linked one, it is depending on libtensorflowlite_c.so to be present on deployment target system.

A crucial component is the BirdNET tflite model. After all, Go-BirdNET essentially serves as an interface to this model, with the core detection functionality residing within the BirdNET tflite model itself.

## Executable Distributions

Ready to run binaries can be found from releases section https://github.com/tphakala/go-birdnet/releases/
Arhives also contains libtensorflow_c library.

## Compiling

### Building TensorFlow Lite C API

Build tflite with cmake

```bash
sudo apt-get install cmake
```

Clone tensorflow repository

```bash
git clone https://github.com/tensorflow/tensorflow.git tensorflow_src
```

Create cmake build directory

```bash
mkdir tflite_build
cd tflite_build
```

Run cmake

```bash
cmake ../tensorflow_src/tensorflow/lite/c
```

Build tflite, In the tflite_build directory do

```bash
cmake --build . -j
```

Copy compiled libtensorflowlite_c.so to /usr/local/lib and update dynamic linker run‐time bindings with ldconfig

```bash
sudo cp libtensorflowlite_c.so /usr/local/lib
sudo ldconfig
```

### Building Go-BirdNET

Clone go-birdnet repository

```bash
git clone https://github.com/tphakala/go-birdnet.git
```

Add CGO_CFLAGS and point it to directory you cloned tensorflow source in

```bash
export CGO_CFLAGS=-I$HOME/src/tensorflow
```

Build Go-BirdNET by make, compiled binary will be placed in go-BirdNET/bin directory

```bash
cd go-birdnet
make
```

## Usage

```bash
$ ./bin/birdnet
Go-BirdNET CLI

Usage:
  birdnet [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  file        Analyze an audio file
  help        Help about any command
  realtime    Analyze audio in realtime mode

Flags:
      --debug               Enable debug output
  -h, --help                help for birdnet
      --locale string       Language to use for labels
      --model string        Path to the model file (default "model/BirdNET_GLOBAL_6K_V2.4_Model_FP32.tflite")
      --sensitivity float   Sigmoid sensitivity value between 0.0 and 1.5 (default 1)

Use "birdnet [command] --help" for more information about a command.
```

## ChangeLog

Realtime detection support is now added, launch with -realtime flag and application will start capturing audio from default audio device and log detections to detections.log.

## Derived works

Feel free to fork this codebase for your own projects, but please follow license which restricts commercial usage

## License

Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International

## Authors

Tomi P. Hakala

BirdNET model by the K. Lisa Yang Center for Conservation Bioacoustics at the Cornell Lab of Ornithology in collaboration with Chemnitz University of Technology. Stefan Kahl, Connor Wood, Maximilian Eibl, Holger Klinck.

BirdNET label translations by Patrick Levin for BirdNET-Pi project by Patrick McGuire.
