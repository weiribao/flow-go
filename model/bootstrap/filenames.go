package bootstrap

import "path/filepath"

// Canonical filenames/paths for bootstrapping files.
var (
	// The Node ID file is used as a helper by the transit scripts
	FilenameNodeID = "node-id"
	PathNodeID     = filepath.Join(DirnamePublicBootstrap, FilenameNodeID) // %v will be replaced by NodeID

	// execution state
	DirnameExecutionState = "execution-state"

	// public genesis information
	DirnamePublicBootstrap    = "public-root-information"
	PathNodeInfosPub          = filepath.Join(DirnamePublicBootstrap, "node-infos.pub.json")
	PathPartnerNodeInfoPrefix = filepath.Join(DirnamePublicBootstrap, "node-info.pub.")
	PathNodeInfoPub           = filepath.Join(DirnamePublicBootstrap, "node-info.pub.%v.json") // %v will be replaced by NodeID
	PathRootBlock             = filepath.Join(DirnamePublicBootstrap, "root-block.json")
	PathRootQC                = filepath.Join(DirnamePublicBootstrap, "root-qc.json")
	PathRootResult            = filepath.Join(DirnamePublicBootstrap, "root-execution-result.json")
	PathRootSeal              = filepath.Join(DirnamePublicBootstrap, "root-block-seal.json")

	// private genesis information
	DirPrivateRoot           = "private-root-information"
	FilenameRandomBeaconPriv = "random-beacon.priv.json"
	PathNodeInfoPriv         = filepath.Join(DirPrivateRoot, "private-node-info_%v", "node-info.priv.json")    // %v will be replaced by NodeID
	PathRandomBeaconPriv     = filepath.Join(DirPrivateRoot, "private-node-info_%v", FilenameRandomBeaconPriv) // %v will be replaced by NodeID
)
