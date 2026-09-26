sed -i 's/result.AllocsPerFrame = int64/result.AllocsPerFrame = int64 \/* #nosec G115 *\//g' benchmark/main.go
sed -i 's/result.AllocBytesPerFrame = int64/result.AllocBytesPerFrame = int64 \/* #nosec G115 *\//g' benchmark/main.go
sed -i 's/r := uint8(rng.Int31n(256))/r := uint8(rng.Int31n(256)) \/* #nosec G115 *\//g' benchmark/main.go
sed -i 's/g := uint8(rng.Int31n(256))/g := uint8(rng.Int31n(256)) \/* #nosec G115 *\//g' benchmark/main.go
sed -i 's/b := uint8(rng.Int31n(256))/b := uint8(rng.Int31n(256)) \/* #nosec G115 *\//g' benchmark/main.go
sed -i 's/x := (int(rng.Int31n(int32(width))) + offset) % width/x := (int(rng.Int31n(int32(width))) + offset) % width \/* #nosec G115 *\//g' benchmark/main.go
sed -i 's/y := int(rng.Int31n(int32(height)))/y := int(rng.Int31n(int32(height))) \/* #nosec G115 *\//g' benchmark/main.go
sed -i 's/rng := rand.New/rng := rand.New \/* #nosec G404 *\//g' benchmark/main.go
