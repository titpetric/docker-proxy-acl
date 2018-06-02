package main

import "github.com/namsral/flag"
import "fmt"

import "os"
import "os/signal"
import "syscall"

import "net"
import "net/http"
import "github.com/gorilla/mux"

import "./proxy"

type stringSlice []string;

func (s *stringSlice) String() string {
	return fmt.Sprintf("%d", *s)
}
func (s *stringSlice) Set(value string) error {
	fmt.Sprintf("Allowing endpont: %s\n", value);
	*s = append(*s, value);
	return nil;
}

func main() {
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "GO", 0)
	var (
		allowed stringSlice;
		allowedMap map[string]bool = make(map[string]bool);
		filename = fs.String("filename", "/tmp/docker-proxy-acl/docker.sock", "Location of socket file");
	)
	fs.Var(&allowed, "a", "Allowed location pattern prefix");
	fs.Parse(os.Args[1:])

	if len(allowed) < 1 {
		fmt.Println("Need at least 1 argument for -a: [containers, networks, version, info, ping]");
		os.Exit(0);
	}

	for _, s := range allowed {
		allowedMap[s] = true;
	}

	var routers [2]*mux.Router;
	routers[0] = mux.NewRouter();
	routers[1] = routers[0].PathPrefix("/{version:[v][0-9]+[.][0-9]+}").Subrouter();

	upstream := proxy.NewSocket("/var/run/docker.sock");

	if allowedMap["containers"] {
		fmt.Printf("Registering container handlers\n");
		for _, m := range routers {
			containers := m.PathPrefix("/containers").Subrouter();
			containers.HandleFunc("/json", upstream.Pass());
			containers.HandleFunc("/{name}/json", upstream.Pass());
		}
	}

	if allowedMap["images"] {
		fmt.Printf("Registering images handlers\n");
		for _, m := range routers {
			containers := m.PathPrefix("/images").Subrouter();
			containers.HandleFunc("/json", upstream.Pass());
			containers.HandleFunc("/{name}/json", upstream.Pass());
			containers.HandleFunc("/{name}/history", upstream.Pass());
		}
	}

	if allowedMap["volumes"] {
		fmt.Printf("Registering volumes handlers\n");
		for _, m := range routers {
			m.HandleFunc("/volumes", upstream.Pass());
			m.HandleFunc("/volumes/{name}", upstream.Pass());
		}
	}

	if allowedMap["networks"] {
		fmt.Printf("Registering networks handlers\n");
		for _, m := range routers {
			m.HandleFunc("/networks", upstream.Pass());
			m.HandleFunc("/networks/{name}", upstream.Pass());
		}
	}

	if allowedMap["services"] {
		fmt.Printf("Registering services handlers\n");
		for _, m := range routers {
			m.HandleFunc("/services", upstream.Pass());
			m.HandleFunc("/services/{name}", upstream.Pass());
		}
	}

	if allowedMap["tasks"] {
		fmt.Printf("Registering tasks handlers\n");
		for _, m := range routers {
			m.HandleFunc("/tasks", upstream.Pass());
			m.HandleFunc("/tasks/{name}", upstream.Pass());
		}
	}

	if allowedMap["events"] {
		fmt.Printf("Registering events handlers\n");
		for _, m := range routers {
			m.HandleFunc("/events", upstream.Pass());
		}
	}

	if allowedMap["version"] {
		fmt.Printf("Registering version handlers\n");
		for _, m := range routers {
			m.HandleFunc("/version", upstream.Pass());
		}
	}

	if allowedMap["info"] {
		fmt.Printf("Registering info handlers\n");
		for _, m := range routers {
			m.HandleFunc("/info", upstream.Pass());
		}
	}

	if allowedMap["ping"] {
		fmt.Printf("Registering ping handlers\n");
		for _, m := range routers {
			m.HandleFunc("/_ping", upstream.Pass());
		}
	}

	http.Handle("/", routers[0]);

	l, err := net.Listen("unix", *filename)
	os.Chmod(*filename, 0666);
	// Looking up group ids coming up for Go 1.7
	// https://github.com/golang/go/issues/2617

	fmt.Println("[docker-proxy-acl] Listening on " + *filename);

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		fmt.Printf("[docker-proxy-acl] Caught signal %s: shutting down.\n", sig)
		l.Close()
		os.Exit(0)
	}(sigc)

	if err != nil {
		panic(err)
	} else {
		err := http.Serve(l, nil)
		if err != nil {
			panic(err)
		}
	}
}