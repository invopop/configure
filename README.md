# configure

Help configure go projects using both a configuration template yaml file and environment variables.

## Usage

Import the package and load a YAML file with template data. If your current directory contains a `.env` file, this will be loaded automatically.

```go
package main

import (
    "log"
    "os"

    "github.com/invopop/configure"
)

const configFile = "samples/config.yaml.tmpl"

// Config definition for our project
type Config struct {
    S3Bucket string `json:"s3_bucket"`
}

func main() {
    conf := new(Config)
    if err := configure.Load(configFile, conf); err != nil {
        log.Fatal("Error loading configuration file")
    }

    fmt.Printf("S3 Bucket is: %v\n", conf.S3Bucket)
}
```
