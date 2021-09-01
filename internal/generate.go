/*
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 *
 */

package codegen

//go:generate jutec --outDir=. -go.moduleMap=org.apache.zookeeper.server:- -go.moduleMap=org.apache.zookeeper.txn:- -go.moduleMap=org.apache.zookeeper: -go.prefix=github.com/facebookincubator/zk/internal zookeeper.jute
