# Go Tomcat Manager

A command-line tool written in Go for managing Tomcat server deployments, updating JSP files, and handling Tomcat configurations.

## Features

- Update JSP files in a running Tomcat server with a fixed-size goroutine pool for efficient copying.
- Manage Tomcat server environment variables and configuration files.
- Copy WAR files to Tomcat deployment directories.

## Installation

1. Clone the repository:
   
git clone https://github.com/nanaki-93/go-tomcat.git
cd go-tomcat
   

2. Build the project:
   
go build -o go-tomcat
   
## Pre requisites
You need a tomcat server in the resources directory when you run the `init` command. 
The project will create a `tomcat` directory in the resources folder with the necessary structure.


## Usage
- Initialize the project (creates the necessary directories and files):
- ./go-tomcat init

- Update JSP files in a running Tomcat server:
  
./go-tomcat update <appName>
  

- Start, stop, and manage Tomcat servers (see available commands):
  
./go-tomcat --help
  

## Configuration

- Edit your application and Tomcat configuration files as needed.
- Place your resources in the appropriate directories (see project structure).

## Development

- Requires Go 1.20+.
- Dependencies are managed via `go.mod`.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

MIT License. See `LICENSE` file for details.