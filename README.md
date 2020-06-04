# Tunnel
 * [Cheatsheet](https://github.com/fedir/go-tooling-cheat-sheet/blob/master/go-tooling-cheat-sheet.pdf)

### Performance
 * [Gotta go fast](https://bravenewgeek.com/so-you-wanna-go-fast/)
 * [Go Networking Patterns](https://youtu.be/afSiVelXDTQ)
 * [Ratelimit](https://github.com/fujiwara/shapeio)
 * [Optimizing Micro Code](https://youtu.be/keydVd-Zn80)

### Minecraft
 * [GO Minecraft Packets](https://github.com/LilyPad/GoLilyPad/tree/0b14610d633f0ffd0af922b0357a24508e2b6cbc/packet/minecraft)
 * [Go Types](https://github.com/LilyPad/GoLilyPad/blob/c4d5d63f848711514698ac36f737a2779efe402e/packet/types.go)
 * [Proxy Server](https://github.com/go-mc/UnitedServer/blob/master/proxy.go)
 * [Custom Status](https://github.com/LilyPad/GoLilyPad/blob/669d6fd610322a0f61fc18bdcf94acaad2d16c1a/server/proxy/session.go#L330-L385)
 
### Go private modules
`go mod edit -replace=github.com/alexedwards/argon2id=/home/alex/code/argon2id`

`go mod edit -dropreplace=github.com/alexedwards/argon2id`

### Benchmark
`go tool trace /tmp/trace.out`

`go tool pprof -http=:5000 /tmp/cpuprofile.out`

`go tool pprof --nodefraction=0.1 -http=:5000`