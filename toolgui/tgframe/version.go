package tgframe

import "runtime/debug"

// modulePath is the module path go.mod declares. A dependent's build info
// lists this module under it.
const modulePath = "github.com/voilelab/toolgui"

// fallbackVersion is what Version reports when the build info names no commit
// at all: `go run`, a source tree with no VCS metadata, or `-buildvcs=false`,
// which the wails CLI passes unconditionally.
//
// The Release workflow rewrites this line on the commit it tags, so a released
// tree reports its own version. What dev carries is deliberately not a release
// number: a build off dev is not one, and naming a release it is not reads as
// fact rather than as the guess it is.
const fallbackVersion = "v0.6.0"

// Version return the version of toolgui this binary was built against.
//
// It reads the module version out of the build info, so a released tag is
// reported without anything to maintain. A build off a tag reports the
// pseudo-version the toolchain derives from the commit. Only a build with no
// version at all falls back to the version recorded at release time.
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

// isRealVersion reject what the toolchain reports when it has no commit to
// name: "", "(devel)", and a bare "v0.0.0", which is the placeholder a go.mod
// carries for a module it reaches through a replace.
//
// A v0.0.0-<date>-<revision> pseudo-version is kept. The toolchain stamps it
// off a checkout with no tag to derive from, so it names no release -- but it
// does name the commit, which is the most a build off a branch can say.
func isRealVersion(v string) bool {
	return v != "" && v != "(devel)" && v != "v0.0.0"
}
