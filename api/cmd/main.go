package main

import (
	"os"
	"runtime"

	"github.com/carlmjohnson/versioninfo"
	"github.com/pyroscope-io/client/pyroscope"
	"github.com/runar-rkmedia/gotally/api"
)

func main() {

	{
		pyroScopeUrl := os.Getenv("PYROSCOPE_URL")
		if pyroScopeUrl != "" {
			err := startPeriscope(pyroScopeUrl, os.Getenv("PYROSCOPE_AUTH_TOKEN"), true, true)
			if err != nil {
				panic(err)
			}
		}
	}
	api.StartServer()
}

// testing if periscope is useful
func startPeriscope(url string, authToken string, withBlock, withMutex bool) error {
	// // These 2 lines are only required if you're using mutex or block profiling
	// // Read the explanation below for how to set these rates:
	// runtime.SetMutexProfileFraction(5)
	// runtime.SetBlockProfileRate(5)

	conf := pyroscope.Config{

		ApplicationName: "gotally.rkmedia.game",

		// replace this with the address of pyroscope server
		ServerAddress: url,

		// you can disable logging by setting this to nil
		Logger: nil, // pyroscope.StandardLogger,
		Tags: map[string]string{
			"version": versioninfo.Version,
			"hash":    versioninfo.Short(),
		},

		// optionally, if authentication is enabled, specify the API key:
		// AuthToken: os.Getenv("PYROSCOPE_AUTH_TOKEN"),

		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	}

	if authToken != "" {
		conf.AuthToken = authToken
	}

	if withMutex {
		runtime.SetMutexProfileFraction(5)
		conf.ProfileTypes = append(conf.ProfileTypes, pyroscope.ProfileMutexCount, pyroscope.ProfileMutexDuration)
	}
	if withBlock {
		runtime.SetMutexProfileFraction(5)
		conf.ProfileTypes = append(conf.ProfileTypes, pyroscope.ProfileBlockCount, pyroscope.ProfileBlockDuration)
	}

	_, err := pyroscope.Start(conf)
	return err

}
