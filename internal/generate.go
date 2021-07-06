package internal

//go:generate jutec --outDir=. -go.moduleMap=org.apache.zookeeper.server:- -go.moduleMap=org.apache.zookeeper.txn:- -go.moduleMap=org.apache.zookeeper: -go.prefix=github.com/facebookincubator/zk/internal zookeeper.jute
