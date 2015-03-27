 > cat /proc/cpuinfo | grep model
 model name      : AMD Opteron(tm) Processor 6172
 
 > free -m
             total       used       free     shared    buffers     cached
Mem:        128562       5905     122656         45        199       4266
-/+ buffers/cache:       1439     127122
Swap:         8197          0       8197

> go version
go version go1.4.2 linux/amd64

> export GOTRACEBACK=2
> export GODEBUG=scheddetail=1
> for i in {1..20}; do go run sparselu-crash-mwe.go -n 201 -m 69 &> bugs/invalid-p-state/invalid-p-state-$i.txt; done