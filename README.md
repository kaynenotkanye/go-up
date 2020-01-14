# go-up! A simple configuration library with placeholders resolution and no magic.

[![Build Status](https://travis-ci.org/ufoscout/go-up.svg?branch=master)](https://travis-ci.org/ufoscout/go-up)
[![codecov](https://codecov.io/gh/ufoscout/go-up/branch/master/graph/badge.svg)](https://codecov.io/gh/ufoscout/go-up)
[![Go Report Card](https://goreportcard.com/badge/github.com/ufoscout/go-up)](https://goreportcard.com/report/github.com/ufoscout/go-up)

go-up provides a simple way to configure an application from multiple sources — built in resources, property files, environment variables, and whatever else you like.

Some features:

- Recursive placeholders resolution
- Less than 10Kb 
- No external dependencies
- Modular and extensible architecture

This library was started after the considerable disappointment I had when I noticed that the configuration library used by my application increased the binary file size from 3MB to 11MB!!! (See the [Rationale](#rationale) section below for detailed info)

Go-Up is written to be lightweight and will always remain lean and straightforward.


Getting Started
---------------

1. To get started, import go-up:

In your go file:
```Go
import (
	"github.com/ufoscout/go-up"
)
```

If you use VGO, add in your go.mod file:
```
require github.com/ufoscout/go-up v0.6.1
```


2. Define some properties. You can use placeholders, for example, in `config.properties`:

```properties
# This line is a comment because it stats with '#'

server.port=9090
server.host=127.0.0.1

# Placeholders here, they will be resolved at runtime
server.url=http://${server.host}:${server.port}/
```

Keys and values in a properties file are automatically trimmed.


3. Build a go-up object that your properties:

```Go
ignoreFileNotFound := false 
up, err := NewGoUp().
		AddFile("./confing.properties", ignoreFileNotFound).
		Build();
```


4. Look up properties by key:

```Go
port := up.GetIntOrDefault("server.port", 8080); // returns 9090
serverUrl := up.GetString("server.url") // returns http://127.0.0.1:9090/
defaultVal := up.GetStringOrDefault("unknown_Key", "defaultValue") // returns defaultValue
```


Readers
-------
In go-up a "Reader" is whatever source of properties.
By default, go-up offers readers to load properties from:

* properties files in the file system
* Environment variables
* Programmatically typed properties

Custom properties readers can be easily created implementing the `github.com/ufoscout/go-up/reader/Reader` interface.


Placeholders resolution
-----------------------
go-up resolves placeholders recursively. 
For example:

fileOne.properties:
```properties
server.url=http://${${environment}.server.host}:${server.port}
server.port=8080
```

fileTwo.properties:
```properties
environment=PROD
PROD.server.host=10.10.10.10
DEV.server.host=127.0.0.1
```

Go code:
```Go
up, err := go_up.NewGoUp().
 AddFile("./fileOne.properties", false).
 AddFile("./fileTwo.properties", false).
 Build()

fmt.Println(up.GetString("server.url")) // this prints 'http://10.10.10.10:8080'
```

By default `${` and `}` delimiters are used. Custom delimiters can be easily defined:

```Go
up, err := go_up.NewGoUp().
 Delimiters("%(", ")"). // using %( and ) as delimiters
 ... etc ...
 Build()
```

Default Values
--------------
Placeholders can have default values which are used if the key is not otherwise provided. Example:

config.properties:
```properties
# the default value "8080" is used if 'PORT_NUMBER' is not provided
server.port=${PORT_NUMBER:8080}
    
# default is 127.0.0.1
server.ip=${IP:127.0.0.1}

server.url=${server.ip}/${server.port}
```

Go code:
```Go
up, err := go_up.NewGoUp().
 AddFile("./config.properties", false).
 Build()

// this prints 'http://127.0.0.1:8080'
fmt.Println(up.GetString("server.url")) 
```

The default separator for the default value is ":". A custom value can be set through the 'DefaultValueSeparator()' method of the GoUpBuilder.

Readers priority -> Last one wins
---------------------------------
Properties defined in later readers will override properties defined in earlier readers, in case of overlapping keys.
Hence, make sure that the most specific readers are the last ones in the given list of locations.

For example:

fileOne.properties:
```properties
server.url=urlFromOne
```

fileTwo.properties:
```properties
server.url=urlFromTwo
```

```Go
up, err := go_up.NewGoUp().
 AddFile("./fileOne.properties", false). 
 AddFile("./fileTwo.properties", false).
 AddReader(go_up.NewEnvReader("", false, false)). // Loading environment variables
 Build()

// this prints 'urlFromTwo'
fmt.Println(up.GetString("server.url"))

```

BTW, due to the fact that we used go_up.NewEnvReader(...)` as last reader, if at runtime an Environment variable called "server.url" is found, it will override the other values.

Finally, it is possible to specify a custom priority:

```Go
up, err := go_up.NewGoUp().

 // load the properties from the file system and specify their priority
 AddFileWithPriority("./fileOne.properties", false, go_up.HighestPriority).

 AddFile("./fileTwo.properties", false). // loads file with default priority
 AddReader(go_up.NewEnvReader("", false, false)). // Loads environment variables
 Build()

// this prints 'urlFromOne'
fmt.Println(up.GetString("server.url"))
```

The default priority is 100. The highest priority is 0.
As usual, when more readers have the same priority, the last one wins in case of overlapping keys.


Working with Environment Variables
----------------------------------
Built in support for environment variables is provided through the EnvReader struct:

```Go
// These should be defined in your system, not in the code. Here only as example.
os.Setenv("KEY_FROM_ENV", "ValueFromEnv")
os.Setenv("CUSTOM_PREFIX_KEY_FROM_ENV", "ValueFromEnv_WithCustomPrefix")

// Loads only variables with this prefix. The prefix is removed at runtime.
prefix := "CUSTOM_PREFIX_" 
// if true, the Env variable key is converted to lower case
toLowerCase := true 
// if true, the underscores "_" of the env variable key are replaced by dots "."
underscoreToDot := true 

up, err := go_up.NewGoUp().
 AddReader(go_up.NewEnvReader(prefix, toLowerCase, underscoreToDot)). // Loading environment variables
 Build()

// this prints 'ValueFromEnv_WithCustomPrefix'
fmt.Println(up.GetString("key.from.env"))
```


Real life example
-----------------
A typical real life configuration would look like:

```Go
up, err := go_up.NewGoUp().

 // load a file
 AddFile("./default.properties", false).

 // load another file
 AddFile("./config/config.properties", false).

 // Add a not mandatory file, resource not found errors are ignored.
 // Here I am adding properties requried only during testing.
 AddFile("./test/test.properties", true).

 // Load the Environment variables.
 // The are used as they are defined, e.g. ENV_VARIABLE=XXX
 AddReaderWithPriority(go_up.NewEnvReader("APP_PREFIX_", false, false)).

 // Load the Environment variables and convert their keys
 // from ENV_VARIABLE=XXX to env.variable=XXX
 // This could be desired to override default properties
 AddReaderWithPriority(go_up.NewEnvReader("APP_PREFIX_", true, true)).
 
 // build the go-up object
 .Build()
```

go-up API
-------------
go-up has a straightforward API that hopefully does not need detailed documentation.

Some examples:

```Go
up, err := go_up.NewGoUp().
 Add("programmatically.added.key", "12345"). // default values can be added programmatically
 AddFile("./myFile.properties", false).
 Build()

// get a String. The empty string value is returned if the key is not found
aString := up.GetString("key")

// get a String. "defaultValue" is returned if the key is not found
aStringOrDefault := up.GetStringOrDefault("key", "defaultValue")

// get a String. An error is returned if the key is not found
aStringOrError, error1 := up.GetStringOrFail("key")

// get a slice from the comma separated tokens of the property value
aslice := up.GetStringSlice("key", ",")
```

All available methods are defined in the GoUp interface:

```Go
type GoUp interface {

  Exists(key string) bool

  GetBool(key string) bool
  GetBoolOrDefault(key string, defaultValue bool) bool
  GetBoolOrFail(key string) (bool, error)
   
  GetFloat64(key string) float64
  GetFloat64OrDefault(key string, defaultValue float64) float64
  GetFloat64OrFail(key string) (float64, error)
    
  GetInt(key string) int
  GetIntOrDefault(key string, defaultValue int) int
  GetIntOrFail(key string) (int, error)
   
  GetString(key string) string
  GetStringOrDefault(key string, defaultValue string) string
  GetStringOrFail(key string) (string, error)

  GetStringSlice(key string, separator string) []string
  GetStringSliceOrDefault(key string, separator string, defaultValue []string) []string
  GetStringSliceOrFail(key string, separator string) ([]string, error)

}
```

GoDoc
-----
See [https://godoc.org/github.com/ufoscout/go-up](https://godoc.org/github.com/ufoscout/go-up)

# Rationale
Before Go-Up, I used to use [Viper](https://github.com/spf13/viper). 
Viper is great and I surely recommend it; nevertheless, all of its features are not for free. In fact, it adds tons of dependencies that make the application much heavier.

So, if like me you aspire to code which is as light as possible and you can live with a smaller set of features (BWT go-up has some unique aces in the hole like recursive placeholders resolution!), then you should probably take into account a lighter alternative like go-up.

To show the impact that a single library can have on your application, I created three small examples (see: [go-up_examples](https://github.com/ufoscout/go-up_examples) ), these are:

- *goup* : a Go application that, using **go-up**, reads "Hello World" from a config file and prints it
- *plain* : a Go application that prints "Hello World"
- *viper* : a Go application that, using **viper**, reads "Hello World" from a config file and prints it

When built, they produce surprising results:

| Library       | Produced Binary Size |
| ------------- |---------------------:|
| Go-Up         |             2.185 KB |
| Plain Go      |             2.000 KB |
| Viper         |            11.782 KB |

<ins>The Viper based binary file is nearly 6 times bigger than the other implementations!</ins>

(Build performed with go1.13 linux/amd64 on Ubuntu 18.04)
