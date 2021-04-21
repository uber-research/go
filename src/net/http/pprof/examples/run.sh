set -ex 
#GOROOT=<set me>
#PATH=<set me>
#if needed and supported
#GO_PPROF_PRECISION=3

HOST=localhost
PORT=6060
L1MISS="r08d1"

for f in fusion.go  goroutine.go  serial.go  transpose.go
do
name=`basename $f .go`
go run $f &
pid=$!
sleep 2
# simple timer profile
curl -o ${name}.timer.prof ${HOST}:${PORT}/debug/pprof/profile?seconds=10
# CPU cycles @ 10M
curl -o ${name}.cycles.prof ${HOST}:${PORT}/debug/pprof/profile?event=cycles\&period=10000000\&seconds=20
# CPU instructions @ 1M
curl -o ${name}.ins.prof ${HOST}:${PORT}/debug/pprof/profile?event=instructions\&period=1000000\&seconds=20
# CPU L1 D-cache miss @ 10000 
curl -o ${name}.l1miss.prof ${HOST}:${PORT}/debug/pprof/profile?event=${L1MISS}\&period=10000\&seconds=20
# CPU LLC cache miss @ 10000 
curl -o ${name}.llcmiss.prof ${HOST}:${PORT}/debug/pprof/profile?event=cacheMisses\&period=10000\&seconds=20
killall $name || "cannot kill"
done
