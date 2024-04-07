# very simple https server

## OS

Windows 11

## Installation

1. Download Golang releases

     https://go.dev/ -> [Download](https://go.dev/dl/) -> zip (e.g. go1.22.2.windows-amd64.zip)

1. Extract that zip to the current folder

    ```txt
    ./
        + go/      # Files from the extracted archive
            + bin/ #
            + ...  #
        + ...
        + .gitignore
        + cmd.bat
        + README.md
        + ...
    ```

## Build

    ```sh
    # Open Command prompt or Powershell
    
    .\cmd.bat

    go build http.go --output run.exe
    ```

## Run

    ```sh
    # Run https server, listen 127.0.0.1:3000 and open browser
    .\run.exe

    # Run http server, listen 0.0.0.0:8080 and not open browser
    .\run.exe -no-tls -no-browse -p 8080 -b 0.0.0.0
    ```
