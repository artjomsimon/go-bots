OUT=bin/
WTIME=wtime.go

TP1=simple-queue
TP2=const-goroutines
TP3=const-goroutines-taskwait
TP4=goroutine-dispatch
TP5=goroutine-dispatch-taskwait
TP6=notaskpool
TP7=seq

all: nqueens sparselu

nqueens:
	go build -o $(OUT)$@-$(TP3) $@-taskpool.go taskpool-$(TP3).go $(WTIME)
	go build -o $(OUT)$@-$(TP5) $@-taskpool.go taskpool-$(TP5).go $(WTIME)
	go build -o $(OUT)$@-$(TP6) $@-$(TP6).go $(WTIME)
	go build -o $(OUT)$@-$(TP7) $@-$(TP7).go $(WTIME)
	gccgo -O3 -o $(OUT)$@-gccgo-$(TP3) $@-taskpool.go taskpool-$(TP3).go $(WTIME)
	gccgo -O3 -o $(OUT)$@-gccgo-$(TP5) $@-taskpool.go taskpool-$(TP5).go $(WTIME)
	gccgo -O3 -o $(OUT)$@-gccgo-$(TP6) $@-$(TP6).go $(WTIME)
	gccgo -O3 -o $(OUT)$@-gccgo-$(TP7) $@-$(TP7).go $(WTIME)

sparselu:
	go build -o $(OUT)$@-$(TP1) $@-goforloop.go taskpool-$(TP1)-chan.go $(WTIME)
	go build -o $(OUT)$@-$(TP3) $@-taskpool.go taskpool-$(TP3).go $(WTIME)
	go build -o $(OUT)$@-$(TP5) $@-taskpool.go taskpool-$(TP5).go $(WTIME)
	go build -o $(OUT)$@-$(TP6) $@-$(TP6).go $(WTIME)
	go build -o $(OUT)$@-$(TP7) $@-$(TP7).go $(WTIME)
	gccgo -O3 -o $(OUT)$@-gccgo-$(TP3) $@-taskpool.go taskpool-$(TP3).go $(WTIME)
	gccgo -O3 -o $(OUT)$@-gccgo-$(TP5) $@-taskpool.go taskpool-$(TP5).go $(WTIME)
	gccgo -O3 -o $(OUT)$@-gccgo-$(TP6) $@-$(TP6).go $(WTIME)
	gccgo -O3 -o $(OUT)$@-gccgo-$(TP7) $@-$(TP7).go $(WTIME)

clean:
	rm bin/*
