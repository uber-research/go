set -ex 
#GOROOT=<set me>
#PATH=<set me>
#if needed and supported
#GO_PPROF_PRECISION=3
export GO_PPROF_ENABLE_MULTIPLE_CPU_PROFILES=true

for f in fusion.go  goroutine.go  serial.go  transpose.go
do
name=`basename $f .go`
go build $f 
./${name}
done
