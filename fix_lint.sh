sed -i 's/pprof.StartCPUProfile(f)/_ = pprof.StartCPUProfile(f)/g' benchmark/main.go
sed -i 's/defer f.Close()/defer func() { _ = f.Close() }()/g' benchmark/main.go
sed -i 's/fmt.Fprintln(f, "Config/_, _ = fmt.Fprintln(f, "Config/g' benchmark/main.go
sed -i 's/fmt.Fprintf(f, "%s/_, _ = fmt.Fprintf(f, "%s/g' benchmark/main.go
sed -i 's/!(mbX+1 >= mbW && trackers\[mbY-1\].col >= mbW-1)/mbX+1 < mbW || trackers[mbY-1].col < mbW-1/g' encoder.go
sed -i 's/rng := rand.New(rand.NewSource(config.RandSeed))/rng := rand.New(rand.NewSource(config.RandSeed)) \/\/nolint:gosec/g' benchmark/main.go
