package tgframe

import "runtime/debug"

// modulePath is the module path go.mod declares, and the key the build info
// lists this module under in a dependent's binary.
const modulePath = "github.com/voilelab/toolgui"

// fallbackVersion is what Version reports when the build info carries no entry
// for the module -- a binary built from inside this repo, where toolgui is the
// main module rather than a dependency.
//
// The Release workflow rewrites this line on the commit it tags, so a released
// tree always carries its own version. The value committed on dev is the last
// release it was bumped to.
const fallbackVersion = "v0.3.0"

// Version return the version of toolgui this binary was built against.
//
// It reads the module version out of the build info, so a released tag is
// reported without anything to maintain, and falls back to the version
// recorded at release time when the build info has no entry for the module.
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fallbackVersion
	}

	if info.Main.Path == modulePath && isRealVersion(info.Main.Version) {
		return info.Main.Version
	}

	for _, dep := range info.Deps {
		if dep.Path != modulePath {
			continue
		}

		// A replace directive points the module somewhere else; the
		// replacement carries the version that was actually built.
		if dep.Replace != nil {
			dep = dep.Replace
		}

		if isRealVersion(dep.Version) {
			return dep.Version
		}
	}

	return fallbackVersion
}

// isRealVersion reject the placeholders the toolchain uses when it has no
// version to report: "" and "(devel)".
func isRealVersion(v string) bool {
	return v != "" && v != "(devel)"
}
